<script setup lang="ts">
defineProps<{
  challenge: {
    name?: string
    description?: string
    image?: string
    buttonText?: string
  }
}>()
</script>

<template>
  <div class="max-w-md">
    <div class="shadow-large rounded-card overflow-clip">
      <div
        v-if="challenge.image"
        class="bg-accent aspect-[1.25] w-full overflow-hidden"
      >
        <img :src="challenge.image" class="size-full object-cover" />
      </div>
      <div
        v-else
        class="bg-accent aspect-[1.25] w-full flex items-center justify-center text-on-accent/50"
      >
        No image
      </div>
      <div class="bg-background-raised p-default gap-default space-y-default">
        <div class="space-y-small">
          <h3 class="text-heading">
            {{ challenge.name || 'Challenge Name' }}
          </h3>
          <div
            v-if="challenge.description"
            class="text-label"
            v-html="challenge.description"
          />
          <div v-else class="text-label text-text-muted">No description</div>
        </div>
        <div class="mt-auto grid">
          <DesignButton size="large">
            {{ challenge.buttonText || 'Button Text' }}
          </DesignButton>
        </div>
      </div>
    </div>
  </div>
</template>
