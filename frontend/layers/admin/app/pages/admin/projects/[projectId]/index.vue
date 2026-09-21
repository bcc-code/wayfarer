<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'

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

// Counts only — each section owns its own list query on its own route.
gql(`
  query AdminProjectOverview($projectId: ID!) {
    challenges(first: 0, filter: { projectId: $projectId }) { totalCount }
    achievements(first: 0, filter: { projectId: $projectId }) { totalCount }
    events(first: 0, filter: { projectId: $projectId }) { totalCount }
    superteams(first: 0, filter: { projectId: $projectId }) { totalCount }
  }
`)

const { isAuthReady } = useAuthReady()
const { data, error, fetching } = useAdminProjectOverviewQuery({
  variables: computed(() => ({ projectId: route.params.projectId })),
  pause: computed(() => !isAuthReady.value),
})

const projectRoute = (name: string): RouteLocationRaw =>
  ({
    name,
    params: { projectId: route.params.projectId },
  }) as RouteLocationRaw

const sections = computed(() => [
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
  <div>
    <AdminQueryState :fetching="fetchingProject" :error>
      <template v-if="project">
        <header class="mb-8 space-y-2">
          <img
            v-if="project.branding.logoImage?.url"
            :src="project.branding.logoImage.url"
            width="64"
            class="mb-4 rounded"
          />
          <h1 class="text-3xl">{{ project.name }}</h1>
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

        <!--
          Container query, not viewport. `lg:grid-cols-4` measured the window, but
          two sidebars take ~600px out of it, so four cards overflowed the panel
          on exactly the widths the breakpoint was meant to cover.
        -->
        <div class="@container">
          <div class="grid gap-4 @md:grid-cols-2 @4xl:grid-cols-4">
            <NuxtLink
              v-for="section in sections"
              :key="section.label"
              :to="projectRoute(section.to)"
            >
              <UCard class="hover:bg-elevated/50 h-full transition-colors">
                <div class="flex items-center gap-3">
                  <UIcon :name="section.icon" class="text-muted size-5" />
                  <span class="font-medium">{{ section.label }}</span>
                </div>
                <div class="mt-2 text-2xl tabular-nums">
                  <USkeleton v-if="fetching" class="h-8 w-12" />
                  <template v-else>{{ section.count ?? 0 }}</template>
                </div>
              </UCard>
            </NuxtLink>
          </div>
        </div>
      </template>
    </AdminQueryState>
  </div>
</template>
