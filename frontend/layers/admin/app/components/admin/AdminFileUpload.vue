<script setup lang="ts">
interface UploadResponse {
  id: string
  filename: string
  storedFilename: string
  fileSize: number
  mimeType: string
  publicUrl: string
  uploadedBy: string
  createdAt: string
}

const props = withDefaults(
  defineProps<{
    accept?: string
    label?: string
    description?: string
    maxSize?: number
  }>(),
  {
    accept: 'image/*',
    label: 'Last opp bilde',
    description: 'Slipp et bilde her eller klikk for å laste opp',
    maxSize: 30 * 1024 * 1024, // 30MB
  },
)

const modelValue = defineModel<string | null>()

const emit = defineEmits<{
  uploaded: [response: UploadResponse]
  error: [error: Error]
}>()

const toast = useToast()
const config = useRuntimeConfig()
const { getAccessToken } = useAuth()

// Derive base URL from apiUrl (remove /graphql)
const baseUrl = computed(() => {
  const apiUrl = config.public.apiUrl
  return apiUrl.replace(/\/graphql$/, '')
})

const uploading = ref(false)
const selectedFile = ref<File | null>(null)

watch(selectedFile, async (file) => {
  if (!file) return
  await handleUpload(file)
  selectedFile.value = null
})

async function handleUpload(file: File) {
  if (file.size > props.maxSize) {
    const maxSizeMB = Math.round(props.maxSize / 1024 / 1024)
    const error = new Error(`File size exceeds ${maxSizeMB}MB limit`)
    emit('error', error)
    toast.add({
      title: 'Filen er for stor',
      description: `Maks filstørrelse er ${maxSizeMB}MB`,
      color: 'error',
    })
    return
  }

  uploading.value = true

  try {
    const token = await getAccessToken()
    if (!token) {
      throw new Error('Ikke autentisert')
    }

    const formData = new FormData()
    formData.append('file', file)

    const response = await $fetch<UploadResponse>(
      `${baseUrl.value}/api/upload`,
      {
        method: 'POST',
        body: formData,
        headers: {
          Authorization: `Bearer ${token}`,
        },
      },
    )

    modelValue.value = response.publicUrl
    emit('uploaded', response)
    toast.add({
      title: 'Opplastning fullført',
      description: 'Filen ble lastet opp',
      color: 'success',
    })
  } catch (err) {
    const error = err instanceof Error ? err : new Error('Opplasting feilet')
    emit('error', error)
    toast.add({
      title: 'Opplasting feilet',
      description: error.message,
      color: 'error',
    })
  } finally {
    uploading.value = false
  }
}

function clear() {
  modelValue.value = null
}
</script>

<template>
  <div class="relative">
    <!-- Preview current/uploaded image -->
    <div v-if="modelValue" class="relative inline-block">
      <img
        :src="modelValue"
        :alt="label"
        class="max-h-48 rounded-lg object-cover"
      />
      <UButton
        icon="i-lucide-x"
        size="xs"
        color="error"
        variant="solid"
        class="absolute right-2 top-2"
        :disabled="uploading"
        @click="clear"
      />
    </div>

    <!-- File upload dropzone -->
    <UFileUpload
      v-else
      v-model="selectedFile"
      :accept="accept"
      :disabled="uploading"
    >
      <template #default="{ open }">
        <div
          class="flex cursor-pointer flex-col items-center justify-center gap-2 rounded-lg border-2 border-dashed border-gray-300 p-6 transition-colors hover:border-gray-400 dark:border-gray-600 dark:hover:border-gray-500"
          @click="() => open()"
        >
          <UIcon
            v-if="uploading"
            name="i-lucide-loader-2"
            class="h-8 w-8 animate-spin text-gray-400"
          />
          <UIcon v-else name="i-lucide-upload" class="h-8 w-8 text-gray-400" />
          <div class="text-center">
            <p class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ uploading ? 'Laster opp...' : label }}
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              {{ description }}
            </p>
          </div>
        </div>
      </template>
    </UFileUpload>
  </div>
</template>
