<script setup lang="ts">
import ProfileProjectCard from '#layers/user/app/components/profile/ProfileProjectCard.vue'

/**
 * The project's home screen, rendered by the participant app's own
 * `ProfileProjectCard` — the first thing anyone sees, and where the banner
 * and the palette have to work together.
 *
 * The numbers are sample values. A project's own colours come from the frame
 * around this, which the settings form feeds from the draft rather than from
 * what is stored, so a palette can be judged before it is saved.
 */
defineProps<{
  projectName?: string
  banner?: string | null
}>()

/** Enough digits to show what the counters do to the layout. */
const SAMPLE_SCORE = 1240
const SAMPLE_RANK = 7
</script>

<template>
  <div class="bg-background-default aspect-[9/19.5] overflow-y-auto p-default">
    <ProfileProjectCard
      :project-name="projectName || 'Prosjektnavn'"
      :banner="banner ? { url: banner } : null"
      :score="SAMPLE_SCORE"
      :rank="SAMPLE_RANK"
    />
  </div>
</template>
