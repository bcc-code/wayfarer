<script setup lang="ts">
const props = defineProps<{
  stats: {
    totalUsers: number
    totalPointsAwarded: number
    newUsersLast7Days: number
  }
}>()

const cards = computed(() => [
  {
    label: 'Total Users',
    value: formatNumber(props.stats.totalUsers),
    icon: 'i-lucide-users',
    subtitle: `+${props.stats.newUsersLast7Days} this week`,
  },
  {
    label: 'Total Points Awarded',
    value: formatNumber(props.stats.totalPointsAwarded),
    icon: 'i-lucide-star',
  },
])
</script>

<template>
  <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
    <UCard v-for="card in cards" :key="card.label">
      <div class="flex flex-col items-start gap-3">
        <div class="rounded-lg p-2 aspect-square bg-muted flex">
          <UIcon :name="card.icon" class="size-4" />
        </div>
        <div>
          <p class="text-muted text-sm">{{ card.label }}</p>
          <p class="text-2xl font-bold">{{ card.value }}</p>
          <p v-if="card.subtitle" class="text-dimmed text-xs">
            {{ card.subtitle }}
          </p>
        </div>
      </div>
    </UCard>
  </div>
</template>
