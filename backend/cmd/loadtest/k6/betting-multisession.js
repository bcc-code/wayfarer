import { SharedArray } from 'k6/data';
import { sleep, check } from 'k6';
import exec from 'k6/execution';
import { Counter, Trend } from 'k6/metrics';

import { parseResponse } from './lib/graphql.js';
import { coldLoad } from './queries/bootstrap.js';
import { maybeAuthDance } from './lib/realism.js';
import { challengePage, challengePageRefetch } from './queries/challenge.js';
import { finalizeQuiz } from './queries/quiz.js';
import { startBettingSession, submitBet } from './queries/betting.js';
import { startBetting, lockBetting, finishBetting } from './queries/gamenight.js';

// Multi-session gamenight (notes/multi-session-loadtest-plan.md §4): ~1,000
// church admins each run their own betting session on the same quiz, around
// a coordinated 19:00 start (T0) and a shared finish time.
//
//   admins   one VU per church (bench/betting_box.sh reset-multi). At its
//            click time (70% normal around T0, sigma 30s; 30% uniform over
//            T0 +/- 5 min) plus the page's 10s countdown it creates, opens,
//            grants (church) and bulk-enrolls, like the gamenight page. Around
//            T0 + SESSION_DURATION it locks, and LOCK_TO_FINISH later
//            finishes, which settles the session's bets.
//   players  one VU per player, connected before T0 on the challenge page.
//            Once their church's session is visible they play the five
//            betting questions and finalize.
//
// Firestore refetches: every session transition pings one project-wide doc,
// and every connected client refetches ChallengePage. Today's frontend aborts
// the in-flight refetch on the next ping (frontend/test/unit/
// firestoreRefetchCoalescing.test.ts). The admin schedule is deterministic, so
// the init code builds the ping timeline, and each refetching player sends one
// ChallengePage per ping with a timeout up to the next ping: k6 drops it then,
// like the browser's abort. Each client receives the pings a fixed random
// delay late (PING_JITTER_MS), so 10k clients don't fire in the same ms. Players keep refetching while
// waiting, between questions and after finalizing, until the last finish.
//
// Needs config.json from tokengen -church-prefix CH01LOADTESTCHURCH (churches
// with admin and player tokens). With SHARDS > 1 every shard needs the same
// START_EPOCH_MS (loadtest-remote.sh sets it with LOADTEST_START_DELAY).

// CHURCH_SAMPLE=k runs only k evenly spaced churches of the fixture (sizes
// ascend with the index, so the sample spans small to largest), e.g. a smoke
// run on the full fixture and config.json
function sampledChurches() {
    const all = JSON.parse(open('../config.json')).churches || [];
    const k = parseInt(__ENV.CHURCH_SAMPLE || '0');
    if (!k || k >= all.length) {
        return all;
    }
    const picked = [];
    for (let j = 0; j < k; j++) {
        picked.push(all[k === 1 ? all.length - 1 : Math.round((j * (all.length - 1)) / (k - 1))]);
    }
    return picked;
}

const config = new SharedArray('config', function () {
    const parsed = JSON.parse(open('../config.json'));
    return [{ baseUrl: parsed.baseUrl, churchCount: sampledChurches().length }];
})[0];
const churches = new SharedArray('churches', function () {
    return sampledChurches().map((c) => ({ churchId: c.churchId, adminToken: c.admin.token, players: c.players.length }));
});
const players = new SharedArray('players', function () {
    const out = [];
    sampledChurches().forEach((c, churchIdx) => {
        for (const p of c.players) {
            out.push({ churchIdx, userId: p.userId, token: p.token, auth0Token: p.auth0Token || null });
        }
    });
    return out;
});

const baseUrl = __ENV.BASE_URL || config.baseUrl;
const challengeId = __ENV.CHALLENGE_ID || 'CL01LOADTESTBETTING000000000';
const quizId = __ENV.QUIZ_ID || 'QZ01LOADTESTBETTING000000000';

