<script setup lang="ts">
defineProps<{
  achievement: Partial<Achievement>
}>()

const state = ref<'pending' | 'completed'>('pending')
</script>

<template>
  <div
    class="border-default bg-background-default aspect-1/2 w-[400px] overflow-clip rounded-xl border text-start"
  >
    <div
      class="flex h-full flex-col items-center justify-center gap-6 relative"
    >
      <USelect
        v-model="state"
        :items="[
          { value: 'pending', label: 'Not Completed' },
          { value: 'completed', label: 'Completed' },
        ]"
        class="absolute top-4"
      />
      <div
        :class="[
          'grid aspect-square size-55 place-items-center overflow-hidden rounded-full',
          { 'shadow-large': achievement.achievedAt },
        ]"
      >
        <img
          v-if="achievement.imageCompleted && state === 'completed'"
          :src="achievement.imageCompleted"
          class="size-full object-cover"
        />
        <img
          v-else-if="achievement.imagePending"
          :src="achievement.imagePending"
          class="size-full object-cover"
        />
        <img
          v-else
          src="/images/achievement-placeholder.png"
          class="size-full object-cover"
        />
      </div>
      <div class="flex flex-col items-center gap-1 text-center text-balance">
        <h3 class="text-heading">
          {{ achievement.name }}
        </h3>
        <p class="text-label">
          {{
            state === 'completed'
              ? achievement.descriptionCompleted
              : achievement.descriptionPending
          }}
        </p>
      </div>
      <div
        v-if="state === 'completed'"
        class="rounded-full bg-background-indent py-2 px-3 text-label text-accent-contrast"
      >
        +{{ achievement.points }} {{ $t('points') }}
      </div>
      <div
        v-else
        class="rounded-full bg-background-indent py-2 px-3 text-label text-text-muted"
      >
        {{ $t('givesYouXPoints', { points: achievement.points }) }}
      </div>
    </div>
  </div>
</template>
