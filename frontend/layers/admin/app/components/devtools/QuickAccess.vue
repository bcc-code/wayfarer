<script setup lang="ts">
const { token } = useAuth()
const { copy } = useClipboard()
const toast = useToast()

function copyToken() {
  if (token.value) {
    copy(token.value)
    toast.add({
      title: 'Success',
      description: 'Token copied to clipboard',
      color: 'success',
      icon: 'lucide:copy-check',
    })
  }
}
</script>

<template>
  <DevOnly>
    <!-- Bottom-right: the admin shell puts the user menu in the sidebar
         footer, which is where this used to sit. -->
    <div class="fixed right-4 bottom-4 z-50">
      <UButton icon="lucide:copy" variant="soft" size="sm" @click="copyToken">
        Copy token
      </UButton>
    </div>
  </DevOnly>
</template>