function num(name, fallback) {
    const v = __ENV[name];
    return v === undefined || v === '' ? fallback : parseFloat(v);
}
const seed = num('SEED', 19);
const coreFraction = num('CORE_FRACTION', 0.7);
const coreSigma = num('CORE_SIGMA', 30); // seconds
const stragglerWindow = num('STRAGGLER_WINDOW', 300); // seconds either side of T0
const sessionDuration = num('SESSION_DURATION', 2700); // T0 -> shared finish time
const lockToFinish = num('LOCK_TO_FINISH', 30);
const countdown = num('COUNTDOWN', 10); // the page's countdown before startBetting
const preconnectWindow = num('PRECONNECT_WINDOW', 300); // players connect over this, before the first click
const finishMargin = num('FINISH_MARGIN', 60); // players keep listening this long after the last finish
const sessionWaitMax = num('SESSION_WAIT_MAX', 600); // give up waiting for the session after this
const grantRetry = num('GRANT_RETRY', 2); // seconds between retries once the grant ping passed
const pluginPings = num('PLUGIN_PINGS', 1); // project pings the plugin sends after a finish webhook
const refetchMax = num('REFETCH_MAX_TIMEOUT', 60); // timeout of a refetch with no next ping
const pingJitterMs = num('PING_JITTER_MS', 250); // per-client ping delivery delay, uniform 0..this
const refetchFraction = num('REFETCH_FRACTION', 1);
const refetchOn = (__ENV.REFETCH || '1') !== '0';
const adminDoubleClickFraction = num('ADMIN_DOUBLE_CLICK_FRACTION', 0);
const correctRate = num('CORRECT_RATE', 0.5);
const noBetFraction = num('NO_BET_FRACTION', 0.05);
const thinkScale = num('THINK_SCALE', 1);
const finishTimeout = __ENV.FINISH_TIMEOUT || '30m';
const gamenight = __ENV.GAMENIGHT || '1';
// PING_DELIVERY=all (every ping reaches every client) or coalesce:<ms> (at
// most one snapshot per window, as the Firestore SDK may deliver)
const coalesceMs = (() => {
    const m = String(__ENV.PING_DELIVERY || 'all').match(/^coalesce:(\d+)$/);
    return m ? parseInt(m[1]) : 0;
})();

const shards = Math.max(1, parseInt(__ENV.SHARDS || '1'));
const shard = parseInt(__ENV.SHARD || '0');

// Seconds from the start of the run to T0: players connect first, then the
// earliest stragglers click
const lead = preconnectWindow + stragglerWindow + 30;

// mulberry32: a small seeded PRNG, so every VU and shard derives the same
// schedule
function rng(s) {
    let a = s >>> 0;
    return function () {
        a = (a + 0x6d2b79f5) >>> 0;
        let t = a;
        t = Math.imul(t ^ (t >>> 15), t | 1);
        t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
        return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
    };
}

// clickOffset draws when an admin clicks, in seconds relative to the
// coordinated time: CORE_FRACTION near it (normal, CORE_SIGMA), the rest
// spread uniformly over +/- STRAGGLER_WINDOW
function clickOffset(r) {
    if (r() < coreFraction) {
        const z = Math.sqrt(-2 * Math.log(r() || 1e-12)) * Math.cos(2 * Math.PI * r());
        return Math.max(-stragglerWindow, Math.min(stragglerWindow, z * coreSigma));
    }
    return (r() * 2 - 1) * stragglerWindow;
}

// Per church, in ms relative to T0: when the session opens (click +
// countdown), locks and finishes
const schedule = new SharedArray('schedule', function () {
    const r = rng(seed);
    const out = [];
    for (let i = 0; i < config.churchCount; i++) {
        const openMs = Math.round((clickOffset(r) + countdown) * 1000);
        const lockMs = Math.round((sessionDuration + clickOffset(r)) * 1000);
        out.push({
            openMs,
            lockMs,
            finishMs: lockMs + lockToFinish * 1000,
            doubleClick: r() < adminDoubleClickFraction,
        });
    }
    return out;
});

// Every project-wide quiz_sessions ping, in ms relative to T0, ascending:
// open and grant just after the session opens, lock, finish, and the plugin's
// pings once the finish webhook arrived
const pings = new SharedArray('pings', function () {
    const all = [];
    for (const s of schedule) {
        all.push(s.openMs + 200, s.openMs + 500, s.lockMs + 100, s.finishMs + 500);
        for (let i = 0; i < pluginPings; i++) {
            all.push(s.finishMs + 1000 + i * 100);
        }
    }
    all.sort((a, b) => a - b);
    if (!coalesceMs) {
        return all;
    }
    // At most one snapshot per window: a ping inside the window of the last
    // delivered snapshot is delivered once, at the end of that window
    const out = [];
    let last = -Infinity;
    for (const p of all) {
        if (p >= last + coalesceMs) {
            last = p;
            out.push(last);
        } else if (p > last) {
            last += coalesceMs;
            out.push(last);
        }
    }
    return out;
});

