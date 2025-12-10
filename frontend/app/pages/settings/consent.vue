<script setup lang="ts">
const { data, fetching, error, executeQuery: refetch } = useConsentsPageQuery()
</script>

<template>
  <PageLayout :title="$t('pages.consents')">
    <template #action>
      <NuxtLink :to="{ name: 'settings' }">
        <DesignIconButton icon="lucide:x" />
      </NuxtLink>
    </template>

    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <div v-else-if="data" class="space-y-list-section-gap">
      <ConsentCard
        v-for="consent in data.me.consentStatus.pendingConsents"
        :key="consent.id"
        :consent
        @update="refetch"
      />
      <ConsentCard
        v-for="consent in data.me.consentStatus.acceptedConsents"
        :key="consent.id"
        :consent
        @update="refetch"
      />
      <ConsentCard
        v-for="consent in data.me.consentStatus.rejectedConsents"
        :key="consent.id"
        :consent
        @update="refetch"
      />
    </div>
  </PageLayout>
</template>
