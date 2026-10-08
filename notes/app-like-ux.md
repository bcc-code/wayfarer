# Making the user-facing app feel app-like

Audit of `frontend/layers/user` against what separates a PWA that reads as a
native app from one that reads as a web page. The foundation was already
there — `display: standalone`, `viewport-fit=cover`, safe-area insets, push
notifications, an iOS-style collapsing title bar, GSAP press animations — so
everything below is about the details that leak "browser".

Status key: `[x]` done, `[ ]` open, `[~]` partially done.

## Tier 1 — touch behaviour (small CSS/meta changes, large perceptual effect)

- [x] **`overscroll-behavior: none`** on `html, body`. Was absent, so Chrome's
      native pull-to-refresh fired on every downward drag at the top and iOS
      rubber-banded the whole document. Also stops scroll chaining out of
      `DesignDrawer` into the page behind it.
- [x] **`-webkit-tap-highlight-color: transparent`**. Every tap flashed a grey
      box over the target.
- [x] **`user-select: none`** on the app chrome, re-enabled for prose and form
      fields. Removes the iOS long-press selection + magnifier loupe.
- [x] **`-webkit-touch-callout: none`** on images. Long-pressing an achievement
      badge or an install screenshot offered "Save Image".
- [x] **`interactive-widget=resizes-content`** on the viewport meta, so the
      virtual keyboard resizes the layout instead of panning the viewport over
      it. Matters most in the quiz free-text questions.

These live in `layers/user/app/assets/styles/user.css`, not the shared
`app/assets/styles/main.css` — the admin panel is a desktop-shaped app where
text selection and normal overscroll are wanted. `user.css` is imported by the
user layout only.

## Tier 2 — navigation feel

- [x] **Page transitions.** No `pageTransition` or `layoutTransition` was set
      anywhere, so every navigation was an instant swap. Now direction-aware:
      moving deeper pushes, moving back pops, switching tabs cross-fades.
      See "Page transitions" below.
- [x] **Sticky title bar.** `PageLayout.vue` carried a literal
      `<!-- TODO: position sticky -->`. `TitleBar.vue` already implemented the
      iOS large-title → small-title collapse driven by `useWindowScroll`, but
      because the wrapper was not sticky the collapsed bar scrolled out of
      view — the effect was 90% built and never paid off.
- [x] **Large-title collapse**, rebuilt on the iOS structure after the original
      was turned off for stuttering. See "Why the title bar collapse stuttered"
      below.
- [x] **Navigation indicator no longer flashes open from nothing** on mount.
      See "The navigation indicator" below.
- [x] **The bottom navigation fades, scales and slides in and out** rather than
      popping, so routes that hide it (`/settings/**`, `/challenges/:id`) hand
      over as smoothly as the page transition itself. `origin-bottom` keeps the
      bar anchored to the bottom edge while it scales, so it drops away rather
      than shrinking toward its own middle.
- [ ] **Per-tab scroll restoration**, plus tap-the-active-tab-to-scroll-to-top.
      No `scrollBehavior` is configured, so switching tabs loses position.
- [x] **Edge-swipe back** needs nothing built. Verified on iPhone 17 / iOS 26:
      the edge-swipe gesture navigates back inside the installed PWA. What it
      does require is that the _pop_ transition plays when the gesture fires,
      which the depth-based direction detection below gives us for free — a
      gesture-back still moves from a deeper path to a shallower one, so it
      resolves to `page-pop` exactly like tapping the close button. Android's
      system back button behaves the same way.

## Tier 3 — platform integration

- [x] **Dynamic `theme-color`.** The manifest used to pin `#E8DFA7` while
      projects are individually branded and `colorMode` defaults to dark, so
      the Android toolbar and iOS status bar matched neither. The user layout
      now publishes the active project's `backgroundDefault` for the resolved
      colour mode, falling back to the cached theme so it survives a cold start.
      Written with `useSeoMeta({ themeColor })` rather than `useHead` — see
      "Route-name scan" below.
- [x] **App icon badge** — `utils/appBadge.ts`, mirroring
      `activeChallengesCount` onto the home-screen icon. The platform rejects
      while the app is only open in a browser tab, so rejections are swallowed
      rather than surfaced.
- [x] **Manifest gaps** — `id`, `start_url`, `scope`, `description`,
      `categories`, `orientation`, `background_color` and `launch_handler`
      (`focus-existing`) added, and `theme_color` changed from `#E8DFA7` to
      `#222222`. **This is a visible change**: `#222222` is
      `--color-background-default` in the dark theme, which is the default
      colour mode, so the install splash and the pre-branding toolbar no longer
      flash a colour the app never shows. Revert if the cream was deliberate.
