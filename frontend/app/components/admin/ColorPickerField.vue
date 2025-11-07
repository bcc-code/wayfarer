<script setup lang="ts">
const color = defineModel<string>({ required: true })
const chip = computed(() => ({ backgroundColor: color.value }))
</script>

<template>
  <UInput v-model="color">
    <template #leading>
      <UPopover :ui="{ content: 'p-1' }">
        <button :style="chip" class="size-4 rounded" />
        <template #content>
          <UColorPicker v-model="color" class="p-2" />
        </template>
      </UPopover>
    </template>
  </UInput>
</template>
