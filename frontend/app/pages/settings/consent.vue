<script setup lang="ts">
gql(`
	query ConsentsPage {
		me {
			consentStatus {
				pendingConsents {
					__typename
					id
					key
					version
					title
					body {
						html
					}
					publishedAt
					managedBy
					managementType
				}
				acceptedConsents {
					__typename
					id
					consent {
						title
						body {
							html
						}
						managedBy
						managementType
						url
					}
					action
					actionDate
				}
				rejectedConsents {
					__typename
					id
					consent {
						title
						body {
							html
						}
						managedBy
						managementType
						url
					}
					action
					actionDate
				}
			}
		}
	}
`)

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
    <template v-else-if="data">
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
    </template>
  </PageLayout>
</template>
