<script setup lang="ts">
import { selectHomeProjects } from '../../utils/homeProjects'

definePageMeta({
  layout: 'admin',
})

/**
 * Split into one query per concern rather than one query for the page.
 *
 * `feedback` is `@requireRole(["admin","superadmin"])` and the directive
 * returns an *error* rather than null (`internal/graph/directives/auth.go`).
 * On a non-null field that fails the whole operation — so while it shared a
 * document with `me` and `projects`, a project admin got an error state for
 * the entire page even though those two resolved fine. Separate documents fail
 * separately, and the gated one is paused for users who cannot read it.
 */
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

/**
 * A project admin holds roles for specific projects, so the global list is
 * filtered to what they may actually open. Superadmins and admins see
 * everything, which is what `canViewProject` already encodes.
 */
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
  <!--
    Capped width, deliberately. The restructure swapped `UContainer` for a plain
    div everywhere so pages could use the whole panel — right for tables and
    card grids, wrong here: this page is a single column of prose-like blocks,
    and stretched across a wide screen the project section became a row of
    seven tiles spread over ~1600px with a feedback list of one-line entries
    under it. `max-w-*` on a deliberate element was always the exception to that
    sweep.
  -->
  <div class="max-w-6xl">
    <h1 v-if="projectData?.me" class="my-8 text-3xl text-balance">
      {{ greeting }}, {{ projectData.me.name }}
    </h1>

    <AdminLoadingState v-if="fetching" />
    <AdminErrorState v-else-if="error" :error />

    <div v-else class="@container space-y-8">
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
  </div>
</template>
