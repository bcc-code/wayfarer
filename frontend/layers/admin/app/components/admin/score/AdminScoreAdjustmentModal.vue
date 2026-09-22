<script setup lang="ts">
const props = defineProps<{ projectId: string }>()

const open = defineModel<boolean>('open', { required: true })
const emit = defineEmits<{ created: [] }>()

gql(`
  mutation CreateScoreAdjustment($input: CreateScoreAdjustmentInput!) {
    createScoreAdjustment(input: $input) {
      id
      points
      reason
    }
  }
`)

const userId = ref('')
const points = ref(0)
const reason = ref('')

const { executeMutation: createAdjustment, fetching } =
  useCreateScoreAdjustmentMutation()
const toast = useToast()

// Zero would be a journal row that changes nothing.
const canSubmit = computed(() => !!userId.value && points.value !== 0)

function reset() {
  userId.value = ''
  points.value = 0
  reason.value = ''
}

watch(open, (isOpen) => {
  if (isOpen) reset()
})

async function handleSubmit() {
  if (!canSubmit.value) return

  const result = await createAdjustment({
    input: {
      projectId: props.projectId,
      userId: userId.value,
      points: points.value,
      reason: reason.value || undefined,
    },
  })

  if (result.error) {
    toast.add({
      title: 'Kunne ikke opprette poengjustering',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Poengjustering opprettet',
    description: `${points.value >= 0 ? '+' : ''}${formatNumber(points.value)} poeng`,
    color: 'success',
  })

  open.value = false
  emit('created')
}
</script>

<template>
  <UModal v-model:open="open">
    <template #header>
      <h3 class="text-lg font-semibold">Ny poengjustering</h3>
    </template>

    <template #body>
      <div class="space-y-4">
        <UFormField label="Bruker" required>
          <!-- Scoped to the project: an adjustment belongs to one project, so
               its participants are the only valid targets. -->
          <AdminUserPicker v-model="userId" :project-id="projectId" />
        </UFormField>

        <UFormField
          label="Poeng"
          required
          description="Bruk negative verdier for å trekke fra poeng"
        >
          <UInput
            v-model.number="points"
            type="number"
            class="w-full"
            placeholder="100"
          />
        </UFormField>

        <UFormField label="Grunn">
          <UTextarea
            v-model="reason"
            class="w-full"
            autoresize
            placeholder="Grunn for denne justeringen..."
          />
        </UFormField>
      </div>
    </template>

    <template #footer>
      <div class="flex w-full justify-end gap-3">
        <UButton
          variant="ghost"
          color="neutral"
          @click="
            () => {
              open = false
            }
          "
        >
          Avbryt
        </UButton>
        <UButton
          :disabled="!canSubmit"
          :loading="fetching"
          @click="handleSubmit"
        >
          Opprett justering
        </UButton>
      </div>
    </template>
  </UModal>
</template>
