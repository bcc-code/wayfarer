<script setup lang="ts">
defineProps<{ responseCount: number }>()
</script>

<template>
  <!-- Kept on the page, uncounted, so question numbering stays intact. -->
  <p class="text-muted text-sm">
    {{ responseCount }} svarte. Denne spørsmålstypen kan ikke oppsummeres
    automatisk.
  </p>
</template>
