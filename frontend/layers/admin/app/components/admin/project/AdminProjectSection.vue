<script setup lang="ts">
import {
  describeProjectTiming,
  formatProjectCountdown,
} from '../../../utils/dates'
import type { RouteLocationRaw } from 'vue-router'
import {
  PROJECT_NAV,
  visibleNavItems,
  type AdminRouteName,
} from '../../../utils/adminNav'

interface SectionProject {
  id: string
  name: string
  startDate: string
  endDate: string
  branding: {
    logoImage?: { url: string } | null
  }
}

const props = defineProps<{
  project: SectionProject
}>()

// One query per section: counting N projects in one document would need
// dynamic aliasing. urql keys on variables, so nothing is fetched twice.
gql(`
  query AdminProjectSectionCounts($projectId: ID!, $withTrend: Boolean!) {
    project(id: $projectId) {
      id
      # Skipped entirely for a project that has not started: its trend is 14
      # empty days, which reads as a broken chart rather than as "not yet".
      activityTrend(days: 14) @include(if: $withTrend) {
        date
        points
        activeUsers
      }
    }
    challenges(first: 0, filter: { projectId: $projectId }) {
      totalCount
    }
    achievements(first: 0, filter: { projectId: $projectId }) {
      totalCount
    }
    # Aliased so the shortcut count and the badge row can ask for different
    # pages of the same field. Skipped with the trend: between camps the home
    # page should stay cheap.
    achievementBadges: achievements(
      first: 50
      filter: { projectId: $projectId }
    ) @include(if: $withTrend) {
      edges {
        node {
          id
          name
          awardedUserCount
          imageCompletedObject {
            ...ImageFields
          }
        }
      }
    }
    events(first: 0, filter: { projectId: $projectId }) {
      totalCount
    }
    superteams(first: 0, filter: { projectId: $projectId }) {
      totalCount
    }
    teams(first: 0, filter: { projectId: $projectId }) {
      totalCount
    }
    users(first: 0, filter: { projectId: $projectId }) {
      totalCount
    }
  }
`)

const timing = computed(() =>
  describeProjectTiming(props.project.startDate, props.project.endDate),
)

const { isAuthReady } = useAuthReady()
const { data, fetching } = useAdminProjectSectionCountsQuery({
  variables: computed(() => ({
    projectId: props.project.id,
    withTrend: timing.value.state === 'running',
  })),
  pause: computed(() => !isAuthReady.value),
})

const countdown = computed(() => formatProjectCountdown(timing.value))

const projectRoute = computed(() => ({
  name: 'admin-projects-projectId' as const,
  params: { projectId: props.project.id },
}))

/** Counts keyed by the route the shortcut points at. */
const countByRoute = computed<Partial<Record<AdminRouteName, number>>>(() => {
  const counts = data.value
  if (!counts) return {}
  return {
    'admin-projects-projectId-challenges': counts.challenges.totalCount,
    'admin-projects-projectId-achievements': counts.achievements.totalCount,
    'admin-projects-projectId-events': counts.events.totalCount,
    'admin-projects-projectId-superteams': counts.superteams.totalCount,
    'admin-projects-projectId-teams': counts.teams.totalCount,
  }
})

// Same `PROJECT_NAV` as the sidebar, so they cannot drift and permission
// gating comes with it. "Oversikt" is the section title's own route.
const permissions = usePermissions()
const shortcuts = computed(() =>
  visibleNavItems(PROJECT_NAV, permissions, {
    projectId: props.project.id,
  })
    .filter((item) => item.to !== 'admin-projects-projectId')
    .map((item) => ({
      ...item,
      count: countByRoute.value[item.to],
      // `item.to` is a union of route names, so typed routes cannot check one
      // `params` shape against all of them.
      to: {
        name: item.to,
        params: { projectId: props.project.id },
      } as RouteLocationRaw,
    })),
)

const participants = computed(() => data.value?.users.totalCount)

const trend = computed(() => data.value?.project?.activityTrend ?? [])

const achievementBadges = computed(
  () => data.value?.achievementBadges?.edges.map((edge) => edge.node) ?? [],
)
</script>

<template>
  <!-- `@container`: the shortcut grid tracks the width this section is given. -->
  <UCard class="@container">
    <div class="space-y-4">
      <div class="flex items-center gap-3">
        <img
          v-if="project.branding.logoImage?.url"
          :src="project.branding.logoImage.url"
          height="40"
          width="40"
          class="size-10 shrink-0 rounded object-cover"
          alt=""
        />
        <div class="min-w-0 grow">
          <NuxtLink
            :to="projectRoute"
            class="text-lg font-semibold hover:underline"
          >
            {{ project.name }}
          </NuxtLink>
          <p class="text-muted text-xs">
            {{ formatDateRange(project.startDate, project.endDate) }}
            <template v-if="participants !== undefined">
              &middot; {{ formatNumber(participants) }} deltakere
            </template>
          </p>
        </div>
        <UBadge
          :color="timing.state === 'running' ? 'success' : 'neutral'"
          variant="subtle"
          class="shrink-0"
        >
          {{ countdown }}
        </UBadge>
      </div>

      <!-- Both only for a running project; `withTrend` skips the fields. -->
      <AdminActivityTrend :trend="trend" :days="14" />

      <AdminAchievementBadges
        v-if="achievementBadges.length"
        :achievements="achievementBadges"
        :project-id="project.id"
        :participants="participants"
        :size="36"
      />

      <!-- `auto-fit`: 4-7 shortcuts depending on permissions, so no fixed
           column count fits without orphaning a tile. -->
      <nav
        class="grid grid-cols-[repeat(auto-fit,minmax(10rem,1fr))] gap-2"
        :aria-label="`Snarveier for ${project.name}`"
      >
        <NuxtLink
          v-for="item in shortcuts"
          :key="item.label"
          :to="item.to"
          class="bg-elevated/50 hover:bg-elevated focus-visible:ring-primary flex items-center gap-2 rounded-lg px-3 py-2 transition focus-visible:ring-2 focus-visible:outline-none"
        >
          <UIcon :name="item.icon" class="text-muted size-4 shrink-0" />
          <span class="truncate text-xs">{{ item.label }}</span>
          <!-- Plain counts, not health signals: challenges are episodic, so a
               zero is ordinary. -->
          <USkeleton
            v-if="fetching && item.count === undefined"
            class="ml-auto h-4 w-6 shrink-0"
          />
          <span
            v-else-if="item.count !== undefined"
            class="ml-auto shrink-0 text-sm font-bold"
          >
            {{ formatNumber(item.count) }}
          </span>
        </NuxtLink>
      </nav>
    </div>
  </UCard>
</template>