- [x] **`apple-mobile-web-app-title`** so the home-screen label is not derived
      from the page title.
- [ ] **Manifest `screenshots`** — still missing, and the one item here that
      actually moves the needle: together with `description` it upgrades Chrome
      from the dismissible mini-infobar to the rich install dialog, which is
      upstream of everything else, since an uninstalled app cannot feel
      app-like. Needs real captures of the running app.
- [ ] **iOS launch screens.** No `apple-touch-startup-image`; the
      `minimal-2023` generator preset does not emit them, so iOS shows a blank
      flash on every cold launch.
- [ ] **Manifest `description` is English** while the default locale is `nb`,
      which is why vite-pwa stamps `lang: "en"`. Consistent as it stands, but
      worth a decision on which language the install dialog should speak.

### The banner shell

`ProjectInfoBanner` is coupled to the project info-message feature —
`projectId`, markdown/html, visibility windows, per-content-hash dismissal in
localStorage — so the offline notice could not reuse it directly without faking
a message and inheriting a "dismiss forever" button that makes no sense for
connectivity. The presentation was extracted to `InfoBanner` instead (border,
icon slot, content slot, optional dismiss) and both now use it, so the two read
as the same object.

### Route-name scan

`test/unit/routes.test.ts` greps every `.vue`/`.ts` file for
`name: '<kebab-case>'` and asserts each one resolves to a real route. A head
meta written as `useHead({ meta: [{ name: 'theme-color', ... }] })` trips it —
a false positive, but tightening the regex to tell head metas from route
references is not worth it. `useSeoMeta({ themeColor })` is the idiomatic Nuxt
API here anyway and has no bare `name:` literal, so it sidesteps the scan
without weakening it. Worth knowing before adding another meta tag in a
component.

## Tier 4 — resilience

- [x] **Offline handling — in-session (2026-10-08).** Scoped deliberately to
      "works while the app is open"; surviving a relaunch was considered and
      declined for now, see below.

      The real find was a bug that existed independently of any offline work.
      All nine query-driven views ordered their branches
      `loading → error → data`, and under urql's `cache-and-network` a failed
      *background refresh* sets `error` while the cached `data` is still
      perfectly good. Going offline on a page you had already visited therefore
      replaced a usable page with a full-screen `ErrorState`. The error branch
      now yields to data (`error && !data`) everywhere, so the error only wins
      when there is genuinely nothing to show.

      `OfflineNotice` (in `PageLayout`, so every page gets it) reports the
      state via `useOnline`.

- [ ] **Offline across a relaunch** — not done, and a different problem.
      urql runs the default *document* `cacheExchange`, which is memory-only,
      and a cold PWA launch is a reload, so nothing survives it. The two routes
      are hand-rolled per-query snapshots to localStorage (precedent:
      `cachedTheme` in the user layout) or migrating to
      `@urql/exchange-graphcache` with IndexedDB. The latter is the general
      answer but is a data-layer migration across one client shared with the
      admin layer, and normalized-cache mistakes surface as subtly wrong data
      rather than errors.

- [ ] **Offline mutations.** Completing a challenge offline still fails. Needs
      a queue plus replay, and backend idempotency — out of scope of the above.

- **A refresh that fails while online is now silent** — the page shows stale
      data and `OfflineNotice` says nothing, because it tracks connectivity
      rather than query state. Accepted: urql retries on the next navigation,
      and the alternative was per-page staleness plumbing. Revisit if stale
      data turns out to mislead anyone.

- [ ] **`LeaderboardItem` styles itself as interactive but is not.** The row
      carries `hover:bg-background-indent active:bg-background-indent`, yet
      neither it nor `LeaderboardList` has a click handler, link or emit. Either
      the rows should do something (open a profile?) or the states should go.
- [ ] **Press feedback on the two surfaces with competing animations.** Left
      alone deliberately, because both already animate the element a press
      would scale: - `QuizAlternative` drives `shake`/`pulse` on the same `buttonRef`. - `DesignTabs` runs its own sliding indicator.
      Both are worth doing, but need a decision on how the animations compose.