const lastFinishMs = schedule.reduce((max, s) => Math.max(max, s.finishMs), 0);
const runSeconds = lead + lastFinishMs / 1000 + finishMargin + 900;

const sessionsCreated = new Counter('sessions_created');
const sessionsFinished = new Counter('sessions_finished');
const adminFailures = new Counter('admin_failures');
const finishDuration = new Trend('finish_duration', true);
const timeToSessionVisible = new Trend('time_to_session_visible', true);
const sessionWaitTimeouts = new Counter('session_wait_timeouts');
const refetches = new Counter('refetches');
const refetchCancelled = new Counter('refetch_cancelled');
const refetchErrors = new Counter('refetch_errors');
const refetchCompletedLatency = new Trend('refetch_completed_latency', true);
// How late a refetch leaves after its ping: near 0 while the generator keeps
// up; when it grows, the run is measuring k6, not the server
const refetchSendLag = new Trend('refetch_send_lag', true);
const quizCompletions = new Counter('quiz_completions');
const quizFailures = new Counter('quiz_failures');
const betsAccepted = new Counter('bets_accepted');
const betsNone = new Counter('bets_none');

function perShard(n) {
    return Math.max(1, Math.ceil(n / shards));
}

export const options = {
    scenarios: {
        admins: {
            executor: 'per-vu-iterations',
            vus: perShard(churches.length),
            iterations: 1,
            maxDuration: `${Math.ceil(runSeconds)}s`,
            exec: 'admin',
            tags: { kind: 'admin' },
        },
        players: {
            executor: 'per-vu-iterations',
            vus: perShard(players.length),
            iterations: 1,
            maxDuration: `${Math.ceil(runSeconds)}s`,
            exec: 'player',
            tags: { kind: 'play' },
        },
    },
    setupTimeout: '2m',
    thresholds: {
        'http_req_failed{kind:play}': ['rate<0.01'],
        'http_req_failed{kind:admin}': ['rate<0.01'],
        graphql_errors: ['rate<0.01'],
        admin_failures: ['count<1'],
        session_wait_timeouts: ['count<1'],
        quiz_failures: [`count<${Math.max(1, Math.round(players.length * 0.005))}`],
        refetch_errors: ['count<1'],
        time_to_session_visible: ['p(95)<15000'],
        'http_req_duration{name:ChallengePage}': ['p(95)<1000'],
        'http_req_duration{name:StartQuizSession}': ['p(95)<1000'],
        'http_req_duration{name:SubmitQuizAnswerBet}': ['p(95)<1000'],
        'http_req_duration{name:FinalizeQuiz}': ['p(95)<1000'],
        'http_req_duration{name:OpenQuizSession}': ['p(95)<2000'],
        'http_req_duration{name:GrantQuizSessionAccessAsync}': ['p(95)<2000'],
        'http_req_duration{name:LockQuizSession}': ['p(95)<2000'],
        'http_req_duration{name:FinishQuizSession}': ['p(95)<10000'],
    },
};

function sleepUntil(epochMs) {
    const seconds = (epochMs - Date.now()) / 1000;
    if (seconds > 0) {
        sleep(seconds);
    }
}

// firstPingAfter returns the index of the first ping later than tMs (relative
// to T0), or pings.length
function firstPingAfter(tMs) {
    let lo = 0;
    let hi = pings.length;
    while (lo < hi) {
        const mid = (lo + hi) >> 1;
        if (pings[mid] <= tMs) {
            lo = mid + 1;
        } else {
            hi = mid;
        }
    }
    return lo;
}

/**
 * The client's Firestore listener until untilEpochMs: one ChallengePage
 * refetch per ping, cancelled by the next ping. Returns the first completed
 * page for which done(data) is true (then stops early), else null.
 */
let loggedRefetchErrors = 0; // per VU

function listen(t0, token, untilEpochMs, done) {
    let i = firstPingAfter(Date.now() - t0);
    for (;;) {
        if (i >= pings.length || t0 + pings[i] >= untilEpochMs) {
            sleepUntil(untilEpochMs);
            return null;
        }
        sleepUntil(t0 + pings[i]);
        const next = i + 1 < pings.length ? t0 + pings[i + 1] : Infinity;
        const timeoutMs = Math.min(next - Date.now(), refetchMax * 1000);
        refetches.add(1);
        refetchSendLag.add(Math.max(0, Date.now() - (t0 + pings[i])));
        const res = challengePageRefetch(baseUrl, token, challengeId, Math.max(20, timeoutMs));
        if (res.cancelled) {
            // Cancelled by ping i+1, which triggers the next refetch right away
            refetchCancelled.add(1);
            i++;
            continue;
        }
        i = firstPingAfter(Date.now() - t0);
        if (!res.ok) {
            refetchErrors.add(1);
            if (loggedRefetchErrors++ < 3) {
                console.error(`refetch failed: ${res.error}`);
            }
            continue;
        }
        refetchCompletedLatency.add(res.durationMs);
        if (done && done(res.data)) {
            return res.data;
        }
    }
}

