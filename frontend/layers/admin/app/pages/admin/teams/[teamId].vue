<script setup lang="ts">
definePageMeta({
  permission: 'teams:view',
  layout: 'admin',
})

/**
 * Legacy redirect: team pages moved under their project.
 *
 * This cannot be a `definePageMeta({ redirect })` — that is synchronous, and the
 * old URL carries no projectId, so the target has to be looked up. `Team.parentProject`
 * exists server-side (gql/teams.graphqls), which is what makes this resolvable
 * at all.
 */
const route = useRoute('admin-teams-teamId')

gql(`
  query LegacyTeamRedirect($id: ID!) {
    team(id: $id) {
      id
      parentProject {
        id
      }
    }
  }
`)

const { isAuthReady } = useAuthReady()
const { data, error, fetching } = useLegacyTeamRedirectQuery({
  variables: computed(() => ({ id: route.params.teamId })),
  pause: computed(() => !isAuthReady.value),
})

watch(
  () => data.value?.team,
  (team) => {
    if (!team) return
    navigateTo(
      {
        name: 'admin-projects-projectId-teams-teamId',
        params: { projectId: team.parentProject.id, teamId: team.id },
      },
      { replace: true },
    )
  },
  { immediate: true },
)
</script>

<template>
  <div>
    <AdminLoadingState v-if="fetching" />
    <UEmpty
      v-else-if="error || !data?.team"
      icon="lucide:users-round"
      title="Fant ikke laget"
      description="Laget finnes ikke, eller du har ikke tilgang til prosjektet det hører til."
      :actions="[{ label: 'Se alle prosjekter', to: '/admin/projects' }]"
    />
  </div>
</template>
