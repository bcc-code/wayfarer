<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'
import {
  csvFilename,
  formatDecimal,
  formatPercent,
  questionResultTables,
  toCsv,
} from '../../../../../../utils/quizResultsExport'

definePageMeta({
  permission: 'projects:view',
  layout: 'admin',
})

gql(`
  query AdminChallengeResultsPage($challengeId: ID!) {
    challenge(id: $challengeId) {
      __typename
      id
      name
      project {
        id
        name
      }
      ... on QuizChallenge {
        quiz {
          id
          name
        }
      }
    }
  }
`)

gql(`
  query AdminQuizResults($quizId: ID!) {
    quizResults(quizId: $quizId) {
      submissionCount
      participantCount
      sessionCount
      averageScore
      averageMaxScore
      averageScorePercentage
      questions {
        __typename
        question {
          id
          questionText
          questionOrder
        }
        responseCount
        ... on PredefinedQuestionResults {
          correctCount
          options {
            answer {
              id
              answerText
            }
            count
            percentage
            isCorrect
          }
        }
        ... on NumberQuestionResults {
          average
          median
          min
          max
          buckets {
            from
            to
            count
            percentage
          }
        }
        ... on FreeTextQuestionResults {
          distinctCount
          groups {
            text
            count
            percentage
          }
          responses
        }
        ... on OrderingQuestionResults {
          fullyCorrectCount
          items {
            item {
              id
              itemText
            }
            correctPosition
            correctlyPlacedCount
            percentage
          }
        }
      }
    }
  }
`)

const route = useRoute(
  'admin-projects-projectId-challenges-challengeId-results',
)

const { isAuthReady } = useAuthReady()
const { data: challengeData } = useAdminChallengeResultsPageQuery({
  variables: computed(() => ({ challengeId: route.params.challengeId })),
  pause: computed(() => !isAuthReady.value),
})

// Trailing breadcrumb crumbs; the path above them is derived from the route.
useAdminPage(() => [
  {
    label: challengeData.value?.challenge.name ?? 'Utfordring',
    to: {
      name: 'admin-projects-projectId-challenges-challengeId',
      params: {
        projectId: route.params.projectId,
        challengeId: route.params.challengeId,
      },
    } as RouteLocationRaw,
  },
  'Resultater',
])

const quiz = computed(() =>
  challengeData.value?.challenge.__typename === 'QuizChallenge'
    ? challengeData.value.challenge.quiz
    : undefined,
)

const { data, fetching, error } = useAdminQuizResultsQuery({
  variables: computed(() => ({ quizId: quiz.value?.id ?? '' })),
  pause: computed(() => !isAuthReady.value || !quiz.value?.id),
  // Results move while a session is running, so never serve a stale copy
  // without going back to the server for the current one.
  requestPolicy: 'cache-and-network',
})

const results = computed(() => data.value?.quizResults)
const questions = computed(() => results.value?.questions ?? [])

/**
 * The header facts. Participants is shown next to submissions rather than
 * instead of it: pooling every session means someone who took the quiz twice
 * counts twice, and showing both makes that visible instead of hiding it.
 */
const facts = computed(() => {
  const r = results.value
  if (!r) return []

  const list = [
    { label: 'Besvarelser', value: formatNumber(r.submissionCount) },
    { label: 'Deltakere', value: formatNumber(r.participantCount) },
    { label: 'Sesjoner', value: formatNumber(r.sessionCount) },
  ]

  // Null until something has actually been scored — better an absent fact than
  // a confident "0 %".
  if (r.averageScore != null && r.averageMaxScore != null) {
    list.push({
      label: 'Snittscore',
      value: `${formatDecimal(r.averageScore)} / ${formatDecimal(r.averageMaxScore)}`,
    })
  }
  if (r.averageScorePercentage != null) {
    list.push({
      label: 'Snitt i prosent',
      value: formatPercent(r.averageScorePercentage),
    })
  }

  return list
})

const hasResults = computed(() => (results.value?.submissionCount ?? 0) > 0)

/**
 * Builds the CSV in the browser from the query result already on the page —
 * there is no second endpoint, and nothing leaves the panel that isn't already
 * on screen.
 */
function downloadCsv() {
  const csv = toCsv(questionResultTables(questions.value))
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)

  const link = document.createElement('a')
  link.href = url
  link.download = csvFilename(quiz.value?.name ?? 'quiz')
  link.click()

  URL.revokeObjectURL(url)
}
</script>

<template>
  <!-- Capped like the other detail pages: results read as a column, and a
       full-width row strands each question's bars far from its text. -->
  <div class="max-w-4xl">
    <AdminQueryState :fetching :error>
      <template v-if="results">
        <header class="mb-8">
          <div class="flex flex-wrap items-start justify-between gap-4">
            <div>
              <h1 class="text-3xl">Resultater</h1>
              <p v-if="quiz" class="text-muted mt-1">{{ quiz.name }}</p>
            </div>
            <UButton
              v-if="hasResults"
              variant="soft"
              icon="lucide:download"
              @click="downloadCsv"
            >
              Last ned CSV
            </UButton>
          </div>

          <dl v-if="hasResults" class="mt-6 flex flex-wrap gap-x-10 gap-y-4">
            <div v-for="fact in facts" :key="fact.label">
              <dt class="text-muted text-xs">{{ fact.label }}</dt>
              <dd class="text-2xl tabular-nums">{{ fact.value }}</dd>
            </div>
          </dl>
        </header>

        <UCard v-if="!hasResults">
          <div class="py-6 text-center">
            <UIcon
              name="lucide:chart-no-axes-column"
              class="text-dimmed mb-2 size-6"
            />
            <p class="text-muted text-sm">Ingen besvarelser ennå</p>
            <p class="text-dimmed mt-1 text-xs">
              Resultatene fyller seg ut etter hvert som deltakerne fullfører
              quizen.
            </p>
          </div>
        </UCard>

        <ul v-else class="space-y-4">
          <li v-for="result in questions" :key="result.question.id">
            <AdminQuizResultsQuestion :result />
          </li>
        </ul>
      </template>
    </AdminQueryState>
  </div>
</template>
