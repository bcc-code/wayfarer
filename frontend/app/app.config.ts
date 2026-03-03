export default defineAppConfig({
  ui: {
    colors: {
      primary: 'emerald',
      neutral: 'zinc',
    },
    formField: {
      slots: {
        labelWrapper: 'justify-start gap-2',
      },
    },
    checkbox: {
      slots: {
        label: 'text-base leading-tight font-normal',
        base: 'size-5!',
      },
    },
    button: {
      slots: {
        base: 'cursor-pointer',
      },
    },
  },
})
