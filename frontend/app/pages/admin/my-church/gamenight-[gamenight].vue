<script lang="ts" setup>
import { onKeyDown } from '@vueuse/core'

definePageMeta({
  layout: 'church-admin',
  middleware: ['admin'],
})

const route = useRoute('admin-my-church-gamenight-gamenight')
const gamenight = computed(() => route.params.gamenight)

// We need to keep track of quite a bit of state
// that should be persisted between refreshes
const state = useLocalStorage(`state:gamenight-${gamenight.value}`, {
  filmsReady: false,
  bigScreenReady: false,
  step: 1,
  quizSessionId: null as string | null,
  quizSessionState: null as QuizSessionState | null,
})

onKeyDown('ArrowDown', () => {
  state.value.step = Math.min(9, state.value.step + 1)
})
onKeyDown('ArrowUp', () => {
  state.value.step = Math.max(1, state.value.step - 1)
})

const config = useRuntimeConfig()
const { getAccessToken } = useAuth()
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

const bigScreenUrl = computed(() => {
  if (!cryptexUrl.value) return
  return `${cryptexUrl.value.url}&redirect=/so`
})

const filmsUrl = 'https://pc26.bcc.media/gamenight-filmer-1237'
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
        :to="{ name: 'admin-my-church-gamenights' }"
      >
        <Icon name="lucide:arrow-left" />
        {{ $t('admin.common.back') }}
      </UButton>

      <h1 class="text-3xl font-semibold my-12">
        {{ $t('admin.churchHome.gameNight', { number: gamenight }) }}
      </h1>

      <section class="space-y-8 my-6 rounded-2xl p-8 border border-muted">
        <h2 class="text-2xl font-semibold">Før du starter</h2>
        <div>
          <UCheckbox
            v-model="state.filmsReady"
            label="Pass på at filmer er klare for avspilling"
          />
          <NuxtLink
            :to="filmsUrl"
            external
            target="_blank"
            class="pl-7 underline text-primary flex items-center gap-1"
          >
            Du finner filmene her
            <Icon name="lucide:external-link" />
          </NuxtLink>
        </div>
        <div>
          <UCheckbox
            v-model="state.bigScreenReady"
            label="Ha storskjermvisningen for spillet klart"
          />
          <NuxtLink
            :to="bigScreenUrl"
            external
            target="_blank"
            :data-loading="!bigScreenUrl"
            class="pl-7 underline text-primary data-[loading=true]:pointer-events-none data-[loading=true]:opacity-50 flex items-center gap-1"
          >
            Du finner storskjermvisningen her
            <Icon name="lucide:external-link" />
          </NuxtLink>
        </div>
      </section>

      <div class="divide-y divide-muted mb-24">
        <AdminStep
          :step="1"
          title="Vis Ladder to Heaven logo på storskjermen"
          description="Dette kan være på storskjermen når ungdommene kommer inn i salen"
          :active="state.step === 1"
        >
          <UButton
            size="xl"
            trailing-icon="lucide:external-link"
            :to="bigScreenUrl"
            target="_blank"
            external
          >
            Åpne storskjermvisningen i ny fane
          </UButton>
          <UButton
            class="flex mt-2"
            variant="ghost"
            trailing-icon="lucide:check"
            size="xl"
            @click="state.step++"
          >
            Logo vises på storskjermen
          </UButton>
        </AdminStep>
        <AdminStep
          :step="2"
          title="Spill av film 1"
          description="Dette gjøres når ungdommene satt seg og kvelden er i gang"
          :active="state.step === 2"
        >
          <UButton
            size="xl"
            :to="`${filmsUrl}?film=1`"
            trailing-icon="lucide:external-link"
          >
            Gå til film 1
          </UButton>
          <UButton
            class="flex mt-2"
            variant="ghost"
            trailing-icon="lucide:check"
            size="xl"
            @click="state.step++"
          >
            Filmen er avspilt
          </UButton>
        </AdminStep>
        <AdminStep
          :step="3"
          title="Start tippingen"
          description="Alle må gå inn på 'Utfordringer' siden i Interact appen"
          :active="state.step === 3"
        >
          <UButton size="xl">Start tipping</UButton>
        </AdminStep>
        <AdminStep
          :step="4"
          title="Spill av film 2"
          description="Dette gjøres etter at tipping er gjennomført"
          :active="state.step === 4"
        >
          <UButton
            size="xl"
            :to="`${filmsUrl}?film=2`"
            trailing-icon="lucide:external-link"
          >
            Gå til film 2
          </UButton>
          <UButton
            class="flex mt-2"
            variant="ghost"
            trailing-icon="lucide:check"
            size="xl"
            @click="state.step++"
          >
            Filmen er avspilt
          </UButton>
        </AdminStep>
        <AdminStep
          :step="5"
          title="Send ut login-kode til pc26.bcc.media"
          :active="state.step === 5"
        >
        </AdminStep>
        <AdminStep
          :step="6"
          title="Gå til adminsiden for å starte spillet"
          :active="state.step === 6"
        >
          <UButton
            :loading="cryptexUrlStatus !== 'success'"
            size="xl"
            :to="cryptexUrl?.url"
            trailing-icon="lucide:arrow-right"
          >
            Gå til adminsiden for spillet
          </UButton>
        </AdminStep>
        <div
          class="py-12 pl-22.75 text-2xl font-semibold data-[active=false]:pointer-events-none data-[active=false]:opacity-50"
          :data-active="state.step === 6"
        >
          <p>Fullfør steg 7 til 11 på den andre siden</p>
        </div>
        <AdminStep
          :step="12"
          title="Spill av film 3"
          :active="state.step === 7"
        >
        </AdminStep>
        <AdminStep
          :step="13"
          title="Gi deltakerne poeng for tippingen"
          :active="state.step === 8"
        >
          <UButton size="xl">Frigi poeng nå</UButton>
        </AdminStep>
        <div
          class="py-24 text-center text-2xl font-semibold data-[active=false]:pointer-events-none data-[active=false]:opacity-50"
          :data-active="state.step === 9"
        >
          <p>Du er nå ferdig! 🎉</p>
          <p>Game Night {{ gamenight }} kan trygt avsluttes.</p>
        </div>
      </div>
    </UContainer>
  </div>
</template>
