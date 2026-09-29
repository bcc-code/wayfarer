<script setup lang="ts">
defineProps<{
  achievement: Partial<Achievement>
}>()

const state = ref<'pending' | 'completed'>('pending')
</script>

<template>
  <div
    class="bg-background-default aspect-[9/19.5] w-full overflow-y-auto rounded-xl text-start p-list-outside"
  >
    <!-- The state switcher is the preview's own control, not part of the
         achievement, so it sits above it rather than floating over the
         image. Sized to content: a fixed height left the badge stranded in
         the middle of an empty box. -->
    <div class="flex h-full flex-col items-center gap-6">
      <USelect
        v-model="state"
        :items="[
          { value: 'pending', label: 'Ikke fullført' },
          { value: 'completed', label: 'Fullført' },
        ]"
      />
      <div class="flex flex-1 flex-col items-center justify-center gap-6">
        <div
          :class="[
            'grid aspect-square size-55 place-items-center overflow-hidden rounded-full',
            { 'shadow-large': state === 'completed' },
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
            {{ achievement.name || 'Achievement Name' }}
          </h3>
          <p
            v-if="state === 'completed' && achievement.descriptionCompleted"
            class="text-label"
          >
            {{ achievement.descriptionCompleted }}
          </p>
          <p
            v-else-if="state === 'pending' && achievement.descriptionPending"
            class="text-label"
          >
            {{ achievement.descriptionPending }}
          </p>
          <p v-else class="text-label text-text-muted">Ingen beskrivelse</p>
        </div>
        <template v-if="achievement.points">
          <div
            v-if="state === 'completed'"
            class="rounded-full bg-background-indent py-2 px-3 text-label text-accent-contrast"
          >
            +{{ formatNumber(achievement.points ?? 0) }} {{ $t('points') }}
          </div>
          <div
            v-else
            class="rounded-full bg-background-indent py-2 px-3 text-label text-text-muted"
          >
            {{
              $t('givesYouXPoints', {
                points: formatNumber(achievement.points ?? 0),
              })
            }}
          </div>
        </template>
      </div>
    </div>
  </div>
</template>
