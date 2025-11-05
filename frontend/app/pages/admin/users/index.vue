<script setup lang="ts">
definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
	query AdminUsersPage {
		users {
			id
			name
		}
	}
`)

const { data, fetching, error } = useAdminUsersPageQuery()
</script>

<template>
  <UContainer class="py-12">
    <h1 class="text-3xl">Users</h1>
    <ErrorState v-if="error" :error />
    <UTable v-else-if="data" :data="data?.users" :loading="fetching" />
  </UContainer>
</template>
