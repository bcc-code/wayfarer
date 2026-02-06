<script lang="ts" setup>
definePageMeta({
  layout: 'church-admin',
  middleware: ['admin'],
})

const route = useRoute('admin-my-church-gamenight-gamenight')
const gamenight = computed(() => route.params.gamenight)

const { getAccessToken } = useAuth()
const config = useRuntimeConfig()
const { data: cryptexUrl, status: cryptexUrlStatus } = useFetch<{
  url: string
}>(
  `${config.public.apiUrl.replace('/graphql', '')}/plugins/ladder-to-heaven/cryptex-admin-url`,
  {
    method: 'get',
    headers: {
      Authorization: `Bearer ${await getAccessToken()}`,
    },
  },
)
</script>

<template>
  <div>
    <div class="border-default border-b py-2">
      <UContainer>
        <UBreadcrumb
          :items="[
            {
              label: $t('admin.breadcrumb.home'),
              to: { name: 'admin-my-church' },
            },
            {
              label: $t('admin.churchHome.gameNights'),
              to: { name: 'admin-my-church-gamenights' },
            },
            {
              label: $t('admin.churchHome.gameNight', { number: gamenight }),
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-6 relative">
      <UButton
        color="neutral"
        variant="soft"
        size="lg"
        :to="{ name: 'admin-my-church' }"
      >
        <Icon name="lucide:arrow-left" />
        {{ $t('admin.common.back') }}
      </UButton>

      <h1 class="text-3xl font-semibold mt-12">
        {{ $t('admin.churchHome.gameNight', { number: gamenight }) }}
      </h1>

      <UButton
        :loading="cryptexUrlStatus !== 'success'"
        size="xl"
        :to="cryptexUrl?.url"
      >
        Gå til Game Night kontrollpanel
      </UButton>
    </UContainer>
  </div>
</template>
