/**
 * Nuxt UI component theming.
 *
 * The surface treatment is adapted from the Nuxt UI calendar template
 * (github.com/nuxt-ui-templates/calendar), rebuilt on stock Tailwind: the glass
 * colours are theme tokens (`bg-glass`, `bg-control`, `bg-well`) declared in
 * `assets/styles/main.css`, and the material is plain `backdrop-*` utilities
 * rather than the bespoke one the reference defines.
 *
 * This config is global, but in practice it only dresses the admin panel: the
 * user-facing app runs on the Design* system and uses barely any U* components.
 *
 * Not taken from the reference: its `--ui-radius: 0.5rem`. Admin stays at
 * 0.3rem (admin.css) because every `rounded-*` in the app derives from that
 * token, including the user-facing Design* components.
 */

// The hairline that frames a piece of glass. On its own it goes on a control,
// which sits on the chrome it belongs to rather than lifting off it.
const ring = 'ring ring-black/8 dark:ring-white/10'
// Turns the material solid for anyone who asks for less transparency. The
// variant is declared in main.css; applying it here means every surface built
// from `content` opts out in one place.
const solid =
  'reduceTransparency:bg-default reduceTransparency:backdrop-blur-none'
// A surface that floats over the body.
const content = `bg-glass backdrop-blur-xl backdrop-saturate-150 backdrop-brightness-105 ${solid} ${ring} shadow-2xl`
// The same hairline in border form, for a section ruled off inside a surface.
const border = 'border-black/8 dark:border-white/10'
// Rules inside glass. `divide-default` is an opaque border colour, which reads
// as painted on once there is a backdrop showing through behind it.
const divide = 'divide-black/8 dark:divide-white/8'
// A translucent overlay filling the viewport. It hazes rather than frosts, so
// it carries a blur of its own rather than the material's.
const overlay = `bg-glass backdrop-blur-sm ${solid}`

