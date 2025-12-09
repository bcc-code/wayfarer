<script
  setup
  lang="ts"
  generic="TConsent extends PendingConsent | AcceptedConsent | RejectedConsent"
>
const props = defineProps<{
  consent: TConsent
}>()

const title = computed(() => {
  if (props.consent.__typename == 'UserConsent') {
    return props.consent.consent.title
  }
  return props.consent.title
})

const body = computed(() => {
  if (props.consent.__typename == 'UserConsent') {
    return props.consent.consent.body.html
  }
  return props.consent.body.html
})
</script>

<template>
  <DesignDrawer :title>
    <slot />
    <template #content>
      <div class="text-label text-text-default" v-html="body" />
    </template>
  </DesignDrawer>
</template>
