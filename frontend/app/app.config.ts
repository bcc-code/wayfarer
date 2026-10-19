/**
 * Nuxt UI component theming.
 *
 * The surface treatment is adapted from the Nuxt UI calendar template
 * (github.com/nuxt-ui-templates/calendar), rebuilt on stock Tailwind: the
 * colours are theme tokens (`bg-control`, `bg-well`) declared in
 * `assets/styles/main.css`.
 *
 * Anything that floats over the page — a dialog, a menu, the sidebar — is
 * opaque. The reference frosts them, and this panel did too until reading a
 * dialog meant reading it against whatever happened to be behind it.
 *
 * This config is global, but in practice it only dresses the admin panel: the
 * user-facing app runs on the Design* system and uses barely any U* components.
 *
 * Not taken from the reference: its `--ui-radius: 0.5rem`. Admin stays at
 * 0.3rem (admin.css) because every `rounded-*` in the app derives from that
 * token, including the user-facing Design* components.
 */

// The hairline that frames a piece of glass. Surfaces only — a form control
// draws the theme's `ring-accented` instead, which you can actually see.
const ring = 'ring ring-black/8 dark:ring-white/10'
// Drops the blur behind the overlay for anyone who asks for less
// transparency. The variant is declared in main.css.
const solid =
  'reduceTransparency:bg-default reduceTransparency:backdrop-blur-none'
// A surface that floats over the body. Opaque: text on it has to be readable
// whatever it is covering.
const content = `bg-default ${ring} shadow-2xl`
// The same hairline in border form, for a section ruled off inside a surface.
const border = 'border-black/8 dark:border-white/10'
// Rules inside glass. `divide-default` is an opaque border colour, which reads
// as painted on once there is a backdrop showing through behind it.
const divide = 'divide-black/8 dark:divide-white/8'
// A translucent overlay filling the viewport. It hazes rather than frosts, so
// it carries a blur of its own rather than the material's.
const overlay = `bg-glass backdrop-blur-sm ${solid}`
// A small surface lifted off the chrome behind it: opaque white on light,
// where the lift is carried by the shadow, and a thin wash on dark, where a
// shadow would be invisible and the ring does the work instead. This is the
// move the reference dashboards use for anything currently selected.
const raised =
  'bg-default shadow-sm dark:bg-control-hover dark:shadow-none ' + ring
// The same, addressed at a pseudo-element for components that paint their
// surface on `before` rather than on the element itself.
const raisedBefore =
  'before:bg-default before:shadow-sm before:ring before:ring-black/8 dark:before:bg-control-hover dark:before:shadow-none dark:before:ring-white/10'

export default defineAppConfig({
  ui: {
    colors: {
      // The admin panel resolves `--ui-primary` to ink (admin.css), so the ramp
      // behind it is only reached by a `primary-<shade>` utility. Keeping it on
      // the neutral scale means such a utility lands in the same family rather
      // than reintroducing a hue nothing else uses.
      primary: 'zinc',
      neutral: 'zinc',
      // Semantic palettes for the admin panel. Softer hues than the Nuxt UI
      // defaults; the shade they resolve to is set per mode in admin.css.
      success: 'emerald',
      error: 'rose',
      warning: 'amber',
      info: 'blue',
    },
    badge: {
      // A tag reads as a tag rather than a small button when it is fully
      // rounded. The radius lives on the size variants in the theme, so it has
      // to be beaten from a compound variant — a slot would be applied before
      // them and lose. The horizontal padding grows with it: a full radius eats
      // its own corners, and `px-2` leaves the label touching them.
      compoundVariants: [
        { size: 'xs', class: { base: 'rounded-full px-2' } },
        { size: 'sm', class: { base: 'rounded-full px-2.5' } },
        { size: 'md', class: { base: 'rounded-full px-2.5' } },
        { size: 'lg', class: { base: 'rounded-full px-3' } },
        { size: 'xl', class: { base: 'rounded-full px-3' } },
      ],
    },
    button: {
      slots: {
        base: 'cursor-pointer',
      },
      variants: {
        // The theme's `xs` is 24px tall — the smallest target WCAG 2.2 allows,
        // with nothing to spare, and an icon-only one is the easiest thing in
        // the panel to miss. A floor per size leaves the padding and the text
        // alone and only grows the ones that come out too small.
        size: {
          xs: { base: 'min-h-7 min-w-7' },
          sm: { base: 'min-h-8 min-w-8' },
        },
      },
      compoundVariants: [
        // The reference's secondary action: a white chip with a hairline and a
        // shadow under it, standing beside the one solid button on the page.
        {
          color: 'neutral',
          variant: 'outline',
          class: `${ring} shadow-sm dark:shadow-none`,
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
    card: {
      // Only the deltas: `extend` concatenates onto the theme's own classes and
      // tailwind-merge settles the conflicts, so the ring and divide colours
      // swap to the glass hairline while the rest of the variant survives.
      slots: {
        root: 'shadow-sm dark:shadow-none',
      },
      variants: {
        variant: {
          outline: { root: `${ring} ${divide}` },
          subtle: { root: `${ring} ${divide}` },
          soft: { root: divide },
        },
      },
    },
    checkbox: {
      slots: {
        // `size-5!` and the larger label are Wayfarer's own; the rounding
        // comes from the reference. The hairline does not: a control you are
        // meant to find and click needs an edge you can see, so the box keeps
        // the theme's `ring-accented` rather than the glass hairline. Testers
        // could not make out an unchecked box against the panel behind it.
        base: ['size-5!', 'rounded-xs'],
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
    input: {
      variants: {
        variant: {
          // Only the lift. The edge stays the theme's `ring-accented`, for the
          // same reason as the checkbox above — and it is what the untouched
          // textarea and select already draw, so every field in a form now
          // has the same edge.
          outline: 'shadow-xs dark:shadow-none',
        },
      },
    },
    inputNumber: {
      // Stacked chevrons at the end of the field, never the horizontal −/+
      // pair. The horizontal variant centres the value between two buttons
      // that are each as wide as the number itself, so a row of them reads
      // as a toolbar rather than as fields, and the value no longer lines up
      // with the plain inputs above and below it in the same form.
      //
      // `defaultVariants` here is read by Nuxt UI's `useComponentProps`,
      // which resolves every prop not written on the tag — so this is a real
      // default for the `orientation` prop, not only for the theme's slots.
      defaultVariants: {
        orientation: 'vertical',
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
        // `primary` is monochrome (admin.css), so the active link's text no
        // longer separates it from its neighbours on colour alone. The
        // reference dashboards answer this the same way: the current item is a
        // chip raised off the rail rather than a coloured label. `ring` and
        // `shadow` are separate box-shadow layers, so both compose on the one
        // pseudo-element that paints the link's surface.
        {
          disabled: false,
          active: true,
          variant: 'pill',
          class: {
            link: raisedBefore,
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
    table: {
      slots: {
        // The header is a band rather than a row of bold text on the same
        // surface as the data, which is what separates the reference's tables
        // from a list that happens to have a first row.
        th: 'bg-elevated/40 first:rounded-s-md last:rounded-e-md',
        tbody: divide,
        tr: divide,
      },
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
            // The same raised chip the sidebar marks its current item with, so
            // a segmented control and the nav say "you are here" the same way.
            indicator: raised,
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
