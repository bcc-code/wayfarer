<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    consent: ConsentStatusFields
    showDate?: boolean
  }>(),
  { showDate: true },
)

const status = computed(() => consentPublicationStatus(props.consent))
const label = computed(() => CONSENT_STATUS_LABELS[status.value])
const color = computed(() => CONSENT_STATUS_COLORS[status.value])
const date = computed(() => consentStatusDate(props.consent))
</script>

<template>
  <div class="flex items-center gap-2">
    <UBadge variant="soft" :color>{{ label }}</UBadge>
    <span v-if="showDate && date" class="text-muted text-sm">
      {{ formatDateTime(date) }}
    </span>
  </div>
</template>
