<script setup lang="ts">
gql(`
	query ConsentsDialog {
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

const {
  data,
  fetching,
  error,
  executeQuery: refetch,
} = useConsentsDialogQuery()

const { open } = useConsentsDialog()
const canClose = computed(
  () => !data.value?.me.consentStatus.pendingConsents.length,
)
watch(canClose, (cc) => {
  if (!cc) {
    open.value = true
  }
})
</script>

<template>
  <UModal
    v-model:open="open"
    :ui="{ content: 'bg-background-default' }"
    fullscreen
    modal
    :close="canClose"
    :dismissible="canClose"
  >
    <template #content="{ close }">
      <PageLayout :title="$t('pages.consents')">
        <template v-if="canClose" #action>
          <DesignIconButton icon="lucide:x" @click="close" />
        </template>

        <LoadingState v-if="fetching" />
        <ErrorState v-else-if="error" :error />
        <template v-else-if="data">
          <ConsentList
            v-if="data.me.consentStatus.pendingConsents.length"
            :consents="data.me.consentStatus.pendingConsents"
            :title="$t('consent.pending')"
            @update="refetch"
          />
          <ConsentList
            v-if="data.me.consentStatus.acceptedConsents.length"
            :consents="data.me.consentStatus.acceptedConsents"
            :title="$t('consent.accepted')"
            @update="refetch"
          />
          <ConsentList
            v-if="data.me.consentStatus.rejectedConsents.length"
            :consents="data.me.consentStatus.rejectedConsents"
            :title="$t('consent.rejected')"
            @update="refetch"
          />
        </template>
      </PageLayout>
    </template>
  </UModal>
</template>
