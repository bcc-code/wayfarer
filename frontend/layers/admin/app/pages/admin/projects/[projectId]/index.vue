<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'
import {
  describeProjectTiming,
  formatProjectCountdown,
} from '../../../../utils/dates'

definePageMeta({
  permission: 'projects:view',
  layout: 'admin',
})

const route = useRoute('admin-projects-projectId')
const { canEditProject } = usePermissions()
const canEdit = computed(() => canEditProject(route.params.projectId))

// The project itself comes from the parent route's composable, so this page
// does not re-fetch what the shell already has.
const { project, fetching: fetchingProject } = useCurrentProject()

// Counts for the section links, plus what the page needs to report status.
// Each section still owns its own list query on its own route.
gql(`
  query AdminProjectOverview($projectId: ID!, $withTrend: Boolean!) {
    project(id: $projectId) {
      id
      # A finished or unstarted project's last 14 days are empty, which reads
      # as a broken chart rather than as "nothing is running".
      activityTrend(days: 14) @include(if: $withTrend) {
        date
        points
        activeUsers
      }
    }
    users(first: 0, filter: { projectId: $projectId }) { totalCount }
    teams(first: 0, filter: { projectId: $projectId }) { totalCount }
    challenges(first: 0, filter: { projectId: $projectId }) { totalCount }
    achievements(first: 0, filter: { projectId: $projectId }) { totalCount }
    events(first: 0, filter: { projectId: $projectId }) { totalCount }
    superteams(first: 0, filter: { projectId: $projectId }) { totalCount }
  }
`)

const timing = computed(() =>
  project.value
    ? describeProjectTiming(project.value.startDate, project.value.endDate)
    : undefined,
)

const { isAuthReady } = useAuthReady()
const { data, error, fetching } = useAdminProjectOverviewQuery({
  variables: computed(() => ({
    projectId: route.params.projectId,
    withTrend: timing.value?.state === 'running',
  })),
  pause: computed(() => !isAuthReady.value),
})

const trend = computed(() => data.value?.project?.activityTrend ?? [])
const participants = computed(() => data.value?.users.totalCount)
const projectRoute = (name: string): RouteLocationRaw =>
  ({
    name,
    params: { projectId: route.params.projectId },
  }) as RouteLocationRaw

const sections = computed(() => [
  {
    label: 'Lag',
    icon: 'lucide:users-round',
    count: data.value?.teams.totalCount,
    to: 'admin-projects-projectId-teams',
  },
  {
    label: 'Utfordringer',
    icon: 'lucide:swords',
    count: data.value?.challenges.totalCount,
    to: 'admin-projects-projectId-challenges',
  },
  {
    label: 'Utmerkelser',
    icon: 'lucide:award',
    count: data.value?.achievements.totalCount,
    to: 'admin-projects-projectId-achievements',
  },
  {
    label: 'Arrangement',
    icon: 'lucide:calendar',
    count: data.value?.events.totalCount,
    to: 'admin-projects-projectId-events',
  },
  {
    label: 'Superlag',
    icon: 'lucide:users',
    count: data.value?.superteams.totalCount,
    to: 'admin-projects-projectId-superteams',
  },
])

// `?tab=` links are in the wild: the old tab state was written to history.
// Redirect them to the route that replaced each tab for one release.
const TAB_ROUTES: Record<string, string> = {
  achievements: 'admin-projects-projectId-achievements',
  challenges: 'admin-projects-projectId-challenges',
  superteams: 'admin-projects-projectId-superteams',
}

onMounted(() => {
  const tab = route.query.tab
  const target = typeof tab === 'string' ? TAB_ROUTES[tab] : undefined
  if (target) {
    navigateTo(projectRoute(target), { replace: true })
  }
})
</script>

<template>
  <!-- Capped like the other detail pages: a full-width row of five tiles
       strands each count far from its label. -->
  <div class="max-w-6xl">
    <AdminQueryState :fetching="fetchingProject" :error>
      <template v-if="project">
        <header class="mb-8 space-y-2">
          <img
            v-if="project.branding.logoImage?.url"
            :src="project.branding.logoImage.url"
            width="64"
            class="mb-4 rounded"
            alt=""
          />
          <div class="flex flex-wrap items-center gap-3">
            <h1 class="text-3xl">{{ project.name }}</h1>
            <UBadge
              v-if="timing"
              :color="timing.state === 'running' ? 'success' : 'neutral'"
              variant="subtle"
            >
              {{ formatProjectCountdown(timing) }}
            </UBadge>
          </div>
          <p v-if="project.description" class="text-muted max-w-2xl">
            {{ project.description }}
          </p>
          <p class="text-dimmed text-sm">
            {{ formatDateRange(project.startDate, project.endDate) }}
          </p>
          <div v-if="canEdit" class="pt-2">
            <UButton
              variant="soft"
              icon="lucide:pencil"
              :to="{
                name: 'admin-projects-projectId-edit',
                params: { projectId: route.params.projectId },
              }"
            >
              Rediger prosjekt
            </UButton>
          </div>
        </header>

        <div class="@container space-y-8">
          <AdminSection title="Status">
            <div class="space-y-4">
              <div class="flex flex-wrap items-baseline gap-x-8 gap-y-2">
                <div>
                  <p class="text-muted text-xs">Deltakere</p>
                  <p class="text-2xl font-semibold tabular-nums">
                    <USkeleton v-if="fetching" class="h-8 w-16" />
                    <template v-else>
                      {{ formatNumber(participants ?? 0) }}
                    </template>
                  </p>
                </div>
              </div>

              <AdminActivityTrend :trend="trend" :days="14" />
              <p
                v-if="timing && timing.state !== 'running'"
                class="text-dimmed text-sm"
              >
                {{
                  timing.state === 'upcoming'
                    ? 'Prosjektet har ikke startet — ingen aktivitet ennå.'
                    : 'Prosjektet er avsluttet.'
                }}
              </p>
            </div>
          </AdminSection>

          <AdminSection title="Innhold">
            <!-- Container query, not viewport: two sidebars take ~600px out of
                 the window before the panel gets any. -->
            <div class="grid gap-3 @md:grid-cols-2 @4xl:grid-cols-5">
              <NuxtLink
                v-for="section in sections"
                :key="section.label"
                :to="projectRoute(section.to)"
                class="hover:bg-elevated bg-elevated/50 rounded-lg p-3 transition-colors"
              >
                <div class="flex items-center gap-2">
                  <UIcon :name="section.icon" class="text-muted size-4" />
                  <span class="truncate text-sm">{{ section.label }}</span>
                </div>
                <div class="mt-1 text-xl font-semibold tabular-nums">
                  <USkeleton v-if="fetching" class="h-6 w-10" />
                  <template v-else>
                    {{ formatNumber(section.count ?? 0) }}
                  </template>
                </div>
              </NuxtLink>
            </div>
          </AdminSection>
        </div>
      </template>
    </AdminQueryState>
  </div>
</template>