function openSession(data) {
    const quiz = data && data.challenge && data.challenge.quiz;
    return quiz && quiz.userActiveSession && quiz.userActiveSession.state === 'OPEN' ? quiz.userActiveSession : null;
}

// sensibleBet: a bet the app's slider allows (see betting-quiz-rampspike.js)
function sensibleBet(question, available) {
    let max = available;
    if (question.bettingMaxAbsolute != null) {
        max = Math.min(max, question.bettingMaxAbsolute);
    }
    if (question.bettingMaxPercentage != null) {
        max = Math.min(max, Math.floor((available * question.bettingMaxPercentage) / 100));
    }
    let min = 1;
    if (question.bettingMinAbsolute != null) {
        min = Math.max(min, question.bettingMinAbsolute);
    }
    if (question.bettingMinPercentage != null) {
        min = Math.max(min, Math.ceil((available * question.bettingMinPercentage) / 100));
    }
    return max < min ? 0 : min + Math.floor(Math.random() * (max - min + 1));
}

export function admin(data) {
    const idx = exec.scenario.iterationInTest * shards + shard;
    if (idx >= churches.length) {
        return;
    }
    const church = churches[idx];
    const s = schedule[idx];
    const t0 = data.t0;

    sleepUntil(t0 + s.openMs);
    const started = startBetting(baseUrl, church.adminToken, quizId, church.churchId,
        `Gamenight ${gamenight} Betting for Load Test Church ${idx + 1}`);
    if (started.sessionId) {
        sessionsCreated.add(1);
    }
    if (!started.ok) {
        adminFailures.add(1);
        console.error(`church ${idx + 1}: startBetting failed (session ${started.sessionId})`);
        if (!started.sessionId) {
            return;
        }
    }

    sleepUntil(t0 + s.lockMs);
    const locked = lockBetting(baseUrl, church.adminToken, started.sessionId, s.doubleClick);
    if (!locked.ok) {
        adminFailures.add(1);
        console.error(`church ${idx + 1}: lock failed`);
    }

    sleepUntil(t0 + s.finishMs);
    const before = Date.now();
    const finished = finishBetting(baseUrl, church.adminToken, started.sessionId, s.doubleClick, finishTimeout);
    finishDuration.add(Date.now() - before);
    if (finished.ok && finished.state === 'FINISHED') {
        sessionsFinished.add(1);
    } else {
        adminFailures.add(1);
        console.error(`church ${idx + 1}: finish failed (state ${finished.state})`);
    }
}

