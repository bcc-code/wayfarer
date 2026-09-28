<script setup lang="ts">
const route = useRoute('plugins-ladder-to-heaven-excalibur')

const game = computed(() => route.query.game?.toString())
const queryParam = computed(() => {
  switch (game.value) {
    case 'escape':
      return '/help-them-escape'
    case 'climb':
      return '/help-them-climb'
    case 'prize':
      return '/help-them-win-the-prize'
    default:
      return '/help-them-escape'
  }
})

const config = useRuntimeConfig()

const { data, status } = useFetch<{ url: string }>(
  () =>
    `${config.public.apiUrl.replace('/graphql', '')}/plugins/ladder-to-heaven/excalibur-user-url`,
  {
    method: 'get',
    query: {
      path: queryParam.value,
    },
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${await useAuth().getAccessToken()}`,
    },
  },
)

watch(status, (s) => {
  if (s === 'success' && data.value?.url) {
    navigateTo(data.value.url, { external: true })
  } else if (s === 'error') {
    navigateTo('/challenges')
  }
})
</script>

<template>
  <LoadingState />
</template>
