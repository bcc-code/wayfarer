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
    <UDrawer
      v-model:open="open"
      :ui="{ overlay: 'bg-black/50' }"
      modal
      class="rounded-t-card! min-h-1/2"
    >
      <template #body>
        <div class="flex flex-col items-center gap-4 pt-4">
          <div
            class="grid aspect-square size-40 place-items-center overflow-hidden rounded-full"
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
          <div class="flex flex-col items-center text-center">
            <h3 class="text-label">
              {{ achievement.name }}
            </h3>
            <p class="text-caption text-muted">{{ achievement.description }}</p>
            <DesignBadge icon="lucide:plus" class="mt-4">
              {{ achievement.points }}
            </DesignBadge>
          </div>
        </div>
      </template>
    </UDrawer>
  </div>
</template>
