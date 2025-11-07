<script setup lang="ts">
gql(`
query ProfilePage {
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
`)

const { data, error, fetching } = useProfilePageQuery()
</script>

<template>
  <PageLayout title="Your profile">
    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <template v-else-if="data">
      <div class="flex flex-col items-center gap-2 p-8 text-center">
        <NuxtImg
          v-if="data.me.image"
          :src="data.me.image"
          height="64"
          width="64"
          class="rounded-full"
        />
        <h1 class="text-xl font-bold">{{ data.me.name }}</h1>
      </div>

      <div class="px-2">
        <h2 class="mb-3 text-lg font-semibold">Achievements</h2>
        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div v-for="project in data.me.projects" :key="project.id">
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
                    class="bg-accented grid size-16 place-items-center rounded"
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
  </PageLayout>
</template>
