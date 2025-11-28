<script setup lang="ts">
const leaderboard: LeaderboardEntry[] = [
  {
    id: '1',
    name: 'Name Nameson',
    score: 100,
    rank: 1,
    description: 'Local Church',
    tags: [LeaderboardEntryTag.TeamLead],
  },
  {
    id: '2',
    name: 'Jane Doe',
    score: 90,
    rank: 2,
    description: 'Global Church',
    tags: [],
  },
  {
    id: '3',
    name: 'Bob Smith',
    score: 80,
    rank: 3,
    description: 'Local Church',
    tags: [LeaderboardEntryTag.Me],
  },
  {
    id: '4',
    name: 'John Doe',
    score: 70,
    rank: 4,
    description: 'Global Church',
    tags: [],
  },
]
</script>

<template>
  <div
    class="border-default bg-background-default aspect-1/2 w-[400px] overflow-clip rounded-xl border text-start"
  >
    <PageLayout title="Preview">
      <div class="gap-list-section-gap flex flex-col">
        <DesignButton>Primary Button</DesignButton>
        <DesignButton variant="secondary">Secondary Button</DesignButton>
        <LeaderboardList
          :leaderboard
          :badge="
            (entry) =>
              entry.tags.includes(LeaderboardEntryTag.TeamLead)
                ? $t('unit.unitLeader')
                : undefined
          "
          class="mt-medium"
        />
        <div
          class="bg-background-indent rounded-list p-default mt-medium flex flex-col gap-1"
        >
          <p class="text-label text-text-default">Default text color</p>
          <p class="text-label text-text-muted">Muted text color</p>
          <p class="text-label text-text-hint">Hint text color</p>
          <p class="text-label text-accent-contrast">Accent contrast color</p>
          <p class="text-label text-on-accent bg-accent">On accent color</p>
        </div>
      </div>
    </PageLayout>
  </div>
</template>
