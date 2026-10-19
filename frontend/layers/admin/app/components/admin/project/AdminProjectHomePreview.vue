<script setup lang="ts">
import ProfileProjectCard from '#layers/user/app/components/profile/ProfileProjectCard.vue'
import ProjectInfoBanner from '#layers/user/app/components/project/ProjectInfoBanner.vue'

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
  /**
   * The saved info message. The HTML is rendered by the server from the
   * stored markdown, so an unsaved edit has none and cannot be shown — the
   * form says so beside the preview.
   */
  infoMessage?: { markdown: string; html: string } | null
}>()

/** Enough digits to show what the counters do to the layout. */
const SAMPLE_SCORE = 1240
const SAMPLE_RANK = 7
</script>

<template>
  <div
    class="bg-background-default gap-medium flex aspect-[9/19.5] flex-col overflow-y-auto p-default"
  >
    <!-- A fixed project id and no time window: the banner hides itself once
         dismissed and outside its dates, and a preview that renders nothing
         answers no question. Dismissing it here cannot reach the real one. -->
    <ProjectInfoBanner
      v-if="infoMessage"
      project-id="preview"
      :info-message="infoMessage"
    />
    <ProfileProjectCard
      :project-name="projectName || 'Prosjektnavn'"
      :banner="banner ? { url: banner } : null"
      :score="SAMPLE_SCORE"
      :rank="SAMPLE_RANK"
    />
  </div>
</template>
