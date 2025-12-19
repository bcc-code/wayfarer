<script setup lang="ts">
import { useSubmitFeedbackMutation } from '~/api/generated'

const open = ref(false)
const message = ref<string>()
const canContactMe = ref(false)
const showValidationError = ref(false)
const hasSent = ref(false)
const isSubmitting = ref(false)
const errorMessage = ref<string>()

const { executeMutation: submitFeedback } = useSubmitFeedbackMutation()

watch(open, (isOpen) => {
  if (!isOpen) {
    message.value = undefined
    canContactMe.value = false
    showValidationError.value = false
    hasSent.value = false
    errorMessage.value = undefined
  }
})

watch(message, (m) => {
  if (m?.length) {
    showValidationError.value = false
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
  }
}

async function handleSubmit() {
  if (!message.value) {
    showValidationError.value = true
    return
  }

  isSubmitting.value = true
  errorMessage.value = undefined

  try {
    const result = await submitFeedback({
      input: {
        message: message.value,
        canContactMe: canContactMe.value,
        device: getDeviceMetadata(),
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
          <Icon name="IconChevronRight" class="size-6 shrink-0" />
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
          <p class="text-label text-text-muted mx-4 mb-4">
            {{ $t('feedback.label') }}
          </p>
          <DesignTextarea
            v-model="message"
            :error="showValidationError"
            :placeholder="$t('feedback.placeholder')"
          />
          <Transition
            enter-active-class="transition ease-out duration-200"
            enter-from-class="opacity-0 -translate-y-4"
          >
            <p
              v-if="showValidationError"
              class="text-accent-negative text-label my-2 px-4"
            >
              {{ $t('feedback.validationError') }}
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
          <DesignButton class="grow-0" @click="close">
            {{ $t('feedback.close') }}
          </DesignButton>
        </div>
      </Transition>
    </template>
  </DesignDrawer>
</template>
