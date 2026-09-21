<script setup lang="ts">
// `formatDateRange` / `formatNumber` are auto-imported from the shared root;
// only layer-local modules need an explicit relative import.
import {
  describeProjectTiming,
  formatProjectCountdown,
} from '../../../utils/dates'
import { dailyAverage } from '../../../utils/sparkline'
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

/**
 * One query per section, not one for the page.
 *
 * Counting per project needs a `filter: { projectId }` on each root
 * connection, and GraphQL has no dynamic aliasing — so a single page-level
 * document cannot count N projects. A component that owns its own query
 * sidesteps that entirely: each mounted section asks for its own project, and
 * urql keys the document cache on variables so nothing is fetched twice.
 */
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

/**
 * Shortcuts come from the same `PROJECT_NAV` the sidebar uses, so they cannot
 * drift from it and permission gating is already handled and tested there.
 * "Oversikt" is dropped because it resolves to the same route as the section
 * title above it.
 */
const permissions = usePermissions()
const shortcuts = computed(() =>
  visibleNavItems(PROJECT_NAV, permissions, {
    projectId: props.project.id,
  })
    .filter((item) => item.to !== 'admin-projects-projectId')
    .map((item) => ({
      ...item,
      count: countByRoute.value[item.to],
      // `item.to` is a union of every project route name, so typed routes
      // cannot verify that one `params` shape satisfies all of them — even
      // though every member of PROJECT_NAV takes exactly `projectId`. Same
      // cast, for the same reason, as `projects/[projectId]/index.vue`.
      to: {
        name: item.to,
        params: { projectId: props.project.id },
      } as RouteLocationRaw,
    })),
)

const participants = computed(() => data.value?.users.totalCount)

const trend = computed(() => data.value?.project?.activityTrend ?? [])

/**
 * Two tiles, not one chart with two axes: points run to thousands and active
 * users to dozens, and a shared scale would flatten one of them into the
 * baseline. A second y-axis is never the answer.
 *
 * The aggregates differ on purpose. Points are additive, so the window total is
 * meaningful. Active users is a distinct count **per day** — summing it would
 * count the same person once per day they appeared — so it is shown as a daily
 * average.
 */
const trendTiles = computed(() => {
  if (trend.value.length === 0) return []
  return [
    {
      key: 'points',
      label: 'Poeng siste 14 dager',
      value: formatNumber(
        trend.value.reduce((sum, point) => sum + point.points, 0),
      ),
      points: trend.value.map((point) => ({
        date: point.date,
        value: point.points,
      })),
    },
    {
      key: 'activeUsers',
      label: 'Aktive deltakere per dag',
      value: formatNumber(
        dailyAverage(trend.value.map((point) => point.activeUsers)),
      ),
      points: trend.value.map((point) => ({
        date: point.date,
        value: point.activeUsers,
      })),
    },
  ]
})
</script>

<template>
  <!--
    `@container` on this component's own root: the shortcut grid tracks the
    width the section is *given*, not the window's. See Conventions in
    notes/admin-ux-improvements.md.

    No accent ring. Tinting the card border with the project's branding colour
    was meant to tell stacked sections apart, but with one active project there
    is nothing to tell apart and an arbitrary per-project hue on the border just
    reads as random. The logo already carries the project's identity.
  -->
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

      <!--
        Stat tile per measure — label, aggregate, sparkline. Rendered only for a
        running project; `withTrend` skips the field otherwise.
      -->
      <div v-if="trendTiles.length" class="grid gap-3 @2xl:grid-cols-2">
        <div
          v-for="tile in trendTiles"
          :key="tile.key"
          class="bg-elevated/50 rounded-lg p-3"
        >
          <p class="text-muted text-xs">{{ tile.label }}</p>
          <p class="mb-2 text-xl font-semibold">{{ tile.value }}</p>
          <AdminSparkline
            :points="tile.points"
            :label="tile.label"
            :height="36"
            empty-label="Ingen aktivitet siste 14 dager"
          />
        </div>
      </div>

      <!--
        `auto-fit` rather than a fixed column count: the number of shortcuts
        varies from 4 to 7 with the viewer's permissions, and a fixed 6-wide
        grid left the 7th tile alone on a second row. auto-fit collapses the
        empty tracks so whatever is visible stretches to fill one row.
      -->
      <nav
        class="grid grid-cols-[repeat(auto-fit,minmax(10rem,1fr))] gap-2"
        :aria-label="`Snarveier for ${project.name}`"
      >
        <!--
          One line per shortcut — icon, label, count — rather than three stacked
          rows. Denser, and it means entries with no count (Poeng,
          Innstillinger) simply have none instead of reserving empty space
          where a number would go.
        -->
        <NuxtLink
          v-for="item in shortcuts"
          :key="item.label"
          :to="item.to"
          class="bg-elevated/50 hover:bg-elevated focus-visible:ring-primary flex items-center gap-2 rounded-lg px-3 py-2 transition focus-visible:ring-2 focus-visible:outline-none"
        >
          <UIcon :name="item.icon" class="text-muted size-4 shrink-0" />
          <span class="truncate text-xs">{{ item.label }}</span>
          <!--
            Plain counts, deliberately not health signals: challenges are
            episodic, so a low or zero count is an ordinary state rather than
            something to flag.
          -->
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
