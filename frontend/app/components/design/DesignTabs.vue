<script
  setup
  lang="ts"
  generic="Tab extends { label: string; value: string; enabled?: boolean }"
>
import { cva } from 'cva'

const props = withDefaults(
  defineProps<{
    tabs: Tab[]
    variant?: 'primary' | 'secondary'
  }>(),
  {
    variant: 'primary',
  },
)

const modelValue = defineModel<Tab['value']>({ required: true })

const containerClasses = cva(
  'rounded-navigation border-border-default flex border',
  {
    variants: {
      variant: {
        primary: 'p-small',
        secondary: 'p-navigation-inset',
      },
    },
  },
)

const buttonClasses = cva('text-accent-contrast relative grow', {
  variants: {
    variant: {
      primary:
        'rounded-button-medium px-button-medium-horizontal py-button-medium-vertical text-label',
      secondary:
        'rounded-navigation-inset px-button-small-horizontal py-button-small-vertical text-tiny',
    },
    active: {
      true: 'bg-background-raised gradient-border',
    },
  },
})

const enabledTabs = computed(() =>
  props.tabs.filter((tab) => tab.enabled !== false),
)
</script>

<template>
  <div :class="containerClasses({ variant })">
    <button
      v-for="tab in enabledTabs"
      :key="tab.value"
      :class="buttonClasses({ variant, active: tab.value == modelValue })"
      @click="modelValue = tab.value"
    >
      <slot name="tab" :tab>{{ tab.label }}</slot>
    </button>
  </div>
</template>