export function player(data) {
    const idx = exec.scenario.iterationInTest * shards + shard;
    if (idx >= players.length) {
        return;
    }
    const p = players[idx];
    const s = schedule[p.churchIdx];
    const t0 = data.t0;
    const r = rng(seed * 7919 + idx);
    const refetcher = refetchOn && r() < refetchFraction;
    // This client's pings arrive this much later than the server sends them
    const pingT0 = t0 + r() * pingJitterMs;
    const endMs = t0 + lastFinishMs + finishMargin * 1000;

    // Connect before the gamenight starts
    sleepUntil(t0 - lead * 1000 + r() * preconnectWindow * 1000);
    const token = maybeAuthDance(baseUrl, p.token, p.auth0Token);
    let page = parseResponse(coldLoad(baseUrl, token, () => challengePage(baseUrl, token, challengeId)));

    // Wait until the church's session is visible: through the refetches, or
    // (without them) by reopening the page once the admin started.
    // The timeline has the grant ping at a planned time, but the server sends
    // it when the async grant job completes, which can be later. So once that
    // planned ping passed, retry every GRANT_RETRY seconds (still following
    // other pings) as a stand-in for the real completion ping.
    const grantedAt = t0 + s.openMs + 500;
    const waitUntil = grantedAt + sessionWaitMax * 1000;
    const visible = (d) => openSession(d) !== null;
    let session = openSession(page);
    if (!session && refetcher) {
        page = listen(pingT0, token, grantedAt + pingJitterMs + 1000, visible) || page;
        session = openSession(page);
    }
    while (!session && Date.now() < waitUntil) {
        if (refetcher) {
            page = listen(pingT0, token, Date.now() + grantRetry * 1000, visible) || page;
            session = openSession(page);
            if (session) {
                break;
            }
        } else {
            sleepUntil(Math.max(grantedAt, Date.now() + grantRetry * 1000));
        }
        page = parseResponse(challengePage(baseUrl, token, challengeId)) || page;
        session = openSession(page);
    }
    if (!session) {
        sessionWaitTimeouts.add(1);
        console.error(`player ${idx}: no open session after ${sessionWaitMax}s (church ${p.churchIdx + 1})`);
        return;
    }
    timeToSessionVisible.add(Math.max(0, Date.now() - grantedAt));

    const pause = (seconds) => {
        if (refetcher) {
            listen(pingT0, token, Date.now() + seconds * 1000, null);
        } else {
            sleep(seconds);
        }
    };

    pause((2 + r() * 8) * thinkScale); // reaction time
    const submission = startBettingSession(baseUrl, token, session.id);
    if (!submission || !submission.orderedQuestions || submission.orderedQuestions.length === 0) {
        quizFailures.add(1);
        console.error(`player ${idx}: failed to start session ${session.id}`);
        return;
    }
    let available = Math.max(0, (page && page.myCurrentProject && page.myCurrentProject.myPoints) || 0);

    for (const question of submission.orderedQuestions) {
        const thinkSeconds = (5 + r() * 10) * thinkScale;
        pause(thinkSeconds);
        const answers = question.predefinedAnswers;
        const correct = answers.find((a) => a.answerOrder === 1);
        const wrong = answers.filter((a) => a.answerOrder !== 1);
        const answerId = r() < correctRate || wrong.length === 0 ? correct.id : wrong[Math.floor(r() * wrong.length)].id;
        const bet = r() < noBetFraction ? 0 : sensibleBet(question, available);
        const result = submitBet(baseUrl, token, submission.id, question.id, answerId, bet, Math.round(thinkSeconds));
        if (result.status !== 'accepted') {
            quizFailures.add(1);
            console.error(`player ${idx}: bet ${bet} of ${available} not accepted: ${result.message}`);
            return;
        }
        if (bet > 0) {
            betsAccepted.add(1);
            available -= bet;
        } else {
            betsNone.add(1);
        }
    }

    const finalized = finalizeQuiz(baseUrl, token, submission.id);
    if (!finalized) {
        quizFailures.add(1);
        return;
    }
    quizCompletions.add(1);
    check(finalized, { 'quiz completed': (f) => Boolean(f.completedAt) });

    // The app stays open: other churches' pings keep arriving until the end
    if (refetcher) {
        listen(pingT0, token, endMs, null);
    }
}

export function setup() {
    if (churches.length === 0) {
        throw new Error('config.json has no churches: run bench/betting_box.sh reset-multi, then tokengen -church-prefix CH01LOADTESTCHURCH');
    }
    if (shards > 1 && !__ENV.START_EPOCH_MS) {
        throw new Error('SHARDS > 1 needs the same START_EPOCH_MS on every shard');
    }
    const start = __ENV.START_EPOCH_MS ? parseInt(__ENV.START_EPOCH_MS) : Date.now();
    if (start < Date.now()) {
        console.warn(`START_EPOCH_MS is ${Math.round((Date.now() - start) / 1000)}s in the past: raise LOADTEST_START_DELAY`);
    }
    const t0 = start + lead * 1000;

    const opens = schedule.map((s) => s.openMs).sort((a, b) => a - b);
    let peak = 0;
    for (let i = 0, j = 0; i < opens.length; i++) {
        while (opens[i] - opens[j] >= 1000) j++;
        peak = Math.max(peak, i - j + 1);
    }
    console.log(
        `Multi-session gamenight: ${churches.length} churches, ${players.length} players, shard ${shard + 1}/${shards}; ` +
        `T0 ${new Date(t0).toISOString()} (in ${Math.round((t0 - Date.now()) / 1000)}s), finish ~T0+${sessionDuration}s; ` +
        `peak ${peak} sessions opening per second; ${pings.length} pings (${coalesceMs ? `coalesce ${coalesceMs}ms` : 'all delivered'}); ` +
        `refetch ${refetchOn ? refetchFraction : 0}; target ${baseUrl}`
    );
    return { t0 };
}
