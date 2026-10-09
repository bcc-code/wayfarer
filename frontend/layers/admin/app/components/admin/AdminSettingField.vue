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
  /** The staged value, which differs from `setting.value` once edited. */
  modelValue: string
  disabled?: boolean
}>()

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

// Free text that is really an enum: `warning` typed into log_level parses as
// `info` server-side rather than erroring.
const ENUM_OPTIONS: Record<string, string[]> = {
  log_level: ['debug', 'info', 'warn', 'error'],
}

const enumOptions = computed(() => ENUM_OPTIONS[props.setting.key])

const isJson = computed(() => props.setting.valueType === SettingValueType.Json)

const draft = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})

const isDirty = computed(
  () =>
    normalizeSetting(props.setting.valueType, props.modelValue) !==
    normalizeSetting(props.setting.valueType, props.setting.value),
)

const isValid = computed(() =>
  isSettingValueValid(props.setting.valueType, props.modelValue),
)

// The API takes the canonical string form; type=number is for the input only.
const numberDraft = computed({
  get: () => (draft.value === '' ? undefined : Number(draft.value)),
  set: (value) => (draft.value = value === undefined ? '' : String(value)),
})

const booleanDraft = computed({
  get: () => draft.value === 'true',
  set: (value) => (draft.value = String(value)),
})
</script>

<template>
  <div class="flex flex-wrap items-start justify-between gap-x-6 gap-y-3 py-4">
    <div class="min-w-48 flex-1">
      <div class="flex flex-wrap items-center gap-2">
        <p class="font-medium">{{ setting.key }}</p>
        <UBadge v-if="isDirty" color="primary" variant="subtle">Endret</UBadge>
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

    <div v-else-if="isJson" class="w-full space-y-2">
      <UTextarea
        v-model="draft"
        :rows="8"
        :disabled="disabled"
        :color="isValid ? undefined : 'error'"
        class="w-full"
        :ui="{ base: 'font-mono text-xs' }"
      />
      <p v-if="!isValid" class="text-error text-sm">Ugyldig JSON</p>
    </div>

    <div v-else class="flex w-full max-w-xs shrink-0 items-start gap-2">
      <USwitch
        v-if="setting.valueType === SettingValueType.Bool"
        v-model="booleanDraft"
        :disabled="disabled"
        class="py-1.5"
      />

      <template v-else>
        <USelect
          v-if="enumOptions"
          v-model="draft"
          :items="enumOptions"
          :disabled="disabled"
          class="w-full"
        />
        <UInput
          v-else-if="
            setting.valueType === SettingValueType.Int ||
            setting.valueType === SettingValueType.Float
          "
          v-model="numberDraft"
          type="number"
          :step="setting.valueType === SettingValueType.Float ? 0.05 : 1"
          :disabled="disabled"
          class="w-full"
        />
        <UInput v-else v-model="draft" :disabled="disabled" class="w-full" />
      </template>
    </div>
  </div>
</template>
