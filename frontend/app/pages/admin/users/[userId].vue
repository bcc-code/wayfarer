<script setup lang="ts">
definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
	query AdminUserPage($id: ID!) {
		user(id: $id) {
			id
			name
			email
			membersId
			gender
			birthdate
			age
			image
			church {
				id
				name
			}
			roles {
				id
				role
				scope {
					id
					type
				}
			}
		}
	}
`)

const route = useRoute('admin-users-userId')

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = useAdminUserPageQuery({
  variables: {
    id: route.params.userId,
  },
  pause: computed(() => !isAuthReady.value),
})
</script>

<template>
  <div>
    <div class="border-default border-b py-2">
      <UContainer>
        <UBreadcrumb
          :items="[
            { label: 'Users', to: { name: 'admin-users' } },
            {
              label: data?.user.name ?? route.params.userId,
              to: {
                name: 'admin-users-userId',
                params: { userId: route.params.userId },
              },
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <LoadingState v-if="fetching" />
      <ErrorState v-else-if="error" :error />
      <div v-else-if="data" class="space-y-6">
        <!-- User Header -->
        <div class="flex items-center gap-6">
          <UAvatar
            :src="data.user.image ?? ''"
            :text="getInitials(data.user.name)"
            size="2xl"
          />
          <div>
            <h1 class="text-3xl font-bold">{{ data.user.name }}</h1>
            <p class="text-dimmed text-lg">{{ data.user.email }}</p>
          </div>
        </div>

        <!-- User Info Below Here -->
        <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
          <!-- Personal Information Card -->
          <UCard>
            <template #header>
              <h2 class="text-xl font-semibold">Personal Information</h2>
            </template>

            <dl class="space-y-4">
              <div>
                <dt class="text-dimmed text-sm">Members ID</dt>
                <dd class="text-base font-medium">{{ data.user.membersId }}</dd>
              </div>
              <div>
                <dt class="text-dimmed text-sm">Gender</dt>
                <dd class="text-base font-medium">
                  {{ capitalizeFirst(data.user.gender) }}
                </dd>
              </div>
              <div>
                <dt class="text-dimmed text-sm">Birthdate</dt>
                <dd class="text-base font-medium">
                  {{ formatDate(data.user.birthdate) }}
                </dd>
              </div>
              <div>
                <dt class="text-dimmed text-sm">Age</dt>
                <dd class="text-base font-medium">{{ data.user.age }} years</dd>
              </div>
            </dl>
          </UCard>

          <!-- Church Information Card -->
          <UCard>
            <template #header>
              <h2 class="text-xl font-semibold">Church</h2>
            </template>

            <dl class="space-y-4">
              <div>
                <dt class="text-dimmed text-sm">Church Name</dt>
                <dd class="font-medium">
                  {{ data.user.church.name }}
                </dd>
              </div>
              <div>
                <dt class="text-dimmed text-sm">Church ID</dt>
                <dd class="font-mono text-sm">
                  {{ data.user.church.id }}
                </dd>
              </div>
            </dl>
          </UCard>

          <!-- Roles Card -->
          <UCard class="md:col-span-2">
            <template #header>
              <h2 class="text-xl font-semibold">Roles & Permissions</h2>
            </template>

            <div v-if="data.user.roles.length > 0" class="space-y-3">
              <div
                v-for="role in data.user.roles"
                :key="role.id"
                class="border-default flex items-center justify-between rounded-md border p-3"
              >
                <div class="flex items-center gap-3">
                  <UBadge variant="soft" size="lg">
                    {{ role.role }}
                  </UBadge>
                  <div v-if="role.scope">
                    <span class="text-dimmed text-sm">Scope: </span>
                    <span class="text-sm font-medium">
                      {{ capitalizeFirst(role.scope.type) }}
                    </span>
                    <span class="text-dimmed ml-2 text-xs">
                      ({{ role.scope.id }})
                    </span>
                  </div>
                </div>
              </div>
            </div>
            <div v-else class="text-dimmed">No roles assigned</div>
          </UCard>
        </div>
      </div>
    </UContainer>
  </div>
</template>
