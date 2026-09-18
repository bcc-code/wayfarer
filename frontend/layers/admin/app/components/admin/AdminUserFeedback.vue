<script setup lang="ts">
import { useSubmitFeedbackMutation } from '~/api/generated'

const { locale, t } = useI18n()
const route = useRoute()

const isOpen = ref(false)
const message = ref('')
const canContactMe = ref(false)
const isSubmitting = ref(false)
const hasSent = ref(false)
const errorMessage = ref<string>()
const maxMessageLength = 2000

const open = () => {
  isOpen.value = true
}

function reset() {
  message.value = ''
  canContactMe.value = true
  isSubmitting.value = false
  hasSent.value = false
  errorMessage.value = undefined
}

function handleClose() {
  isOpen.value = false
  // Reset after modal closes
  setTimeout(reset, 200)
}

const { executeMutation: submitFeedback } = useSubmitFeedbackMutation()

watch(message, (m) => {
  if (m?.length) {
    errorMessage.value = undefined
  }
})

function getDeviceMetadata() {
  const platform =
    // @ts-expect-error userAgentData is not yet in all TypeScript DOM types
    navigator.userAgentData?.platform || navigator.platform || 'unknown'

  return {
    userAgent: navigator.userAgent,
    platform,
    screenWidth: window.screen.width,
    screenHeight: window.screen.height,
    appVersion: useRuntimeConfig().public.appVersion,
    locale: locale.value,
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    contextUrl: route.fullPath,
  }
}

async function handleSubmit() {
  const trimmedMessage = message.value?.trim()

  if (!trimmedMessage?.length) {
    errorMessage.value = t('feedback.validationError')
    return
  }

  if (trimmedMessage.length > maxMessageLength) {
    errorMessage.value = t('feedback.lengthError', { max: maxMessageLength })
    return
  }

  isSubmitting.value = true
  errorMessage.value = undefined

  try {
    const result = await submitFeedback({
      input: {
        message: trimmedMessage,
        canContactMe: canContactMe.value,
        device: getDeviceMetadata(),
        tags: ['admin'],
      },
    })

    if (result.error) {
      errorMessage.value = result.error.message
      return
    }

    hasSent.value = true
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div class="flex items-center">
    <slot :open>
      <UButton variant="link" @click="open">
        <UIcon name="lucide:message-square" />
        {{ $t('feedback.cta') }}
      </UButton>
    </slot>
    <UModal
      v-model:open="isOpen"
      :title="$t('feedback.title')"
      :description="$t('feedback.description')"
      :dismissible="false"
      :close="false"
    >
      <template #body>
        <div v-if="!hasSent" class="flex flex-col gap-4">
          <UFormField :error="errorMessage" :help="$t('feedback.disclosure')">
            <UTextarea
              v-model="message"
              :rows="6"
              :placeholder="$t('feedback.placeholder')"
              :maxlength="maxMessageLength"
              class="w-full"
            />
          </UFormField>
          <div class="flex justify-end gap-2">
            <UButton variant="ghost" @click="handleClose">
              {{ $t('admin.common.cancel') }}
            </UButton>
            <UButton :loading="isSubmitting" @click="handleSubmit">
              {{ $t('feedback.send') }}
            </UButton>
          </div>
        </div>
        <div v-else class="flex flex-col items-center gap-4 py-4">
          <UIcon name="lucide:check-circle" class="size-12 text-green-500" />
          <p class="text-lg font-medium">{{ $t('feedback.thankYou') }}</p>
          <p class="text-muted text-center">
            {{ $t('feedback.thankYouDescription') }}
          </p>
          <UButton @click="handleClose">{{ $t('feedback.close') }}</UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>
