<script setup lang="ts">
const leaderboard: LeaderboardEntry[] = [
  {
    id: '1',
    name: 'Ola Nordmann',
    score: 100,
    rank: 1,
    description: 'Lokal menighet',
    tags: [LeaderboardEntryTag.TeamLead],
  },
  {
    id: '2',
    name: 'Kari Hansen',
    score: 90,
    rank: 2,
    description: 'Global menighet',
    tags: [],
  },
  {
    id: '3',
    name: 'Per Olsen',
    score: 80,
    rank: 3,
    description: 'Lokal menighet',
    tags: [LeaderboardEntryTag.Me],
  },
  {
    id: '4',
    name: 'Lisa Andersen',
    score: 70,
    rank: 4,
    description: 'Global menighet',
    tags: [],
  },
]
</script>

<template>
  <div
    class="border-default bg-background-default aspect-1/2 w-[400px] overflow-clip rounded-xl border text-start"
  >
    <PageLayout title="Forhåndsvisning">
      <div class="gap-list-section-gap flex flex-col p-list-outside">
        <DesignButton>Primærknapp</DesignButton>
        <DesignButton variant="secondary">Sekundærknapp</DesignButton>
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
          <p class="text-label text-text-default">Standard tekstfarge</p>
          <p class="text-label text-text-muted">Dempet tekstfarge</p>
          <p class="text-label text-text-hint">Hint tekstfarge</p>
          <p class="text-label text-accent-contrast">Aksentkontrastfarge</p>
          <p class="text-label text-on-accent bg-accent">På aksentfarge</p>
        </div>
      </div>
    </PageLayout>
  </div>
</template>
