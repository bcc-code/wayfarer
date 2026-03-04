<script setup lang="ts">
const open = defineModel<boolean>('open', { default: false })

const emit = defineEmits<{
  subscribe: []
  dismiss: []
}>()

const { subscribe, isLoading } = usePushNotifications()
const { track } = useAnalytics()

async function handleSubscribe() {
  track(AnalyticsEvent.NotificationPromptAccepted)
  const result = await subscribe()
  if (result) {
    open.value = false
    emit('subscribe')
  }
}

function handleDismiss() {
  track(AnalyticsEvent.NotificationPromptDismissed)
  open.value = false
  emit('dismiss')
}

const badges = [
  {
    scr: '/images/notification-prompt/badge-1.png',
    alt: 'Badge 1',
    z: 1,
    size: '48px',
    class: '-mx-3 -rotate-15',
  },
  {
    scr: '/images/notification-prompt/badge-2.png',
    alt: 'Badge 2',
    z: 2,
    size: '77px',
    class: '-rotate-5',
  },
  {
    scr: '/images/notification-prompt/badge-3.png',
    alt: 'Badge 3',
    z: 3,
    size: '95px',
    class: '-mx-6',
  },
  {
    scr: '/images/notification-prompt/badge-4.png',
    alt: 'Badge 4',
    z: 2,
    size: '77px',
    class: 'rotate-5',
  },
  {
    scr: '/images/notification-prompt/badge-5.png',
    alt: 'Badge 5',
    z: 1,
    size: '48px',
    class: '-mx-3 rotate-15',
  },
]
</script>

<template>
  <UDrawer
    v-model:open="open"
    :ui="{
      content:
        'bg-background-default rounded-t-modal h-[66dvh] gradient-border ring-0 max-w-xl mx-auto',
      overlay: 'bg-black/50',
    }"
    :set-background-color-on-scale="false"
    :handle="false"
    :dismissible="false"
    :prevent-scroll-restoration="false"
  >
    <template #content>
      <div
        class="p-default h-full flex flex-col gap-list-section-gap overflow-auto"
      >
        <div class="flex h-full flex-col items-center justify-center gap-6">
          <div class="flex flex-col items-center grow gap-6 justify-center">
            <!-- Badge icons row -->
            <div class="flex items-end justify-center">
              <img
                v-for="badge in badges"
                :key="badge.alt"
                :src="badge.scr"
                :alt="badge.alt"
                :class="['rounded-full object-cover', badge.class]"
                :style="{ zIndex: badge.z, height: badge.size }"
              >
            </div>
            <!-- Title and description -->
            <div
              class="flex flex-col items-center gap-2 text-center text-balance"
            >
              <h3 class="text-heading text-text-default">
                {{ $t('notifications.promptTitle') }}
              </h3>
              <p class="text-label text-text-muted">
                {{ $t('notifications.promptDescription') }}
              </p>
            </div>
          </div>

          <!-- Buttons -->
          <div class="flex w-full flex-col gap-3 mt-auto">
            <DesignButton
              size="large"
              variant="primary"
              :loading="isLoading"
              class="w-full"
              @click="handleSubscribe"
            >
              {{ $t('notifications.turnOn') }}
            </DesignButton>
            <DesignButton
              size="large"
              variant="secondary"
              :disabled="isLoading"
              class="w-full"
              @click="handleDismiss"
            >
              {{ $t('notifications.notNow') }}
            </DesignButton>
          </div>
        </div>
      </div>
    </template>
  </UDrawer>
</template>
