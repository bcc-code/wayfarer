<script setup lang="ts">
import { SettingValueType } from '~/api/generated'

export interface SettingFieldValue {
  key: string
  value: string
  valueType: SettingValueType
  description?: string | null
  requiresRestart: boolean
  envVar?: string | null
  editable: boolean
}

const props = defineProps<{
  setting: SettingFieldValue
  saving?: boolean
}>()

const emit = defineEmits<{ save: [value: string] }>()

// Free text that is really an enum: `warning` typed into log_level parses as
// `info` server-side rather than erroring.
const ENUM_OPTIONS: Record<string, string[]> = {
  log_level: ['debug', 'info', 'warn', 'error'],
}

const enumOptions = computed(() => ENUM_OPTIONS[props.setting.key])

const draft = ref(props.setting.value)
watch(
  () => props.setting.value,
  (value) => (draft.value = value),
)

const isDirty = computed(() => draft.value !== props.setting.value)

// The API takes the canonical string form; type=number is for the input only.
const numberDraft = computed({
  get: () => (draft.value === '' ? undefined : Number(draft.value)),
  set: (value) => (draft.value = value === undefined ? '' : String(value)),
})

const booleanDraft = computed({
  get: () => draft.value === 'true',
  set: (value) => {
    draft.value = String(value)
    // A switch has no separate save step.
    emit('save', draft.value)
  },
})

const isValid = computed(() => {
  if (props.setting.valueType === SettingValueType.Json) {
    try {
      JSON.parse(draft.value)
      return true
    } catch {
      return false
    }
  }
  if (
    props.setting.valueType === SettingValueType.Int ||
    props.setting.valueType === SettingValueType.Float
  ) {
    return draft.value !== '' && Number.isFinite(Number(draft.value))
  }
  return draft.value !== ''
})

function save() {
  if (!isDirty.value || !isValid.value) return
  emit('save', draft.value)
}
</script>

<template>
  <div class="flex flex-wrap items-start justify-between gap-x-6 gap-y-3 py-4">
    <div class="min-w-48 flex-1">
      <div class="flex flex-wrap items-center gap-2">
        <p class="font-medium">{{ setting.key }}</p>
        <UBadge v-if="setting.requiresRestart" color="warning" variant="subtle">
          Krever omstart
        </UBadge>
      </div>
      <p v-if="setting.description" class="text-muted mt-0.5 text-sm">
        {{ setting.description }}
      </p>
      <p v-if="setting.envVar" class="text-dimmed mt-0.5 text-xs">
        Overstyrer <code>{{ setting.envVar }}</code>
      </p>
    </div>

    <div v-if="!setting.editable" class="w-full max-w-xs shrink-0 text-right">
      <code class="text-dimmed text-sm">{{ setting.value }}</code>
      <p class="text-dimmed text-xs">Ikke redigerbar</p>
    </div>

    <div v-else class="flex w-full max-w-xs shrink-0 items-start gap-2">
      <USwitch
        v-if="setting.valueType === SettingValueType.Bool"
        v-model="booleanDraft"
        :disabled="saving"
        class="py-1.5"
      />

      <template v-else>
        <USelect
          v-if="enumOptions"
          v-model="draft"
          :items="enumOptions"
          :disabled="saving"
          class="w-full"
        />
        <UTextarea
          v-else-if="setting.valueType === SettingValueType.Json"
          v-model="draft"
          :rows="3"
          :disabled="saving"
          :color="isValid ? undefined : 'error'"
          class="w-full font-mono"
        />
        <UInput
          v-else-if="
            setting.valueType === SettingValueType.Int ||
            setting.valueType === SettingValueType.Float
          "
          v-model="numberDraft"
          type="number"
          :step="setting.valueType === SettingValueType.Float ? 0.05 : 1"
          :disabled="saving"
          class="w-full"
          @keyup.enter="save"
        />
        <UInput
          v-else
          v-model="draft"
          :disabled="saving"
          class="w-full"
          @keyup.enter="save"
        />

        <UButton
          v-if="isDirty"
          icon="lucide:check"
          :loading="saving"
          :disabled="!isValid"
          @click="save"
        >
          Lagre
        </UButton>
      </template>
    </div>
  </div>
</template>
