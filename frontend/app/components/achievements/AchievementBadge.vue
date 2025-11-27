<script setup lang="ts">
defineProps<{
  achievement: Partial<Achievement>
}>()

const open = ref(false)
</script>

<template>
  <div>
    <button
      v-if="!achievement.hidden"
      class="grid aspect-square place-items-center overflow-hidden rounded-full"
      @click="open = true"
    >
      <NuxtImg
        v-if="achievement.image && achievement.achievedAt != null"
        :src="achievement.image"
        class="size-full object-cover"
      />
      <NuxtImg
        v-else
        src="/images/achievement-placeholder.png"
        class="size-full object-cover"
      />
    </button>
    <UModal
      v-model:open="open"
      :ui="{ content: 'bg-background-default' }"
      :transition="false"
      modal
      fullscreen
    >
      <template #content="{ close }">
        <PageLayout
          :title="achievement.achievedAt ? 'Unlocked!' : 'Achievement'"
        >
          <template #action>
            <DesignIconButton icon="lucide:x" @click="close" />
          </template>

          <div class="flex h-full flex-col items-center justify-center gap-6">
            <div
              :class="[
                'grid aspect-square size-55 place-items-center overflow-hidden rounded-full',
                { 'shadow-large': achievement.achievedAt },
              ]"
            >
              <NuxtImg
                v-if="achievement.image && achievement.achievedAt != null"
                :src="achievement.image"
                class="size-full object-cover"
              />
              <NuxtImg
                v-else
                src="/images/achievement-placeholder.png"
                class="size-full object-cover"
              />
            </div>
            <div
              class="flex flex-col items-center gap-1 text-center text-balance"
            >
              <h3 class="text-heading">
                {{ achievement.name }}
              </h3>
              <p class="text-label">
                {{ achievement.description }}
              </p>
            </div>
          </div>
        </PageLayout>
      </template>
    </UModal>
  </div>
</template>
