<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    availablePoints: number
    minPercentage?: number | null
    maxPercentage?: number | null
    minAbsolute?: number | null
    maxAbsolute?: number | null
    disabled?: boolean
    mode?: 'betting' | 'locked' | 'results'
    pointsEarned?: number | null
    betAmount?: number | null
  }>(),
  {
    minPercentage: null,
    maxPercentage: null,
    minAbsolute: null,
    maxAbsolute: null,
    disabled: false,
    mode: 'betting',
    pointsEarned: null,
    betAmount: null,
  },
)

const betAmountModel = defineModel<number>({ default: 0 })

const { t } = useI18n()

// Calculate effective min/max based on both percentage and absolute limits
const effectiveMin = computed(() => {
  const percentageMin = props.minPercentage
    ? Math.floor((props.availablePoints * props.minPercentage) / 100)
    : 0
  const absoluteMin = props.minAbsolute ?? 0
  return Math.max(percentageMin, absoluteMin)
})

const effectiveMax = computed(() => {
  const percentageMax = props.maxPercentage
    ? Math.floor((props.availablePoints * props.maxPercentage) / 100)
    : props.availablePoints
  const absoluteMax = props.maxAbsolute ?? props.availablePoints
  return Math.min(percentageMax, absoluteMax, props.availablePoints)
})

// Calculate step size based on range
const stepSize = computed(() => {
  const range = effectiveMax.value - effectiveMin.value
  if (range <= 100) return 10
  if (range <= 500) return 25
  if (range <= 1000) return 50
  return 100
})

// Max bet message showing the percentage limit
const maxBetMessage = computed(() => {
  if (props.maxPercentage) {
    return t('quiz.betting.maxBetMessage', { percentage: props.maxPercentage })
  }
  return ''
})

const remainingAvailablePoints = computed(() => {
  return props.availablePoints - betAmountModel.value
})

// Results mode computations
const isWin = computed(() => {
  if (props.pointsEarned === null || props.pointsEarned === undefined)
    return null
  return props.pointsEarned >= 0
})

const resultAmount = computed(() => {
  if (props.pointsEarned === null || props.pointsEarned === undefined) return 0
  return Math.abs(props.pointsEarned)
})
</script>

<template>
  <!-- Results mode: show win/loss -->
  <div
    v-if="mode === 'results'"
    class="flex flex-col gap-default pb-default text-on-accent"
  >
    <p class="text-center text-caption opacity-50">
      {{ t('quiz.betting.results') }}
    </p>
    <div class="grid grid-cols-2 divide-x divide-on-accent/20">
      <div class="text-center pr-default pl-medium">
        <p class="text-caption">
          {{ isWin ? t('quiz.betting.winnings') : t('quiz.betting.losses') }}
        </p>
        <p class="text-heading tabular-nums">
          {{ resultAmount }}
        </p>
      </div>
      <div class="text-center pl-default pr-medium">
        <p class="text-caption">
          {{ t('quiz.betting.yourPoints') }}
        </p>
        <p class="text-heading tabular-nums">
          {{ availablePoints + resultAmount }}
        </p>
      </div>
    </div>
  </div>

  <!-- Betting mode: show slider and stats -->
  <div v-else class="flex flex-col gap-default pb-default">
    <div class="grid grid-cols-2 divide-x divide-border-default">
      <div class="text-center pr-default pl-medium">
        <p class="text-text-hint text-caption">
          {{ t('quiz.betting.remainingPoints') }}
        </p>
        <p class="text-heading tabular-nums text-text-default">
          {{ remainingAvailablePoints }}
        </p>
      </div>
      <div class="text-center pl-default pr-medium">
        <p class="text-accent-contrast text-caption">
          {{ t('quiz.betting.yourBet') }}
        </p>
        <p class="text-heading tabular-nums text-text-default">
          {{ betAmountModel }}
        </p>
      </div>
    </div>

    <div
      v-if="disabled || mode === 'locked'"
      class="bg-background-indent text-center p-small rounded-modal text-accent-positive text-caption flex items-center justify-center gap-1"
    >
      {{ t('quiz.betting.registeredBet') }}
      <Icon name="lucide:check" />
    </div>
    <div v-else class="space-y-2">
      <DesignSlider
        v-model="betAmountModel"
        :min="effectiveMin"
        :max="effectiveMax"
        :step="stepSize"
        :disabled="disabled"
      />
      <p v-if="maxBetMessage" class="text-caption text-text-hint text-center">
        {{ maxBetMessage }}
      </p>
    </div>
  </div>
</template>
