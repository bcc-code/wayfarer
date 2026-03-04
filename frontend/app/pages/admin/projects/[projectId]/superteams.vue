<script setup lang="ts">
import { VisAxis, VisGroupedBar, VisXYContainer } from '@unovis/vue'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

const route = useRoute('admin-projects-projectId-superteams')
const config = useRuntimeConfig()
const toast = useToast()

// Response types matching backend
interface TeamInfo {
  team_id: string
  team_name: string
  church_id: string
  church_name: string
  total_score: number
  member_count: number
}

interface SuperteamResult {
  super_team_id: string
  name: string
  total_score: number
  team_count: number
  member_count: number
  teams: TeamInfo[]
  churches: string[]
}

interface DistributionResponse {
  superteams: SuperteamResult[]
  variance: number
}

// Color mapping for superteams
const superteamColors: Record<string, string> = {
  Purple: '#a855f7',
  Green: '#22c55e',
  Red: '#ef4444',
  Yellow: '#eab308',
}

// State
const previewData = ref<DistributionResponse | null>(null)
const isPreviewLoading = ref(false)
const isDistributeLoading = ref(false)
const error = ref<string | null>(null)

// Get auth token
const wayfarerToken = useLocalStorage<string>('token', '')

// API helpers
async function makeRequest<T>(
  endpoint: string,
  method: 'GET' | 'POST',
  body?: Record<string, unknown>,
): Promise<T> {
  // apiUrl includes /graphql suffix, so we need to strip it for REST endpoints
  const baseUrl = (config.public.apiUrl as string).replace(/\/graphql$/, '')
  const url = `${baseUrl}${endpoint}`
  const options: RequestInit = {
    method,
    headers: {
      Authorization: `Bearer ${wayfarerToken.value}`,
      'Content-Type': 'application/json',
    },
  }

  if (body) {
    options.body = JSON.stringify(body)
  }

  console.log('Making request to:', url) // Debug log
  const response = await fetch(url, options)

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}))
    console.error('Request failed:', response.status, errorData) // Debug log
    throw new Error(
      (errorData as { error?: string }).error || `Request failed: ${response.status}`,
    )
  }

  return response.json()
}

// Preview distribution
async function previewDistribution() {
  isPreviewLoading.value = true
  error.value = null

  try {
    const data = await makeRequest<DistributionResponse>(
      `/plugins/ladder-to-heaven/preview-superteams?project_id=${route.params.projectId}`,
      'GET',
    )
    previewData.value = data
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to preview distribution'
    toast.add({
      title: 'Preview failed',
      description: error.value,
      color: 'error',
    })
  } finally {
    isPreviewLoading.value = false
  }
}

// Execute distribution
async function executeDistribution() {
  isDistributeLoading.value = true
  error.value = null

  try {
    const data = await makeRequest<DistributionResponse>(
      '/plugins/ladder-to-heaven/distribute-superteams',
      'POST',
      { project_id: route.params.projectId },
    )
    previewData.value = data

    toast.add({
      title: 'Distribution complete',
      description: `Created ${data.superteams.length} superteams`,
      color: 'success',
    })
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to execute distribution'
    toast.add({
      title: 'Distribution failed',
      description: error.value,
      color: 'error',
    })
  } finally {
    isDistributeLoading.value = false
  }
}

// Chart data
const chartData = computed(() => {
  if (!previewData.value) return []
  return previewData.value.superteams.map((st, i) => ({
    index: i,
    name: st.name,
    score: st.total_score,
  }))
})

type ChartDataItem = (typeof chartData.value)[number]

const barX = (_d: ChartDataItem, i: number) => i
const barY = (d: ChartDataItem) => d.score
const barColor = (d: ChartDataItem) => superteamColors[d.name] ?? '#888'

// Calculate balance metric
const balanceMetric = computed(() => {
  if (!previewData.value || previewData.value.variance === 0) return 'Perfect'
  const variance = previewData.value.variance
  const stdDev = Math.sqrt(variance)
  return `Std Dev: ${formatNumber(Math.round(stdDev))}`
})

