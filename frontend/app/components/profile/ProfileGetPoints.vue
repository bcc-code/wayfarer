<script setup lang="ts">
gql(`
  query ProjectRules {
    myCurrentProject {
      rules {
        markdown
        html
      }
    }
  }
`)

const { track } = useAnalytics()

const open = ref(false)

watch(open, (isOpen) => {
  if (isOpen) {
    track(AnalyticsEvent.HowToGetPointsOpened)
  }
})

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = useProjectRulesQuery({
  pause: computed(() => !isAuthReady.value),
})
</script>

<template>
  <DesignDrawer v-model:open="open" :title="$t('pages.rules')">
    <slot />
    <template #content>
      <LoadingState v-if="fetching" />
      <ErrorState v-else-if="error" :error />
      <div
        v-else-if="data?.myCurrentProject.rules"
        id="project-rules"
        class="p-medium"
        v-html="data.myCurrentProject.rules.html"
      />
    </template>
  </DesignDrawer>
</template>

<style>
#project-rules :is(h1, h2, h3) {
  font-size: var(--font-size-title);
  line-height: var(--line-height-title);
  font-weight: var(--font-weight-title);
  letter-spacing: var(--letter-spacing-title);
  color: var(--color-text-default);
  margin-bottom: 4px;

  &:not(:first-of-type) {
    margin-top: 32px;
  }
}

#project-rules p {
  font-size: var(--font-size-paragraph);
  line-height: var(--line-height-paragraph);
  font-weight: var(--font-weight-paragraph);
  letter-spacing: var(--letter-spacing-paragraph);
  color: var(--color-text-default);
  margin-block: 1rem;
}

#project-rules a {
  color: var(--color-accent-contrast);
  text-decoration: underline;
}

#project-rules ul {
  list-style-type: disc;
  padding-inline-start: 1.5em;
}

#project-rules li {
  font-size: var(--font-size-paragraph);
  line-height: var(--line-height-paragraph);
  font-weight: var(--font-weight-paragraph);
  letter-spacing: var(--letter-spacing-paragraph);
  color: var(--color-text-default);
  padding-inline-start: 0.5em;
  margin-block: 0.3em;
}
</style>
