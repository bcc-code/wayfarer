<script setup lang="ts">
/**
 * Legacy redirect: the score journal is per-project now, and there is no
 * current project on this URL, so it lands on the project picker.
 */
definePageMeta({
  permission: 'scores:view',
  layout: 'admin',
  redirect: { name: 'admin-projects' },
})
</script>
