<script setup lang="ts">
import { useSubmitFeedbackMutation } from '~/api/generated'

const props = defineProps<{
  projectId?: string
}>()

const { locale } = useI18n()

const open = ref(false)
const message = ref<string>()
const canContactMe = ref(false)
const showValidationError = ref(false)
const hasSent = ref(false)
const isSubmitting = ref(false)
const errorMessage = ref<string>()
const showLengthError = ref(false)
const maxMessageLength = 2000

function reset() {
  open.value = false
  message.value = undefined
  canContactMe.value = false
  showValidationError.value = false
  showLengthError.value = false
  hasSent.value = false
  errorMessage.value = undefined
}

const { executeMutation: submitFeedback } = useSubmitFeedbackMutation()

watch(message, (m) => {
  if (m?.length) {
    showValidationError.value = false
  }
  if (m && m.length <= maxMessageLength) {
    showLengthError.value = false
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
  }
}

const { track } = useAnalytics()

async function handleSubmit() {
  const trimmedMessage = message.value?.trim()

  if (!trimmedMessage?.length) {
    showValidationError.value = true
    return
  }

  if (trimmedMessage.length > maxMessageLength) {
    showLengthError.value = true
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
        projectId: props.projectId,
      },
    })

    if (result.error) {
      errorMessage.value = result.error.message
      return
    }

    track(AnalyticsEvent.FeedbackSubmitted)

    hasSent.value = true
  } finally {
    isSubmitting.value = false
  }
}

function closeFeedback(cb: () => void) {
  cb()
  reset()
}

watch(open, (isOpen) => {
  if (isOpen) {
    track(AnalyticsEvent.FeedbackModalOpened)
  } else {
    track(AnalyticsEvent.FeedbackModalClosed)
  }
})
</script>

<template>
  <div class="contents">
    <DesignDrawer v-model:open="open" :title="$t('feedback.title')">
      <slot>
        <DesignPanel class="flex flex-col">
          <button
            class="flex text-start items-center justify-between gap-4 px-3 py-4 disabled:opacity-25 disabled:cursor-not-allowed"
          >
            <div>
              <p class="text-label text-text-default mb-1">
                {{ $t('feedback.cta') }}
              </p>
              <p class="text-caption text-text-muted">
                {{ $t('feedback.description') }}
              </p>
            </div>
            <IconChevronRight class="size-6 shrink-0" />
          </button>
        </DesignPanel>
      </slot>
      <template #content="{ close }">
        <Transition
          mode="out-in"
          enter-active-class="transition ease-out duration-400"
          enter-from-class="opacity-0 scale-95"
          leave-active-class="transition ease-out duration-200"
          leave-to-class="opacity-0 scale-95"
        >
          <form
            v-if="!hasSent"
            class="flex flex-col grow"
            @submit.prevent="handleSubmit"
          >
            <p class="text-label mx-4 mb-4">
              {{ $t('feedback.label') }}
            </p>
            <DesignTextarea
              v-model="message"
              :error="showValidationError || showLengthError"
              :placeholder="$t('feedback.placeholder')"
              :maxlength="maxMessageLength"
            />
            <Transition
              enter-active-class="transition ease-out duration-200"
              enter-from-class="opacity-0 -translate-y-4"
            >
              <p
                v-if="showValidationError || showLengthError"
                class="text-accent-negative text-label my-2 px-4"
              >
                {{
                  showValidationError
                    ? $t('feedback.validationError')
                    : $t('feedback.lengthError', { max: maxMessageLength })
                }}
              </p>
            </Transition>
            <Transition
              enter-active-class="transition ease-out duration-200"
              enter-from-class="opacity-0 -translate-y-4"
            >
              <p
                v-if="errorMessage"
                class="text-accent-negative text-label my-2 px-4"
              >
                {{ errorMessage }}
              </p>
            </Transition>
            <!-- <div class="px-medium py-6">
              <UCheckbox
                v-model="canContactMe"
                :ui="{
                  label: 'text-label',
                  base: 'size-11 rounded-lg!',
                  container: 'h-auto',
                  root: 'gap-2',
                  indicator: 'bg-accent text-on-accent',
                }"
                :label="$t('feedback.canContactMe')"
              />
            </div> -->
            <p class="px-medium text-caption text-text-muted pt-medium">
              {{ $t('feedback.disclosure') }}
            </p>
            <div class="p-medium grow-0 mt-auto">
              <DesignButton
                size="large"
                class="w-full"
                type="submit"
                :disabled="isSubmitting"
              >
                {{ $t('feedback.send') }}
              </DesignButton>
            </div>
          </form>
          <div
            v-else
            class="flex flex-col items-center justify-center grow p-default text-center mb-24"
          >
            <IconSupport class="size-12 mb-6" />
            <p class="text-title mb-2">{{ $t('feedback.thankYou') }}</p>
            <p class="text-label text-text-muted mb-8">
              {{ $t('feedback.thankYouDescription') }}
            </p>
            <DesignButton class="grow-0" @click="() => closeFeedback(close)">
              {{ $t('feedback.close') }}
            </DesignButton>
          </div>
        </Transition>
      </template>
    </DesignDrawer>
  </div>
</template>
