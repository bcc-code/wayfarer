import type {
  LeaderboardConfigFieldsFragment,
  LeaderboardFilter,
} from '~/api/generated'
import { ChurchCategory, Gender, LeaderboardEntityType } from '~/api/generated'

type FilterView = LeaderboardConfigFieldsFragment['filter']

export const LEADERBOARD_ENTITY_TYPE_LABELS: Record<
  LeaderboardEntityType,
  string
> = {
  [LeaderboardEntityType.Persons]: 'Personer',
  [LeaderboardEntityType.Teams]: 'Lag',
  [LeaderboardEntityType.Superteams]: 'Superlag',
  [LeaderboardEntityType.Churches]: 'Menigheter',
}

export const LEADERBOARD_ENTITY_TYPE_ITEMS = Object.values(
  LeaderboardEntityType,
).map((value) => ({ label: LEADERBOARD_ENTITY_TYPE_LABELS[value], value }))

export const CHURCH_CATEGORY_ITEMS = Object.values(ChurchCategory).map(
  (value) => ({ label: value.toUpperCase(), value }),
)

export const GENDER_LABELS: Record<Gender, string> = {
  [Gender.Male]: 'Gutt',
  [Gender.Female]: 'Jente',
}

export const GENDER_ITEMS = Object.values(Gender).map((value) => ({
  label: GENDER_LABELS[value],
  value,
}))

/**
 * `LeaderboardConfig.filter` is a `LeaderboardFilterView` (output type); the
 * create/update mutations take a `LeaderboardFilter` (input type). The fields
 * are identical except `ageRange`, which is `AgeRange` out and `AgeRangeInput`
 * in — so this is a field-by-field copy that also drops the `__typename`s
 * urql adds, which the server rejects on an input object.
 *
 * Needed wherever a config is written back without going through the form:
 * `UpdateLeaderboardConfigInput` is full-replace, so a reorder still has to
 * resend the filter it is not changing.
 */
export function leaderboardFilterViewToInput(
  view: FilterView,
): LeaderboardFilter | null {
  if (!view) return null

  const filter: LeaderboardFilter = {}

  if (view.minScore != null) filter.minScore = view.minScore
  if (view.maxScore != null) filter.maxScore = view.maxScore
  if (view.churchId) filter.churchId = view.churchId
  if (view.country) filter.country = view.country
  if (view.churchCategory) filter.churchCategory = view.churchCategory
  if (view.gender) filter.gender = view.gender
  if (view.ageRange) {
    filter.ageRange = { min: view.ageRange.min, max: view.ageRange.max }
  }
  if (view.teamId) filter.teamId = view.teamId
  if (view.superTeamId) filter.superTeamId = view.superTeamId

  return Object.keys(filter).length ? filter : null
}

/**
 * Short chips describing a config's filter, for the admin list.
 *
 * The ID-valued fields name the dimension rather than the thing — a
 * `LeaderboardFilterView` carries `teamId`, not the team — and resolving those
 * would cost a lookup per row for a label nobody reads twice.
 */
export function summarizeLeaderboardFilter(view: FilterView): string[] {
  if (!view) return []

  const parts: string[] = []

  if (view.ageRange) parts.push(`${view.ageRange.min}–${view.ageRange.max} år`)
  if (view.gender) parts.push(GENDER_LABELS[view.gender])
  if (view.churchCategory)
    parts.push(`Størrelse ${view.churchCategory.toUpperCase()}`)
  if (view.country) parts.push(view.country)
  if (view.churchId) parts.push('Én menighet')
  if (view.teamId) parts.push('Ett lag')
  if (view.superTeamId) parts.push('Ett superlag')
  if (view.minScore != null) parts.push(`Fra ${view.minScore} p`)
  if (view.maxScore != null) parts.push(`Til ${view.maxScore} p`)

  return parts
}
