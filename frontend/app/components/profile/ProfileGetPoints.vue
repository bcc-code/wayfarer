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

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = useProjectRulesQuery({
  pause: computed(() => !isAuthReady.value),
})
</script>

<template>
  <UModal
    :ui="{ content: 'bg-background-default' }"
    :transition="false"
    fullscreen
    modal
  >
    <slot />
    <template #content="{ close }">
      <PageLayout :title="$t('pages.getPoints')">
        <template #action>
          <DesignIconButton icon="lucide:x" @click="close" />
        </template>

        <LoadingState v-if="fetching" />
        <ErrorState v-else-if="error" :error />
        <div
          v-else-if="data?.myCurrentProject.rules"
          id="project-rules"
          class="p-default"
          v-html="data.myCurrentProject.rules.html"
        />
      </PageLayout>
    </template>
  </UModal>
</template>

<style>
#project-rules h1,
#project-rules h2,
#project-rules h3 {
  font-size: var(--font-size-title);
  line-height: var(--line-height-title);
  font-weight: var(--font-weight-title);
  letter-spacing: var(--letter-spacing-title);
  color: var(--color-text-default);
  margin-bottom: 4px;
  margin-top: 16px;
}

#project-rules p {
  font-size: var(--font-size-paragraph);
  line-height: var(--line-height-paragraph);
  font-weight: var(--font-weight-paragraph);
  letter-spacing: var(--letter-spacing-paragraph);
  color: var(--color-text-default);
}
</style>
