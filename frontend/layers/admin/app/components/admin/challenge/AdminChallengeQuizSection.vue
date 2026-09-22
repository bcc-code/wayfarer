<script setup lang="ts">
const props = defineProps<{
  projectId?: string
  challengeId?: string
  quiz?: {
    id: string
    questions: { id: string }[]
    completionPoints: number
    allowRetakes: boolean
    randomizeQuestions: boolean
    revealCorrectAnswers: boolean
    timeoutSeconds?: number | null
  } | null
}>()

// Labelled pairs, not a row of sentences: "Ett forsøk / Fast rekkefølge" side
// by side reads as loose text rather than as this quiz's settings.
const facts = computed(() => {
  const quiz = props.quiz
  if (!quiz) return []

  return [
    { label: 'Spørsmål', value: formatNumber(quiz.questions.length) },
    {
      label: 'Poeng for fullføring',
      value: formatNumber(quiz.completionPoints),
    },
    { label: 'Forsøk', value: quiz.allowRetakes ? 'Flere tillatt' : 'Ett' },
    {
      label: 'Rekkefølge',
      value: quiz.randomizeQuestions ? 'Tilfeldig' : 'Fast',
    },
    {
      label: 'Riktige svar',
      value: quiz.revealCorrectAnswers ? 'Vises' : 'Skjules',
    },
    {
      label: 'Tidsgrense',
      value: quiz.timeoutSeconds
        ? `${quiz.timeoutSeconds} sek per spørsmål`
        : 'Ingen',
    },
  ]
})
</script>

<template>
  <AdminSection title="Quiz">
    <div v-if="!challengeId || !projectId" class="text-dimmed text-sm">
      Quiz-innstillinger blir tilgjengelig etter at utfordringen er opprettet.
    </div>

    <div v-else class="@container space-y-4">
      <!-- An empty quiz is the state worth flagging: the challenge is live and
           there is nothing to answer. -->
      <p v-if="quiz && !quiz.questions.length" class="text-warning text-sm">
        Quizen har ingen spørsmål ennå.
      </p>

      <dl
        v-if="quiz"
        class="grid gap-x-8 gap-y-3 @md:grid-cols-3 @3xl:grid-cols-6"
      >
        <div v-for="fact in facts" :key="fact.label">
          <dt class="text-muted text-xs">{{ fact.label }}</dt>
          <dd class="text-sm">{{ fact.value }}</dd>
        </div>
      </dl>
      <p v-else class="text-dimmed text-sm">
        Konfigurer quiz-innstillinger og spørsmål.
      </p>

      <UButton
        variant="soft"
        icon="lucide:pencil"
        :to="{
          name: 'admin-projects-projectId-challenges-challengeId-quiz',
          params: { projectId, challengeId },
        }"
      >
        Rediger quiz
      </UButton>
    </div>
  </AdminSection>
</template>
