<script setup lang="ts">
defineProps<{ responseCount: number }>()
</script>

<template>
  <!--
    JSON answers hold arbitrary structured data, so there is no shape to count.
    The question stays on the page with its response count so the numbering
    matches the quiz the presenter is reading from.
  -->
  <p class="text-muted text-sm">
    {{ responseCount }} svarte. Denne spørsmålstypen kan ikke oppsummeres
    automatisk.
  </p>
</template>
