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
</script>

<template>
  <!-- Results mode: show win/loss -->
  <div
    v-if="mode === 'results'"
    class="flex flex-col gap-default py-default text-on-accent"
  >
    <div class="grid grid-cols-2 divide-x divide-on-accent/20">
      <div class="text-center pl-default pr-medium">
        <p class="text-caption">
          {{ t('quiz.betting.resultYourBet') }}
        </p>
        <p class="text-title tabular-nums">-{{ betAmount }}</p>
      </div>
      <div class="text-center pr-default pl-medium">
        <p class="text-caption">
          {{ t('quiz.betting.winnings') }}
        </p>
        <p class="text-title tabular-nums">+{{ resultAmount }}</p>
      </div>
    </div>
    <div class="text-center">
      <p class="text-caption">
        {{
          isWin ? t('quiz.betting.pointsEarned') : t('quiz.betting.pointsLost')
        }}
      </p>
      <p class="text-hero tabular-nums leading-tight">
        {{ isWin ? `+${pointsEarned}` : pointsEarned }}
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
