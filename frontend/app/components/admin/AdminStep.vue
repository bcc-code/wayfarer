<script setup lang="ts">
defineProps<{
  step: number
  title: string
  description?: string
  active?: boolean
}>()
</script>

<template>
  <section
    :id="`step-${step}`"
    :data-active="active"
    class="grid grid-cols-[75px_1fr] gap-4 items-baseline py-12 data-[active=false]:pointer-events-none data-[active=false]:opacity-50"
  >
    <div>
      <p
        :data-active="active"
        class="text-2xl rounded-full grid place-items-center bg-accented aspect-square size-12 data-[active=true]:bg-primary data-[active=true]:text-neutral-900"
      >
        {{ step }}
      </p>
    </div>
    <div>
      <h2 class="text-2xl font-bold mb-2">{{ title }}</h2>
      <p v-if="description">
        {{ description }}
      </p>
      <div class="mt-6">
        <slot />
      </div>
    </div>
  </section>
</template>
