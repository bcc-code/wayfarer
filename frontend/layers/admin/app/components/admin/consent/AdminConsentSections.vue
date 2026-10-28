<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    consents: AdminConsentRow[]
    fetching?: boolean
    canManage?: boolean
    showProject?: boolean
    showArchived?: boolean
    publishedEmptyDescription?: string
  }>(),
  {
    fetching: false,
    canManage: false,
    showProject: false,
    showArchived: false,
    publishedEmptyDescription:
      'Opprett ditt første samtykke for å komme i gang.',
  },
)

const emit = defineEmits<{ refresh: [] }>()

const groups = computed(() => groupConsentsForAdmin(props.consents))
</script>

<template>
  <div class="space-y-8">
    <AdminSection
      v-if="groups.drafts.length > 0"
      title="Utkast"
      :count="groups.drafts.length"
    >
      <p class="text-muted mb-3 text-sm">
        Ikke publisert ennå, og ikke synlige for brukere. Åpne et utkast for å
        publisere det.
      </p>
      <AdminConsentTable
        :consents="groups.drafts"
        :can-manage
        :show-project
        :show-status="false"
        @refresh="emit('refresh')"
      />
    </AdminSection>

    <AdminSection title="Publiserte" :count="groups.published.length">
      <AdminConsentTable
        :consents="groups.published"
        :fetching
        :can-manage
        :show-project
        :empty-description="publishedEmptyDescription"
        @refresh="emit('refresh')"
      />
    </AdminSection>

    <AdminSection
      v-if="showArchived"
      title="Arkiverte"
      :count="groups.archived.length"
    >
      <p class="text-muted mb-3 text-sm">
        Tatt ut av bruk. Svarene brukerne har gitt er beholdt, og et arkivert
        samtykke kan gjenopprettes.
      </p>
      <AdminConsentTable
        :consents="groups.archived"
        :can-manage
        :show-project
        empty-title="Ingen arkiverte samtykker"
        empty-description="Samtykker du arkiverer havner her."
        @refresh="emit('refresh')"
      />
    </AdminSection>
  </div>
</template>
