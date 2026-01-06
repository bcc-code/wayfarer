<script setup lang="ts">
definePageMeta({
  layout: 'church-admin',
  middleware: ['admin'],
})

gql(`
  query AdminMyChurchUnitsPage($churchId: String!) {
    myCurrentProject {}
  }
`)
</script>

<template>
  <div>
    <div class="border-default border-b py-2">
      <UContainer>
        <UBreadcrumb
          :items="[
            {
              label: 'Forside',
              to: { name: 'admin-my-church' },
            },
            {
              label: 'Units',
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-6">
      <UButton
        color="neutral"
        variant="soft"
        size="lg"
        :to="{ name: 'admin-my-church' }"
      >
        <Icon name="lucide:arrow-left" />
        Tilbake til forsiden
      </UButton>

      <h1 class="text-4xl mt-12 mb-4">Units</h1>
      <UButton icon="lucide:plus" size="lg">Opprett ny unit</UButton>
    </UContainer>
  </div>
</template>
