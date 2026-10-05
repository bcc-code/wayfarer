<script setup lang="ts">
/**
 * A push notification as it arrives on a phone.
 *
 * The one preview in the panel with no real component behind it: a
 * notification is drawn by the operating system, not by the app, so this is a
 * mock rather than the participant app's own markup. It is here because the
 * text is otherwise written completely blind — it is the only field that
 * reaches someone who never opened the app.
 *
 * The payload is `{ Title: <name>, Body: <notification text>, Icon: <image> }`
 * for both challenges (`SendChallengeEnrollmentNotification`) and achievements
 * (`SendAchievementNotification`), which is what the two lines below are.
 */
defineProps<{
  title?: string
  body?: string
  icon?: string | null
}>()
</script>

<template>
  <div class="bg-elevated/60 rounded-2xl p-3">
    <div class="flex items-start gap-3">
      <img
        v-if="icon"
        :src="icon"
        alt=""
        class="size-8 shrink-0 rounded-lg object-cover"
      />
      <div
        v-else
        class="bg-accented size-8 shrink-0 rounded-lg"
        aria-hidden="true"
      />
      <div class="min-w-0 flex-1">
        <p class="text-dimmed text-[0.6875rem] uppercase">Interact</p>
        <!-- Two lines, as a phone shows it: past that the text is cut off
             wherever the OS decides, so a longer message is not a longer
             notification. -->
        <p class="text-highlighted truncate text-sm font-medium">
          {{ title || 'Utfordringens navn' }}
        </p>
        <p v-if="body" class="text-muted line-clamp-2 text-sm">{{ body }}</p>
        <p v-else class="text-dimmed line-clamp-2 text-sm italic">
          Uten tekst sendes ingen varsling.
        </p>
      </div>
    </div>
  </div>
</template>
