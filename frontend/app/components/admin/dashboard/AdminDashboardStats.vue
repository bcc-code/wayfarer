<script setup lang="ts">
const props = defineProps<{
  stats: {
    totalUsers: number
    totalProjects: number
    totalChallenges: number
    totalPointsAwarded: number
    newUsersLast7Days: number
    activeProjectsCount: number
  }
}>()

const cards = computed(() => [
  {
    label: 'Total Users',
    value: formatNumber(props.stats.totalUsers),
    icon: 'i-lucide-users',
    subtitle: `+${props.stats.newUsersLast7Days} this week`,
    color: 'primary',
  },
  {
    label: 'Active Projects',
    value: props.stats.activeProjectsCount,
    icon: 'i-lucide-folder-open',
    subtitle: `${props.stats.totalProjects} total`,
    color: 'success',
  },
  {
    label: 'Challenges',
    value: formatNumber(props.stats.totalChallenges),
    icon: 'i-lucide-target',
    color: 'warning',
  },
  {
    label: 'Points Awarded',
    value: formatNumber(props.stats.totalPointsAwarded),
    icon: 'i-lucide-star',
    color: 'info',
  },
])
</script>

<template>
  <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
    <UCard v-for="card in cards" :key="card.label">
      <div class="flex items-center gap-3">
        <div
          :class="[
            'rounded-lg p-2',
            `bg-${card.color}-500/10`,
          ]"
        >
          <UIcon :name="card.icon" :class="`size-5 text-${card.color}-500`" />
        </div>
        <div>
          <p class="text-muted text-sm">{{ card.label }}</p>
          <p class="text-2xl font-bold">{{ card.value }}</p>
          <p v-if="card.subtitle" class="text-muted text-xs">
            {{ card.subtitle }}
          </p>
        </div>
      </div>
    </UCard>
  </div>
</template>
