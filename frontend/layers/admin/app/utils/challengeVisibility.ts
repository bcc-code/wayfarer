import dayjs from 'dayjs'

/** A `datetime-local` string, `YYYY-MM-DDTHH:mm`, as the admin forms store it. */
type DatetimeLocal = string | undefined

/**
 * Who can see a challenge, the way an organiser thinks about it.
 *
 * The API models this as a single nullable `visibleAt`, where an empty value
 * quietly means "only enrolled users ever see it" — a rule nobody can infer
 * from an empty date picker, and the default a new challenge would otherwise
 * get.
 */
export type ChallengeVisibility = 'everyone' | 'enrolled' | 'scheduled'

export function visibilityFromVisibleAt(
  visibleAt: DatetimeLocal,
  now: Date = new Date(),
): ChallengeVisibility {
  if (!visibleAt) return 'enrolled'

  const parsed = dayjs(visibleAt)
  if (!parsed.isValid()) return 'enrolled'

  return parsed.isAfter(now) ? 'scheduled' : 'everyone'
}

/**
 * The `visibleAt` to store for a chosen visibility. A challenge that is
 * already visible keeps its original timestamp, so saving an unrelated edit
 * does not quietly move the date to today.
 */
export function visibleAtForVisibility(
  visibility: ChallengeVisibility,
  scheduledAt: DatetimeLocal,
  currentVisibleAt: DatetimeLocal,
  now: Date = new Date(),
): DatetimeLocal {
  if (visibility === 'enrolled') return undefined
  if (visibility === 'scheduled') return scheduledAt

  return visibilityFromVisibleAt(currentVisibleAt, now) === 'everyone'
    ? currentVisibleAt
    : dayjs(now).format('YYYY-MM-DDTHH:mm')
}
