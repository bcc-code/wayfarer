<script setup lang="ts">
import { useProfilePageQuery } from '~/api/generated'

gql(`
query ProfilePage {
  user {
    me {
      id
      name
      image
      church {
        id
        name
      }
      projects {
        id
        achievements {
          id
          name
          image
          hidden
          achievedAt
          points
        }
      }
    }
  }
}
`)

const { data, error, fetching } = useProfilePageQuery()
</script>

<template>
  <div class="h-full">
    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <template v-else-if="data">
      <div class="p-8 text-center flex flex-col items-center gap-2">
        <NuxtImg
          v-if="data.user.me.image"
          :src="data.user.me.image"
          height="64"
          width="64"
          class="rounded-full"
        />
        <h1 class="text-xl font-bold">{{ data.user.me.name }}</h1>
      </div>

      <div class="px-2">
        <h2 class="text-lg font-semibold mb-3">Achievements</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div v-for="project in data.user.me.projects" :key="project.id">
            <UCard>
              <div class="grid grid-cols-4">
                <div
                  v-for="achievement in project.achievements.filter(
                    (a) => !a.hidden,
                  )"
                  :key="achievement.id"
                >
                  <NuxtImg
                    v-if="achievement.image && achievement.achievedAt"
                    :src="achievement.image"
                    height="64"
                    width="64"
                    class="shrink-0 overflow-hidden rounded"
                  />
                  <div
                    v-else-if="!achievement.achievedAt"
                    class="size-16 rounded bg-accented grid place-items-center"
                  >
                    ?
                  </div>
                </div>
              </div>
            </UCard>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