// Expanded state for team tables
const expandedSuperteams = ref<Set<string>>(new Set())

function toggleExpanded(name: string) {
  if (expandedSuperteams.value.has(name)) {
    expandedSuperteams.value.delete(name)
  } else {
    expandedSuperteams.value.add(name)
  }
}

// Church distribution analysis
interface ChurchInfo {
  churchId: string
  label: string
  color: string
  superteams: string[]
  teamCount: number
  totalScore: number
}

const churchColors = [
  '#8b5cf6', // violet
  '#06b6d4', // cyan
  '#f97316', // orange
  '#ec4899', // pink
  '#14b8a6', // teal
  '#f59e0b', // amber
  '#6366f1', // indigo
  '#84cc16', // lime
  '#e11d48', // rose
  '#0ea5e9', // sky
  '#a855f7', // purple
  '#10b981', // emerald
]

const churchAnalysis = computed((): ChurchInfo[] => {
  if (!previewData.value) return []

  const churchMap = new Map<string, ChurchInfo>()
  let colorIndex = 0

  for (const st of previewData.value.superteams) {
    for (const team of st.teams) {
      const churchId = team.church_id || 'Unknown'
      const churchName = team.church_name || churchId
      if (!churchMap.has(churchId)) {
        churchMap.set(churchId, {
          churchId,
          label: churchName || 'Unknown',
          color: churchColors[colorIndex % churchColors.length],
          superteams: [],
          teamCount: 0,
          totalScore: 0,
        })
        colorIndex++
      }
      const info = churchMap.get(churchId)!
      info.teamCount++
      info.totalScore += team.total_score
      if (!info.superteams.includes(st.name)) {
        info.superteams.push(st.name)
      }
    }
  }

  return Array.from(churchMap.values()).sort((a, b) => b.totalScore - a.totalScore)
})

// Get church color by ID
function getChurchColor(churchId: string): string {
  const info = churchAnalysis.value.find(c => c.churchId === churchId)
  return info?.color ?? '#888'
}

// Get teams grouped by church for a superteam
function getTeamsByChurch(st: SuperteamResult): Map<string, TeamInfo[]> {
  const grouped = new Map<string, TeamInfo[]>()
  for (const team of st.teams) {
    const churchId = team.church_id || 'Unknown'
    if (!grouped.has(churchId)) {
      grouped.set(churchId, [])
    }
    grouped.get(churchId)!.push(team)
  }
  return grouped
}
</script>

