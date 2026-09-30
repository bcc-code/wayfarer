<script setup lang="ts">
import type { AvatarProps, DropdownMenuItem } from '@nuxt/ui'
import type { RouteLocationRaw } from 'vue-router'

// `projectId` is what the trailing check matches on; DropdownMenuItem has no
// field for it.
type ProjectMenuItem = DropdownMenuItem & { projectId?: string }

defineProps<{ collapsed?: boolean }>()

// Lighter than AdminProjectsPage: no branding colours needed here.
gql(`
  query AdminProjectSwitcher {
    projects(first: 100, filter: { archived: false }) {
      edges {
        node {
          id
          name
          startDate
          endDate
          branding {
            logoImage {
              url
            }
          }
        }
      }
    }
  }
`)

const { isAuthReady } = useAuthReady()
const { data } = useAdminProjectSwitcherQuery({
  pause: computed(() => !isAuthReady.value),
  // The sidebar is mounted on every admin route; without this the switcher
  // refetches the whole project list on each navigation.
  requestPolicy: 'cache-first',
})

const route = useRoute()
const { canViewProject } = usePermissions()

const projects = computed(
  () =>
    data.value?.projects.edges
      .map((edge) => edge.node)
      .filter((project) => canViewProject(project.id)) ?? [],
)

const { currentProjects, futureProjects, pastProjects } = useGroupedProjects(
  () => projects.value,
)

// `route.params` is a union across every route under typedPages, so the key
// has to be probed rather than read directly.
const projectId = computed(() =>
  'projectId' in route.params ? route.params.projectId : undefined,
)

const activeProject = computed(() =>
  projects.value.find((project) => project.id === projectId.value),
)

type SwitcherProject = (typeof projects.value)[number]

// The icon fallback keeps rows aligned when a project has no logo.
function projectAvatar(project: SwitcherProject): AvatarProps {
  return {
    src: project.branding.logoImage?.url ?? undefined,
    alt: project.name,
    icon: 'lucide:layers',
    ui: { root: 'rounded-md', image: 'object-contain' },
  }
}

/**
 * Switching keeps you in the same *section* of the new project where that makes
 * sense. A leaf route carries an id (`challengeId`, `superTeamId`) that has no
 * counterpart in the target project, so those fall back to the overview.
 */
function switchTo(targetProjectId: string) {
  const currentName = String(route.name ?? '')
  const section = PROJECT_NAV.find(
    (item) => item.match && isNavItemActive(item, currentName),
  )
  // Every PROJECT_NAV target takes exactly `projectId`, but the name is only
  // known at runtime so TypeScript cannot correlate the two.
  return navigateTo({
    name: section?.to ?? 'admin-projects-projectId',
    params: { projectId: targetProjectId },
  } as RouteLocationRaw)
}

function groupToItems(group: SwitcherProject[]): ProjectMenuItem[] {
  return group.map((project) => ({
    label: project.name,
    projectId: project.id,
    avatar: projectAvatar(project),
    onSelect: () => switchTo(project.id),
  }))
}

const items = computed<ProjectMenuItem[][]>(() => {
  const groups: ProjectMenuItem[][] = []

  for (const [label, group] of [
    ['Aktive', currentProjects.value],
    ['Kommende', futureProjects.value],
    ['Tidligere', pastProjects.value],
  ] as const) {
    if (!group.length) continue
    groups.push([{ label, type: 'label' }, ...groupToItems(group)])
  }

  groups.push([
    {
      label: 'Alle prosjekter',
      icon: 'lucide:layers',
      to: { name: 'admin-projects' },
    },
  ])

  return groups
})
</script>

<template>
  <UDropdownMenu
    :items="items"
    :content="{ align: 'center', collisionPadding: 12 }"
    :ui="{
      content: collapsed ? 'w-60' : 'w-(--reka-dropdown-menu-trigger-width)',
    }"
  >
    <template #item-trailing="{ item }">
      <UIcon
        v-if="item.projectId && item.projectId === projectId"
        name="lucide:check"
        class="text-dimmed size-5 shrink-0"
      />
    </template>

    <UButton
      v-bind="{
        label: collapsed ? undefined : (activeProject?.name ?? 'Velg prosjekt'),
        trailingIcon: collapsed ? undefined : 'lucide:chevrons-up-down',
        ...(activeProject
          ? { avatar: projectAvatar(activeProject) }
          : { icon: 'lucide:layers' }),
      }"
      color="neutral"
      variant="ghost"
      block
      :square="collapsed"
      class="data-[state=open]:bg-elevated"
      :ui="{ trailingIcon: 'text-dimmed', label: 'truncate' }"
    />
  </UDropdownMenu>
</template>
