<script setup lang="ts">
gql(`
	query SettingsPage {
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

const { track } = useAnalytics()
const { locale, locales, setLocale } = useI18n()

const localeName = computed(() => {
  return locales.value.find((l) => l.code === locale.value)?.name
})

const localeComp = computed({
  get() {
    return locale.value
  },
  set(v) {
    track(AnalyticsEvent.LanguageChanged, { from: locale.value, to: v })
    setLocale(v)
  },
})

const colorMode = useColorMode()
const colorModes = ['system', 'dark', 'light']

watch(
  () => colorMode.preference,
  (newMode, oldMode) => {
    if (oldMode) {
      track(AnalyticsEvent.ColorModeChanged, { from: oldMode, to: newMode })
    }
  },
)

const { me } = useAuth()

const { data, fetching, error, executeQuery: refetch } = useSettingsPageQuery()
</script>

<template>
  <PageLayout :title="$t('pages.settings')" :bottom-padding="false">
    <template #action>
      <NuxtLink :to="{ name: 'index' }">
        <DesignIconButton icon="lucide:x" />
      </NuxtLink>
    </template>

    <div class="space-y-list-section-gap">
      <DesignPanel class="gap-list-section-inset flex flex-col">
        <UDropdownMenu
          :ui="{
            content:
              'bg-background-raised ring-border-default rounded-list w-(--reka-dropdown-menu-trigger-width)',
          }"
          :content="{ align: 'end', side: 'bottom', sideOffset: -4 }"
          :items="
            locales.map((l) => ({
              label: l.name,
              value: l.code,
              type: 'checkbox',
              checked: l.code == localeComp,
              onSelect: () => (localeComp = l.code),
            }))
          "
          size="xl"
          checked-icon="lucide:check"
        >
          <div class="flex items-center justify-between gap-2.5 px-3 py-2">
            <p class="text-label">{{ $t('settings.language') }}</p>
            <DesignButton size="small" variant="secondary" class="grow-0">
              {{ localeName }}
            </DesignButton>
          </div>
        </UDropdownMenu>
        <hr class="border-border-default mx-3" />
        <UDropdownMenu
          :ui="{
            content:
              'bg-background-raised ring-border-default rounded-list w-(--reka-dropdown-menu-trigger-width)',
          }"
          :content="{ align: 'end', side: 'bottom', sideOffset: -4 }"
          :items="
            colorModes.map((mode) => ({
              label: $t('settings.colorModes.' + mode),
              value: mode,
              type: 'checkbox',
              checked: mode == colorMode.preference,
              onSelect: () => (colorMode.preference = mode),
            }))
          "
        >
          <div class="flex items-center justify-between gap-2.5 px-3 py-2">
            <p class="text-label">{{ $t('settings.colorMode') }}</p>
            <DesignButton size="small" variant="secondary" class="grow-0">
              {{ $t('settings.colorModes.' + colorMode.preference) }}
            </DesignButton>
          </div>
        </UDropdownMenu>
        <hr class="border-border-default mx-3" />
        <button class="flex items-center justify-between gap-2.5 px-3 py-2">
          <p class="text-label">{{ $t('settings.notifications') }}</p>
          <DesignButton size="small" variant="secondary" class="grow-0">
            {{ $t('settings.notificationsEnabled') }}
          </DesignButton>
        </button>
      </DesignPanel>
      <DesignPanel class="gap-list-section-inset flex flex-col">
        <button
          class="flex items-center justify-between gap-2.5 px-3 py-2 h-12"
        >
          <p class="text-label">{{ $t('settings.addToHomeScreen') }}</p>
          <Icon name="lucide:chevron-right" class="size-6" />
        </button>
      </DesignPanel>
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
      <div
        v-if="me"
        class="text-text-hint text-caption p-medium mt-auto text-center"
      >
        <p>{{ me.id }}</p>
        <p>{{ me.church.id }}</p>
      </div>
    </div>
  </PageLayout>
</template>