export default defineAppConfig({
  ui: {
    colors: {
      primary: 'emerald',
      neutral: 'zinc',
    },
    button: {
      slots: {
        base: 'cursor-pointer',
      },
      compoundVariants: [
        {
          color: 'neutral',
          variant: 'outline',
          class: ring,
        },
        {
          color: 'neutral',
          variant: 'soft',
          class:
            'bg-control hover:bg-control-hover active:bg-control-hover disabled:bg-control aria-disabled:bg-control',
        },
        {
          color: 'neutral',
          variant: 'ghost',
          class: 'hover:bg-control active:bg-control',
        },
      ],
    },
    checkbox: {
      slots: {
        // `size-5!` and the larger label are Wayfarer's own; the rounding and
        // ring come from the reference.
        base: ['size-5!', 'rounded-xs', ring],
        label: 'text-base leading-tight font-normal',
      },
    },
    chip: {
      slots: {
        base: 'ring-0',
      },
    },
    commandPalette: {
      slots: {
        root: divide,
      },
      variants: {
        // The viewport rules its groups from this variant rather than from the
        // slot, so a slot override alone would lose the merge to it.
        virtualize: {
          false: {
            viewport: divide,
          },
        },
      },
    },
    contextMenu: {
      slots: {
        content,
      },
    },
    dropdownMenu: {
      slots: {
        content,
      },
    },
    formField: {
      slots: {
        labelWrapper: 'justify-start gap-2',
        label: 'grow',
      },
    },
    kbd: {
      compoundVariants: [
        {
          color: 'neutral',
          variant: 'soft',
          class: 'bg-control',
        },
      ],
    },
    modal: {
      slots: {
        content: [content, divide],
      },
      variants: {
        // On a phone the modal is a sheet: the gutter the sidebar keeps on
        // every side, and all the height that leaves. Above `sm` it goes back
        // to the box the theme centres on the viewport. The ring and shadow are
        // restated because the theme paints them from this variant, which lands
        // after the slot they came from.
        fullscreen: {
          false: {
            content: [
              ring,
              'shadow-2xl w-[calc(100vw-1rem)] h-[calc(100dvh-1rem)] sm:w-[calc(100vw-2rem)] sm:h-auto',
            ],
          },
        },
        overlay: {
          true: {
            overlay,
          },
        },
      },
      compoundVariants: [
        {
          fullscreen: false,
          scrollable: false,
          class: {
            content: 'max-h-none',
          },
        },
        {
          fullscreen: false,
          scrollable: true,
          class: {
            overlay: 'p-2 sm:p-4 sm:py-8',
          },
        },
      ],
    },
    navigationMenu: {
      slots: {
        link: 'hover:before:bg-control-hover',
      },
      compoundVariants: [
        {
          disabled: false,
          active: false,
          variant: 'pill',
          class: {
            link: 'hover:before:bg-control',
          },
        },
      ],
    },
    popover: {
      slots: {
        content,
      },
    },
    select: {
      slots: {
        content,
        item: 'data-highlighted:not-data-disabled:before:bg-control',
      },
      variants: {
        variant: {
          soft: 'bg-control hover:bg-control-hover focus:bg-control-hover disabled:bg-control',
        },
      },
    },
    // The reference styles `USidebar`'s `floating` variant. The admin shell uses
    // `UDashboardSidebar` instead — it carries the resizable, persisted width
    // and the mobile slideover that `USidebar` has no companion for — so the
    // same treatment is applied to that component's slots: the root floats off
    // the chrome with a margin and is cut from the same glass.
    dashboardSidebar: {
      slots: {
        // The surface only. The geometry that makes it float (margin, radius,
        // undoing the theme's `min-h-svh` and its `border-e`) is passed as `ui`
        // from `layouts/admin.vue`: it belongs to that one layout rather than
        // to every dashboard sidebar, and an instance `ui` prop is appended
        // after the theme rather than merged into it.
        root: content,
        header: 'px-3',
        body: 'p-2 gap-2',
        footer: ['p-2 border-t', border],
      },
    },
    // The sidebar menu below `lg`, cut from the same glass as the floating
    // sidebar it stands in for.
    slideover: {
      slots: {
        overlay,
        content: [content, 'divide-none sm:shadow-2xl'],
      },
      // The theme insets by 4, the floating sidebar this stands in for by 2, so
      // the gutter is restated per side. It has to be a compound variant: the
      // theme sets the edges from one of its own, which lands after anything on
      // a slot or a variant.
      compoundVariants: [
        {
          side: 'top',
          inset: true,
          class: { content: 'max-h-[calc(100%-1rem)] inset-x-2 top-2' },
        },
        {
          side: 'right',
          inset: true,
          class: { content: 'w-[calc(100%-1rem)] inset-y-2 right-2' },
        },
        {
          side: 'bottom',
          inset: true,
          class: { content: 'max-h-[calc(100%-1rem)] inset-x-2 bottom-2' },
        },
        {
          side: 'left',
          inset: true,
          class: { content: 'w-[calc(100%-1rem)] inset-y-2 left-2' },
        },
      ],
    },
    tabs: {
      slots: {
        trigger: 'w-full rounded-full',
      },
      variants: {
        // Same path as the theme's own pill classes, or `rounded-lg` and
        // `rounded-md` would come later in the merge and win. The track is a
        // well cut into the chrome it sits on, so the indicator riding in it
        // reads as raised rather than sunk.
        variant: {
          pill: {
            list: ['rounded-full gap-0.5 bg-well', ring],
            indicator: 'rounded-full',
          },
        },
      },
      // After the theme's colour compound, which paints the indicator
      // `bg-inverted` and flips the active text. This one is a lighter surface
      // than the track with a shadow under it, so the active trigger keeps its
      // text colour.
      compoundVariants: [
        {
          color: 'neutral',
          variant: 'pill',
          class: {
            indicator: 'bg-white dark:bg-control shadow-sm',
            trigger: [
              'data-[state=active]:text-highlighted',
              'hover:data-[state=inactive]:not-disabled:bg-glass',
              // The theme paints the active pill on `before` whenever the list
              // renders without an indicator, so it stands in for the one the
              // indicator draws, at the same radius.
              'in-[[data-slot=list]:not(:has([data-slot=indicator]))]:data-[state=active]:before:bg-control',
              'in-[[data-slot=list]:not(:has([data-slot=indicator]))]:data-[state=active]:before:rounded-full',
            ],
          },
        },
      ],
    },
  },
})
