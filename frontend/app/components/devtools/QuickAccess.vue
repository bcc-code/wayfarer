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
    <div class="fixed bottom-4 left-4">
      <UButton icon="lucide:copy" variant="soft" size="sm" @click="copyToken">
        Copy token
      </UButton>
    </div>
  </DevOnly>
</template>
