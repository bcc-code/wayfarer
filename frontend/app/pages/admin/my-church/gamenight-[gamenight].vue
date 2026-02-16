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

      <h1 class="text-3xl font-semibold my-12">
        {{ $t('admin.churchHome.gameNight', { number: gamenight }) }}
      </h1>

      <h2 class="text-2xl font-semibold mb-4">Før du starter</h2>
      <ul class="space-y-4">
        <li>
          <UCheckbox
            label="Pass på at alle brukere har installert Interact appen"
          />
        </li>
      </ul>

      <div class="divide-y divide-muted mb-24">
        <AdminStep :step="1" title="Vis logo på storskjermen"> </AdminStep>
        <AdminStep :step="2" title="Spill av film 1"> </AdminStep>
        <AdminStep
          :step="3"
          title="Start tippingen"
          description="Alle må gå inn på 'Utfordringer' siden i Interact appen"
        >
          <UButton size="xl">Start tipping</UButton>
        </AdminStep>
        <AdminStep :step="4" title="Spill av film 2"> </AdminStep>
        <AdminStep :step="5" title="Send ut login-kode til pc26.bcc.media">
        </AdminStep>
        <AdminStep :step="6" title="Gå til adminsiden for å starte spillet">
          <UButton
            :loading="cryptexUrlStatus !== 'success'"
            size="xl"
            :to="cryptexUrl?.url"
            trailing-icon="lucide:arrow-right"
          >
            Gå til adminsiden for spillet
          </UButton>
        </AdminStep>
        <div class="py-12 pl-22.75 text-2xl font-semibold">
          <p>Fullfør steg 7 til 11 på den andre siden</p>
        </div>
        <AdminStep :step="12" title="Spill av film 3"> </AdminStep>
        <AdminStep :step="13" title="Gi deltakerne poeng for tippingen">
          <UButton size="xl">Frigi poeng nå</UButton>
        </AdminStep>
        <div class="py-24 text-center text-2xl font-semibold">
          <p>Du er nå ferdig! 🎉</p>
          <p>Game Night kan avsluttes.</p>
        </div>
      </div>
    </UContainer>
  </div>
</template>
