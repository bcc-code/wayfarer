<script setup lang="ts">
import { renderSVG } from 'uqr'

const props = defineProps<{
  challengeId: string
  challengeName: string
}>()

const isOpen = ref(false)
const toast = useToast()

// Deep link that drops the user straight into the challenge and auto-enrolls
// them (handled by the ?enroll=true param on the user-facing challenge page).
const challengeUrl = computed(() =>
  import.meta.client
    ? new URL(
        `/challenges/${props.challengeId}?enroll=true`,
        window.location.origin,
      ).href
    : '',
)

const qrSvg = computed(() =>
  challengeUrl.value
    ? renderSVG(challengeUrl.value, { border: 2, ecc: 'M' })
    : '',
)

function fileName(ext: string): string {
  const slug =
    props.challengeName
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-+|-+$/g, '') || 'utfordring'
  return `${slug}-qr.${ext}`
}

function triggerDownload(url: string, name: string) {
  const a = document.createElement('a')
  a.href = url
  a.download = name
  a.click()
}

function copyLink() {
  navigator.clipboard.writeText(challengeUrl.value)
  toast.add({ title: 'Lenke kopiert', color: 'success' })
}

function downloadSvg() {
  const blob = new Blob([qrSvg.value], { type: 'image/svg+xml' })
  const url = URL.createObjectURL(blob)
  triggerDownload(url, fileName('svg'))
  URL.revokeObjectURL(url)
}

async function downloadPng() {
  const size = 1024
  const svgBlob = new Blob([qrSvg.value], {
    type: 'image/svg+xml;charset=utf-8',
  })
  const svgUrl = URL.createObjectURL(svgBlob)
  try {
    const img = new Image()
    await new Promise<void>((resolve, reject) => {
      img.onload = () => resolve()
      img.onerror = () => reject(new Error('Kunne ikke laste QR-kode'))
      img.src = svgUrl
    })

    const canvas = document.createElement('canvas')
    canvas.width = size
    canvas.height = size
    const ctx = canvas.getContext('2d')
    if (!ctx) return
    ctx.fillStyle = '#ffffff'
    ctx.fillRect(0, 0, size, size)
    ctx.imageSmoothingEnabled = false
    ctx.drawImage(img, 0, 0, size, size)

    canvas.toBlob((blob) => {
      if (!blob) return
      const pngUrl = URL.createObjectURL(blob)
      triggerDownload(pngUrl, fileName('png'))
      URL.revokeObjectURL(pngUrl)
    }, 'image/png')
  } catch {
    toast.add({ title: 'Kunne ikke laste ned PNG', color: 'error' })
  } finally {
    URL.revokeObjectURL(svgUrl)
  }
}
</script>

<template>
  <div>
    <UButton
      variant="soft"
      icon="lucide:qr-code"
      @click="
        () => {
          isOpen = true
        }
      "
    >
      QR-kode
    </UButton>

    <UModal
      v-model:open="isOpen"
      title="QR-kode"
      description="Skann for å åpne og bli med på utfordringen"
    >
      <template #body>
        <div class="flex flex-col items-center gap-4">
          <div
            class="qr-container w-64 rounded-lg bg-white p-4"
            v-html="qrSvg"
          />
          <div class="flex w-full items-center gap-2">
            <UInput
              :model-value="challengeUrl"
              readonly
              class="flex-1"
              @focus="
                (e: FocusEvent) => (e.target as HTMLInputElement).select()
              "
            />
            <UButton
              variant="soft"
              icon="lucide:copy"
              aria-label="Kopier lenke"
              @click="copyLink"
            />
          </div>
          <div class="flex w-full gap-2">
            <UButton
              variant="outline"
              icon="lucide:download"
              class="flex-1 justify-center"
              @click="downloadSvg"
            >
              Last ned SVG
            </UButton>
            <UButton
              variant="outline"
              icon="lucide:download"
              class="flex-1 justify-center"
              @click="downloadPng"
            >
              Last ned PNG
            </UButton>
          </div>
        </div>
      </template>
    </UModal>
  </div>
</template>

<style scoped>
.qr-container :deep(svg) {
  width: 100%;
  height: auto;
  display: block;
}
</style>
