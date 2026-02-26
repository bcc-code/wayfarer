<script setup lang="ts">
import { formatNumber } from '~/utils/formatters'

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
    correctCount?: number | null
    totalCount?: number | null
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
    correctCount: null,
    totalCount: null,
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
  return props.pointsEarned + (props.betAmount ?? 0)
})

const multiplier = computed(() => {
  if (!props.betAmount || props.betAmount === 0) return 0
  return resultAmount.value / props.betAmount
})

const formattedMultiplier = computed(() => {
  const m = multiplier.value
  return `x ${Number.isInteger(m) ? m : parseFloat(m.toFixed(2))}`
})

const correctCountLabel = computed(() => {
  if (props.correctCount === null || props.totalCount === null) return null
  if (props.correctCount === 0) return t('quiz.betting.correctCountNone')
  if (props.correctCount === props.totalCount)
    return t('quiz.betting.correctCountAll')
  return t(
    'quiz.betting.correctCount',
    { count: props.correctCount },
    props.correctCount,
  )
})
</script>

<template>
  <!-- Results mode: show win/loss -->
  <div
    v-if="mode === 'results'"
    class="flex flex-col pb-default pt-medium text-on-accent"
  >
    <!-- Info rows -->
    <div
      class="flex justify-between items-center pb-3 border-b border-on-accent/20"
    >
      <span class="text-body">{{ t('quiz.betting.resultYourBet') }}</span>
      <span class="text-label tabular-nums">
        {{ formatNumber(betAmount ?? 0) }}
      </span>
    </div>
    <div
      v-if="correctCountLabel !== null"
      class="flex justify-between items-center py-3 border-b border-on-accent/20"
    >
      <span class="text-body">{{ correctCountLabel }}</span>
      <span
        v-if="correctCount !== null && correctCount > 0"
        class="text-label tabular-nums"
      >
        {{ formattedMultiplier }}
      </span>
    </div>
    <div
      v-if="
        correctCountLabel !== null && correctCount !== null && correctCount > 0
      "
      class="flex justify-between items-center py-3 border-b border-on-accent/20"
    >
      <span class="text-body">{{ t('quiz.betting.bettingResult') }}</span>
      <span class="text-label tabular-nums">
        {{ formatNumber(resultAmount) }}
      </span>
    </div>

    <!-- Large result display -->
    <div class="text-center pt-default">
      <p class="text-caption">
        {{
          isWin ? t('quiz.betting.pointsEarned') : t('quiz.betting.pointsLost')
        }}
      </p>
      <p class="text-hero tabular-nums leading-tight">
        {{
          isWin
            ? `+${formatNumber(pointsEarned ?? 0)}`
            : formatNumber(pointsEarned ?? 0)
        }}
      </p>
    </div>
  </div>

  <!-- Betting mode: show slider and stats -->
  <div v-else class="flex flex-col gap-default pb-default">
    <div class="grid grid-cols-2 divide-x divide-border-default">
      <div class="text-center pr-default pl-medium">
        <p class="text-text-muted text-caption">
          {{ t('quiz.betting.remainingPoints') }}
        </p>
        <p class="text-heading tabular-nums text-text-default">
          {{ remainingAvailablePoints }}
        </p>
      </div>
      <div class="text-center pl-default pr-medium">
        <p class="text-text-muted text-caption">
          {{ t('quiz.betting.yourBet') }}
        </p>
        <p class="text-heading tabular-nums text-accent-contrast">
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
        :step="1"
        :disabled="disabled"
      />
      <p v-if="maxBetMessage" class="text-caption text-text-muted text-center">
        {{ maxBetMessage }}
      </p>
    </div>
  </div>
</template>
