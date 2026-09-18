<script setup lang="ts">
const { state, settle } = useConfirm()
const { t } = useI18n()

const open = computed({
  get: () => state.open,
  // Dismissing via overlay/escape resolves as "cancelled".
  set: (value: boolean) => {
    if (!value) settle(false)
  },
})
</script>

<template>
  <UModal
    v-model:open="open"
    :title="state.title"
    :ui="{ content: 'max-w-md' }"
  >
    <template #body>
      <div class="flex gap-3">
        <UIcon
          v-if="state.icon"
          :name="state.icon"
          class="text-error mt-0.5 size-5 shrink-0"
        />
        <p v-if="state.description" class="text-muted text-sm">
          {{ state.description }}
        </p>
      </div>
    </template>
    <template #footer>
      <div class="flex w-full justify-end gap-2">
        <UButton variant="soft" color="neutral" @click="settle(false)">
          {{ state.cancelLabel ?? t('admin.common.cancel') }}
        </UButton>
        <UButton :color="state.color" @click="settle(true)">
          {{ state.confirmLabel ?? t('admin.common.delete') }}
        </UButton>
      </div>
    </template>
  </UModal>
</template>