<template>
  <div>
    <div class="border-default border-b py-2">
      <UContainer>
        <UBreadcrumb
          :items="[
            {
              label: 'Prosjekter',
              to: { name: 'admin-projects' },
            },
            {
              label: route.params.projectId,
              to: {
                name: 'admin-projects-projectId',
                params: { projectId: route.params.projectId },
              },
            },
            {
              label: 'Superteams Distribution',
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <header class="mb-12">
        <h1 class="text-3xl font-semibold">LADD Superteams</h1>
        <p class="text-muted mt-2 max-w-2xl">
          Distribute teams with scores into 4 balanced superteams (Purple, Green,
          Red, Yellow). Teams from the same church are kept together.
        </p>
      </header>

      <!-- Action buttons -->
      <div class="mb-8 flex gap-4">
        <UButton
          :loading="isPreviewLoading"
          :disabled="isDistributeLoading"
          variant="soft"
          @click="previewDistribution"
        >
          <Icon name="lucide:eye" class="mr-2" />
          Preview Distribution
        </UButton>
        <UButton
          :loading="isDistributeLoading"
          :disabled="isPreviewLoading"
          color="primary"
          @click="executeDistribution"
        >
          <Icon name="lucide:play" class="mr-2" />
          Execute Distribution
        </UButton>
      </div>

      <!-- Error state -->
      <UAlert v-if="error" color="error" class="mb-6">
        <template #title>Error</template>
        {{ error }}
      </UAlert>

      <!-- Results -->
      <div v-if="previewData" class="space-y-8">
        <!-- Summary cards -->
        <div class="grid grid-cols-2 gap-4 md:grid-cols-4">
          <div
            v-for="st in previewData.superteams"
            :key="st.name"
            class="border-default rounded-lg border p-4"
            :style="{ borderLeftColor: superteamColors[st.name], borderLeftWidth: '4px' }"
          >
            <div class="text-lg font-semibold">{{ st.name }}</div>
            <div class="text-muted text-sm">
              {{ st.team_count }} teams, {{ st.member_count }} members
            </div>
            <div class="mt-2 text-2xl font-bold">
              {{ formatNumber(st.total_score) }}
            </div>
          </div>
        </div>

        <!-- Balance metric -->
        <div class="text-muted">
          Balance: {{ balanceMetric }}
        </div>

        <!-- Chart -->
        <div class="border-default rounded-lg border p-4">
          <h2 class="mb-4 text-xl font-semibold">Score Distribution</h2>
          <VisXYContainer :data="chartData" :height="300">
            <VisGroupedBar
              :x="barX"
              :y="barY"
              :color="barColor"
              :rounded-corners="4"
            />
            <VisAxis type="y" :label="'Total Score'" />
            <VisAxis
              type="x"
              :tick-format="(i: number) => chartData[i]?.name ?? ''"
            />
          </VisXYContainer>
        </div>

        <!-- Church Distribution Visualization -->
        <div class="border-default rounded-lg border p-4">
          <h2 class="mb-4 text-xl font-semibold">Church Distribution</h2>

          <!-- Church Legend -->
          <div class="mb-6 flex flex-wrap gap-2">
            <div
              v-for="church in churchAnalysis"
              :key="church.churchId"
              class="flex items-center gap-1.5 rounded-full px-2 py-1 text-xs"
              :style="{ backgroundColor: church.color + '20', color: church.color }"
            >
              <div
                class="size-2.5 rounded-full"
                :style="{ backgroundColor: church.color }"
              />
              <span class="font-medium">{{ church.label }}</span>
              <span class="opacity-70">({{ church.teamCount }})</span>
            </div>
          </div>

          <!-- Visual bars for each superteam -->
          <div class="space-y-4">
            <div
              v-for="st in previewData.superteams"
              :key="st.name"
              class="flex items-center gap-4"
            >
              <div class="w-20 font-medium" :style="{ color: superteamColors[st.name] }">
                {{ st.name }}
              </div>
              <div class="flex flex-1 gap-0.5 overflow-hidden rounded">
                <div
                  v-for="(teams, churchId) in Object.fromEntries(getTeamsByChurch(st))"
                  :key="churchId"
                  class="flex"
                >
                  <div
                    v-for="team in teams"
                    :key="team.team_id"
                    class="h-8 w-6 border-r border-white/20 transition-all hover:scale-110 hover:z-10"
                    :style="{ backgroundColor: getChurchColor(churchId) }"
                    :title="`${team.team_name}\n${team.church_name || churchId}\nScore: ${formatNumber(team.total_score)}`"
                  />
                </div>
              </div>
              <div class="w-24 text-right text-sm text-gray-500">
                {{ st.team_count }} teams
              </div>
            </div>
          </div>

          <!-- Church split summary -->
          <div class="mt-6 border-t pt-4">
            <h3 class="mb-3 text-sm font-medium text-gray-500">Church Assignment Summary</h3>
            <div class="grid gap-2 md:grid-cols-2 lg:grid-cols-3">
              <div
                v-for="church in churchAnalysis"
                :key="church.churchId"
                class="flex items-center justify-between rounded-lg px-3 py-2"
                :class="church.superteams.length > 1 ? 'bg-amber-50 dark:bg-amber-900/20' : 'bg-green-50 dark:bg-green-900/20'"
              >
                <div class="flex items-center gap-2">
                  <div
                    class="size-3 rounded-full"
                    :style="{ backgroundColor: church.color }"
                  />
                  <span class="font-medium">{{ church.label }}</span>
                </div>
                <div class="flex items-center gap-2 text-sm">
                  <span class="text-gray-500">{{ church.teamCount }} teams in</span>
                  <div class="flex gap-0.5">
                    <div
                      v-for="stName in church.superteams"
                      :key="stName"
                      class="size-4 rounded"
                      :style="{ backgroundColor: superteamColors[stName] }"
                      :title="stName"
                    />
                  </div>
                  <span
                    class="rounded px-1.5 py-0.5 text-xs font-medium"
                    :class="church.superteams.length > 1 ? 'bg-amber-200 text-amber-800 dark:bg-amber-800 dark:text-amber-200' : 'bg-green-200 text-green-800 dark:bg-green-800 dark:text-green-200'"
                  >
                    {{ church.superteams.length > 1 ? 'SPLIT' : 'TOGETHER' }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Team tables -->
        <div class="space-y-4">
          <h2 class="text-xl font-semibold">Teams per Superteam</h2>
          <div
            v-for="st in previewData.superteams"
            :key="st.name"
            class="border-default rounded-lg border"
          >
            <button
              class="flex w-full items-center justify-between p-4 text-left hover:bg-gray-50 dark:hover:bg-gray-800"
              @click="toggleExpanded(st.name)"
            >
              <div class="flex items-center gap-3">
                <div
                  class="size-4 rounded"
                  :style="{ backgroundColor: superteamColors[st.name] }"
                />
                <span class="font-semibold">{{ st.name }}</span>
                <span class="text-muted text-sm">
                  ({{ st.team_count }} teams, {{ formatNumber(st.total_score) }}
                  points)
                </span>
              </div>
              <Icon
                :name="
                  expandedSuperteams.has(st.name)
                    ? 'lucide:chevron-up'
                    : 'lucide:chevron-down'
                "
              />
            </button>

            <div v-if="expandedSuperteams.has(st.name)" class="border-t">
              <table class="w-full">
                <thead class="bg-gray-50 dark:bg-gray-800">
                  <tr>
                    <th
                      class="px-4 py-2 text-left text-sm font-medium text-gray-500"
                    >
                      Team
                    </th>
                    <th
                      class="px-4 py-2 text-left text-sm font-medium text-gray-500"
                    >
                      Church
                    </th>
                    <th
                      class="px-4 py-2 text-right text-sm font-medium text-gray-500"
                    >
                      Members
                    </th>
                    <th
                      class="px-4 py-2 text-right text-sm font-medium text-gray-500"
                    >
                      Score
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="team in st.teams"
                    :key="team.team_id"
                    class="border-t"
                  >
                    <td class="px-4 py-2">{{ team.team_name }}</td>
                    <td class="text-muted px-4 py-2 text-sm">
                      {{ team.church_name || team.church_id || '-' }}
                    </td>
                    <td class="px-4 py-2 text-right">{{ team.member_count }}</td>
                    <td class="px-4 py-2 text-right font-medium">
                      {{ formatNumber(team.total_score) }}
                    </td>
                  </tr>
                  <tr v-if="st.teams.length === 0">
                    <td
                      colspan="4"
                      class="text-muted px-4 py-4 text-center text-sm"
                    >
                      No teams assigned
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>

      <!-- Empty state -->
      <div
        v-else-if="!isPreviewLoading && !error"
        class="text-muted rounded-lg border border-dashed p-12 text-center"
      >
        <Icon name="lucide:users" class="mx-auto mb-4 size-12 opacity-50" />
        <p class="text-lg">Click "Preview Distribution" to see the proposed superteam assignments</p>
      </div>
    </UContainer>
  </div>
</template>

<style scoped>
.dark #chart {
  --vis-axis-grid-color: #333;
}
</style>
