<script setup lang="ts">
const message = ref<string>()
const showValidationError = ref(false)
const hasSent = ref(false)

watch(message, (m) => {
  if (m?.length) {
    showValidationError.value = false
  }
})

function handleSubmit() {
  if (!message.value) {
    showValidationError.value = true
    return
  }

  hasSent.value = true
}
</script>

<template>
  <DesignDrawer :title="$t('feedback.title')">
    <slot>
      <DesignPanel class="flex flex-col" v-bind="$attrs">
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
    <template #content>
      <form
        v-if="!hasSent"
        class="flex flex-col grow"
        @submit.prevent="handleSubmit"
      >
        <p class="text-label text-text-muted mx-4 mb-6">
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
        <div class="p-medium grow-0 mt-auto">
          <DesignButton size="large" class="w-full" type="submit">
            {{ $t('feedback.send') }}
          </DesignButton>
        </div>
      </form>
      <div v-else class="flex flex-col items-center justify-center">
        <p class="text-title">{{ $t('feedback.thankYou') }}</p>
      </div>
    </template>
  </DesignDrawer>
</template>
