<script setup lang="ts">
/**
 * Creating an adjustment is a dialog on the list now (see the note's
 * conventions: a page only when the thing created gets its own detail page).
 * The route stays as a redirect so older links still land somewhere useful.
 */
definePageMeta({
  permission: 'scores:view',
  layout: 'admin',
  // `to.params` is a union of every typed route's params, so the one this
  // route actually has needs naming.
  redirect: (to) => ({
    name: 'admin-projects-projectId-scores',
    params: { projectId: (to.params as { projectId: string }).projectId },
  }),
})
</script>
