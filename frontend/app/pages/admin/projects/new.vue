<script setup lang="ts">
definePageMeta({
  layout: 'admin',
})

const state = reactive({
  name: '',
  description: '',
})
</script>

<template>
  <div>
    <div class="py-2 border-b border-default">
      <UContainer>
        <UBreadcrumb
          :items="[
            { label: 'Projects', to: { name: 'admin-projects' } },
            {
              label: state.name || 'New project',
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer>
      <UForm :state class="py-12 space-y-4">
        <UFormField name="name">
          <UInput
            v-model="state.name"
            :ui="{ base: 'text-3xl' }"
            placeholder="New project"
            size="xl"
            type="text"
            autofocus
          />
        </UFormField>
        <UFormField name="description" label="Description">
          <UTextarea v-model="state.description" />
        </UFormField>
      </UForm>
    </UContainer>
  </div>
</template>
