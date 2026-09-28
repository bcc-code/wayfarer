<script setup lang="ts">
/**
 * Legacy redirect: the team list is per-project now, and there is no current
 * project on this URL, so it lands on the project picker.
 */
definePageMeta({
  permission: 'teams:view',
  layout: 'admin',
  redirect: { name: 'admin-projects' },
})
</script>
