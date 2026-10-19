<script setup lang="ts">
const isOnline = useOnline()
</script>

<template>
  <Transition
    enter-active-class="transition duration-200 ease-out motion-reduce:transition-none"
    enter-from-class="opacity-0 -translate-y-2"
    leave-active-class="transition duration-150 ease-in motion-reduce:transition-none"
    leave-to-class="opacity-0 -translate-y-2"
  >
    <div v-if="!isOnline" class="p-list-outside pb-0">
      <InfoBanner role="status">
        {{ $t('offline.notice') }}
      </InfoBanner>
    </div>
  </Transition>
</template>
