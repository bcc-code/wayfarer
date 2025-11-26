<script setup lang="ts" generic="Tab extends { label: string; value: string }">
withDefaults(
  defineProps<{
    tabs: Tab[]
    variant?: 'primary' | 'secondary'
  }>(),
  {
    variant: 'primary',
  },
)

const modelValue = defineModel<Tab['value']>({ required: true })
</script>

<template>
  <div class="p-small rounded-navigation border-border-default flex border">
    <DesignButton
      v-for="tab in tabs"
      :key="tab.value"
      :variant="modelValue == tab.value ? variant : 'tertiary'"
      @click="modelValue = tab.value"
    >
      {{ tab.label }}
    </DesignButton>
  </div>
</template>
