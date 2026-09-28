<script setup lang="ts">
/**
 * Legacy redirect: creating an adjustment needs a project, which this URL does
 * not carry.
 */
definePageMeta({
  permission: 'scores:view',
  layout: 'admin',
  redirect: { name: 'admin-projects' },
})
</script>
