<script setup lang="ts" generic="Tab extends { label: string; value: string }">
defineProps<{
  tabs: Tab[]
}>()

const modelValue = defineModel({ required: true })
</script>

<template>
  <div class="p-small rounded-navigation border-border-default flex border">
    <DesignButton
      v-for="tab in tabs"
      :key="tab.value"
      :variant="modelValue == tab.value ? 'primary' : 'tertiary'"
      @click="modelValue = tab.value"
    >
      {{ tab.label }}
    </DesignButton>
  </div>
</template>
