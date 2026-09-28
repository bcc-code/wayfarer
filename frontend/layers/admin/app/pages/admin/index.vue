<script setup lang="ts">
import { selectHomeProjects } from '../../utils/homeProjects'

definePageMeta({
  layout: 'admin',
})

// One query per concern: `feedback` is admin/superadmin-only and the directive
// errors rather than returning null, which on a non-null field fails the whole
// document. Separate documents fail separately.
gql(`
  query AdminHomeProjects($now: DateTime!) {
    me {
      id
      name
    }
    projects(filter: { endDateAfter: $now, archived: false }, first: 100) {
      edges {
        node {
          id
          name
          description
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

gql(`
  query AdminHomeFeedback {
    feedback(first: 5, filter: { handled: false }) {
      totalCount
      edges {
        node {
          id
          message
          createdAt
          tags
          user {
            id
            name
          }
        }
      }
    }
  }
`)

const { isAuthReady } = useAuthReady()
const { canAccessFeedback, canViewProject } = usePermissions()

const {
  data: projectData,
  fetching,
  error,
} = useAdminHomeProjectsQuery({
  variables: { now: new Date().toISOString() },
  pause: computed(() => !isAuthReady.value),
})

const { data: feedbackData } = useAdminHomeFeedbackQuery({
  pause: computed(() => !isAuthReady.value || !canAccessFeedback.value),
})

// A project admin only sees projects they may open.
const visibleProjects = computed(() =>
  (projectData.value?.projects.edges ?? [])
    .map((edge) => edge.node)
    .filter((project) => canViewProject(project.id)),
)

const selection = computed(() => selectHomeProjects(visibleProjects.value))

const unhandledFeedback = computed(
  () => feedbackData.value?.feedback.edges.map((edge) => edge.node) ?? [],
)

const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 12) return 'God morgen'
  if (hour < 18) return 'God ettermiddag'
  return 'God kveld'
})
</script>

<template>
  <!-- Capped: a single column of blocks, not a table or a card grid. -->
  <div class="max-w-6xl">
    <h1 v-if="projectData?.me" class="my-8 text-3xl text-balance">
      {{ greeting }}, {{ projectData.me.name }}
    </h1>

    <AdminQueryState :fetching :error>
      <div class="@container space-y-8">
        <!--
          One section per active project, stacked. Usually there is exactly one;
          the list is ordered by ending soonest and capped, with the remainder
          behind the "flere aktive" link.
        -->
        <section v-if="selection.active.length">
          <div class="mb-3 flex items-baseline gap-4">
            <h2>
              {{
                selection.active.length > 1
                  ? 'Aktive prosjekter'
                  : 'Aktivt prosjekt'
              }}
            </h2>
            <UButton variant="soft" size="xs" :to="{ name: 'admin-projects' }">
              {{
                selection.hiddenActiveCount
                  ? `${selection.hiddenActiveCount} flere aktive`
                  : 'Se alle'
              }}
            </UButton>
          </div>
          <div class="space-y-4">
            <AdminProjectSection
              v-for="project in selection.active"
              :key="project.id"
              :project
            />
          </div>
        </section>

        <!--
          Nothing running: the next one, in the same section. Its counts double as
          the readiness view — challenges, achievements and teams at zero is
          exactly what "not set up yet" looks like — so there is no separate
          readiness block or query.
        -->
        <section v-else-if="selection.next">
          <h2 class="mb-3">Neste prosjekt</h2>
          <AdminProjectSection :project="selection.next" />
        </section>

        <UEmpty
          v-else
          icon="lucide:square-dashed-mouse-pointer"
          title="Ingen aktive eller kommende prosjekter"
          description="Alle prosjekter er avsluttet eller arkivert."
          :actions="[{ label: 'Se alle prosjekter', to: '/admin/projects' }]"
        />

        <section v-if="feedbackData?.feedback">
          <div class="mb-3 flex items-baseline gap-4">
            <h2>Ubehandlede tilbakemeldinger</h2>
            <UBadge v-if="feedbackData.feedback.totalCount" variant="subtle">
              {{ feedbackData.feedback.totalCount }}
            </UBadge>
            <UButton variant="soft" size="xs" :to="{ name: 'admin-feedback' }">
              Se alle
            </UButton>
          </div>
          <AdminRecentActivity :feedback-entries="unhandledFeedback" />
        </section>
      </div>
    </AdminQueryState>
  </div>
</template>
