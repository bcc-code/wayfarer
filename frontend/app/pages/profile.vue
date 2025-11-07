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
      <div v-if="data.me.image" class="flex items-center justify-center p-4">
        <div
          class="shadow-large bg-background-raised p-list-section-inset aspect-square size-42 rounded-full"
        >
          <NuxtImg
            :src="data.me.image"
            height="160"
            width="160"
            class="bg-background-default text-accent-contrast rounded-full"
          />
        </div>
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
