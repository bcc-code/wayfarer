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
- [ ] **The large-title collapse is currently switched off.** `PageLayout.vue`
      passes `:animate="false"` to `TitleBar`, which disables the
      title-repositioning half of the effect — only the header height change
      (`min-h-24` → `min-h-20`) survives. Now that the bar is sticky the full
      animation would actually be visible, so flipping this on is a cheap win,
      but it changes the look of every page and is a design call rather than a
      mechanical fix.
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

- [ ] **Dynamic `theme-color`.** The manifest pins `#E8DFA7` while projects are
      individually branded and `colorMode` defaults to dark, so the Android
      toolbar and iOS status bar do not match the project. `applyTheme()` in
      `layers/user/app/layouts/default.vue` already runs on every branding
      change and is the natural place to update the meta tag.
- [ ] **App icon badge** — `navigator.setAppBadge(activeChallengesCount)`. The
      count is already queried in the layout for the nav badge.
- [ ] **iOS launch screens.** No `apple-touch-startup-image`; the
      `minimal-2023` generator preset does not emit them, so iOS shows a blank
      flash on every cold launch.
- [ ] **Manifest gaps**: no `id`, `start_url`, `background_color`,
      `orientation`, `description`, `screenshots`, `categories`,
      `launch_handler`. `screenshots` + `description` are what upgrade Chrome
      from the dismissible mini-infobar to the rich install dialog — upstream of
      everything else here, since an uninstalled app cannot feel app-like.
- [ ] **`apple-mobile-web-app-title`** so the home-screen label is not derived
      from the page title.

## Tier 4 — resilience

- [ ] **Offline handling.** `navigator.onLine` / `useOnline` appear nowhere. The
      service worker precaches the shell so the app _launches_ offline, then
      every GraphQL query fails into `<ErrorState>`. Minimum: an offline banner.
      Better: persist the urql cache so last-seen profile/standings render.
- [ ] **Press feedback beyond buttons.** `useButtonPress`
      (`composables/useGsap.ts:140`) is used only by `DesignButton` and
      `DesignIconButton`. `ChallengeCard`, `LeaderboardItem`,
      `ProfileProjectCard` and the nav tabs have no press state.
- [ ] **Haptics** on challenge completion / achievement unlock via
      `navigator.vibrate`. Android only — iOS Safari does not expose it.

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

## Why the title bar collapse stutters

`TitleBar` takes an `animate` prop, but both call sites (`PageLayout.vue` and
`DesignDrawer.vue`) pass `:animate="false"`, so the animated path is currently
dead code. It was turned off because the collapse snapped rather than glided.
Reading the component, that is exactly what it had to do:

1. **The heading has no transition at all.** `headingClasses` has a base of
   just `'text-text-default'`, so every property it changes jumps in one frame.
2. **Its properties are not animatable anyway.** The two states differ by
   `position: static` → `absolute` (a discrete property — it can only snap),
   by `font-size` 30px → 16px via the `text-heading`/`text-label` utilities,
   and by swapping the anchor from `bottom-3 left-6` to `top-1/2 left-1/2`.
   All of those are layout properties.
3. **Only the actions animate.** `actionsClasses` carries
   `transition-all duration-300 ease-out`, so the action button glides for
   300ms while the title teleports — two halves of one bar on different clocks,
   which reads as a stutter by itself.
4. **It is a threshold, not a scroll link.** `hasScrolled` is `y > 25`, so the
   change fires at a single point instead of tracking the scroll, and jitters
   if you linger around 25px.
5. **The header height change cannot be transitioned either.** `min-h-24` →
   `min-h-20` is layout on an element that is in flow, so animating it would
   reflow everything below it on every frame.

A smooth version has to animate only compositor properties (transform and
opacity) and be driven continuously by scroll offset. Point 5 is the real
constraint: a header whose height changes per frame cannot be smooth while it
occupies space in flow. That rules out a continuous version of the current
structure and points at the way iOS actually does it — a constant-height bar
plus a large title that is ordinary scrolling content.

Options are recorded in the open item under Tier 2; the choice between them is
a design decision about how tall the bar is once scrolled, so it is not taken
here.