- **Haptics — decided against (2026-10-08).** Not a todo; do not re-propose
  without new information.

      `navigator.vibrate` is Android-only. WebKit has never implemented the
      Vibration API, and every iOS browser is WebKit, so Chrome and Firefox on
      iOS lack it too. (An MDN compat issue claims otherwise, but it is a
      one-off report from ~2023 against an unstated version and contradicts
      WebKit's position.)

      iOS haptics are reachable only through a hack: Safari 17.4 added
      `<input type="checkbox" switch>`, and toggling one fires the Taptic
      Engine as a side effect. Apple has since narrowed it twice — 18.4 began
      requiring a user gesture within roughly a one-second grant, and 26.5
      reportedly killed purely programmatic toggling. The only variant said to
      survive renders an invisible switch overlay *under* every haptic element
      so the tap reads as direct switch interaction.

      That overlay is the reason to decline rather than the version support:
      invisible interactive elements beneath every tappable surface sit exactly
      where the press listeners and `NuxtLink` navigation now live, and are a
      good way to reintroduce subtle tap bugs for a nicety. Twice-patched
      behaviour is a maintenance liability.

      If it is ever revisited, Android-only through a single
      `utils/haptics.ts` (shaped like `appBadge.ts` — feature-detect, no-op
      silently) keeps call sites free of the decision.

## Page transitions

Direction is derived from path depth rather than from history state, so it is a
pure function of the two paths and can be unit tested:
`layers/user/app/utils/pageTransition.ts`.

The three tab routes (`/`, `/standings`, `/challenges`) are all treated as
depth 0 so that moving between them cross-fades rather than pushing — they are
siblings even though their segment counts differ. Everything else is its
segment count, so `/challenges/:id` and `/settings/archive` are depth 2 and
push in from the right, and going back pops out to the right.

The transition is applied by a `router.beforeEach` registered in the user
layout's setup rather than by a global middleware file: Nuxt gathers middleware
per layer with extended layers first, so a `*.global.ts` inside
`layers/user/` would run before the root `01.auth.global.ts` (see
`frontend/CLAUDE.md`). The hook skips any route under `/admin`.

### Why the direction is a CSS variable and not a per-route transition name

The first version stored a direction-specific name on each route
(`to.meta.pageTransition = { name: 'page-pop', ... }`). Navigating out worked;
navigating back slid the outgoing page the wrong way.

The cause is in `nuxt/dist/pages/runtime/page.js`:

```js
const hasTransition = !!(
  props.transition ??
  routeProps.route.meta.pageTransition ??
  appPageTransition
);
```

`routeProps.route` is the route **RouterView is currently rendering**. Under
`mode: 'out-in'` the outgoing page is still being rendered while it leaves, so
its transition classes come from the route being _left_, while the incoming
page's come from the route being _entered_. Two different objects, two
different names, and the two halves of a single navigation travel in opposite
directions. (Nuxt's own documented dynamic-transition example sidesteps this by
mutating `.name` on the single shared object from `app.pageTransition`.)

So the name is now constant (`page`) for every route and only
`--page-direction` changes — `1` deeper, `-1` back out, `0` cross-fade in
place. It is set on `document.documentElement` before the navigation resolves,
so both halves read one value that cannot disagree with itself.
`pageTransition.test.ts` pins the symmetry: every push/pop pair must be exact
opposites.

The slide is deliberately short (16px, not a full-width push) and uses
`mode: 'out-in'`. Pages scroll the window rather than an internal container and
differ in height, so overlapping a full-width push would jump the scroll
position. Moving pages to an internal scroll container would unlock a true
full-width push — worth revisiting together with per-tab scroll restoration.

All transitions collapse to a plain instant swap under
`prefers-reduced-motion: reduce`.

## Why the title bar collapse stuttered, and what replaced it

`TitleBar` took an `animate` prop, and both call sites passed
`:animate="false"` — the animated path was dead code, turned off because the
collapse snapped instead of gliding. Reading the component, that is all it
could ever have done:

1. **The heading had no transition at all.** `headingClasses` had a base of
   just `'text-text-default'`, so every property it changed jumped in one frame.
2. **Its properties were not animatable anyway.** The two states differed by
   `position: static` → `absolute` (a discrete property — it can only snap), by
   `font-size` 30px → 16px via the `text-heading`/`text-label` utilities, and by
   swapping the anchor from `bottom-3 left-6` to `top-1/2 left-1/2`. All layout
   properties.
3. **Only the actions animated.** `actionsClasses` carried
   `transition-all duration-300 ease-out`, so the action button glided for 300ms
   while the title teleported — two halves of one bar on different clocks, which
   reads as a stutter by itself.
4. **It was a threshold, not a scroll link.** `hasScrolled` was `y > 25`, so the
   change fired at a single point rather than tracking the scroll, and jittered
   if you lingered around 25px.
5. **The header height could not be transitioned either.** `min-h-24` →
   `min-h-20` is layout on an in-flow element, so animating it would reflow
   everything below on every frame.

Point 5 is the real constraint and is why the fix is structural rather than a
matter of adding a transition. The bar is now a **constant height**, and the two
titles are overlaid on one row and cross-faded: the large left-aligned one fades
out as the compact centred one fades in, driven continuously from scroll offset
over `TITLE_HANDOVER_DISTANCE`. Opacity is the only thing that moves, and
opacity does not reflow.

The first attempt put the large title into page content instead, iOS-style, so
that it scrolled away under the bar on its own. That is smoother still in
principle — the large title needs no JS at all — but it puts the large title on
a different row from the action button, which does not match this app's design.
Keeping them on one row means the large title has to live inside the bar, and a
bar containing the large title cannot also shrink. That is the trade accepted
here: one constant bar height, no shrink, titles aligned with the action.

`PageLayout` owns the scroll-to-opacity mapping; `TitleBar` just renders what it
is given. `size="small"` keeps the plain always-visible heading for
`DesignDrawer`, whose title is the only one a sheet has.

Two slots, because they are genuinely different things:

- `#title` — a custom large title. `pages/index.vue` uses it for the superteam
  badge beside the name.
- `#bar` — content that replaces both titles. `QuizChallenge` uses it for
  `QuizProgress`, which was previously passed through `#title` but is a progress
  indicator that must stay visible while scrolling, not a title.

### Two alignment traps, both from the same cause

The padding lives on a wrapper so that the `<header>` itself is the content box.
Both traps come from forgetting that distinction:

- **Vertical.** Centring the compact title with `absolute top-1/2` resolves
  against the containing block, which is the **padding box** — and the top
  padding carries the safe-area inset
  (`max(env(safe-area-inset-top) + 0.75rem, 3rem)`). The action button, a flex
  item under `items-center`, centres in the **content box**. On a notched phone
  that left the title roughly 18px above the gear it was meant to line up with.
  With the padding moved out to the wrapper the two boxes coincide and `top-1/2`
  is correct.
- **Horizontal.** The compact title is centred on the whole row
  (`absolute inset-x-12`), not on the space left beside the action — centring it
  within a flex sibling would shift it left by half the action's width plus the
  gap.

## The navigation indicator

The sliding highlight behind the active tab is an empty absolutely-positioned
div that `gsap` sizes from the active link's `offsetWidth`/`offsetHeight`. Until
that first measurement it is zero-sized in the nav's top-left corner, so
animating _to_ its first position looks like it grows out of nothing.

It did that in two different situations:

- **On first mount**, because `onMounted` ran the positioning behind a
  `setTimeout(…, 50)` — a guess. Landing before layout meant the link measured
  zero, the indicator was `gsap.set` to zero size, and the "already positioned"
  flag flipped, so the _next_ call animated it open from a point. A zero
  measurement is now rejected outright, and the timer is replaced by a
  `useResizeObserver` on the nav, which fires when the element genuinely has
  size — and again on rotation, or when the tab count changes as the project
  query resolves.
- **On returning from a route that hides the nav** (`/settings/**` sets
  `showNavigation` false), because the nav block unmounts while the layout
  holding the flag does not. Coming back mounted a brand-new zero-size indicator
  while the flag still claimed it had been positioned. Fixed by asking the
  element — `indicatorRef.offsetWidth > 0` — rather than trusting a flag that
  can outlive the DOM it describes.

The indicator also stays `opacity-0` until it has been positioned once, so there
is nothing to see before it is correct.

## The drawer's rounded top

`DesignDrawer` lost the rounded corners and top border of `rounded-t-modal`.

Two separate causes, both the same shape of mistake — an effect painting to the
bar's **rectangular** box, filling the corner region that lies outside the
radius and where the sheet's own background is therefore not painted:

- The **scroll shadow**, a regression introduced while rebuilding `TitleBar`.
  `shadow` originally had no default, so it was `undefined` — falsy — for
  `DesignDrawer`, which never passes it. Giving it a `true` default switched on
  a gradient the drawer had never had.
- The **progressive blur**, which predates that change. Its layers are
  `absolute inset-0` with `backdrop-filter` and no radius, masked _strongest at
  the top_ for `direction="up"`, so they blur the overlay behind the corner and
  fill it in.

Both are switched off for the sheet, which is also what the structure calls for:
`DesignDrawer` renders the bar above its own `overflow-auto` container, so sheet
content scrolls _beside_ the bar, never under it. Neither a scroll shadow nor a
progressive blur has anything to act on. `TitleBar.test.ts` pins both props.
