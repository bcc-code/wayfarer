export async function hashUserId(userId: string): Promise<string> {
  const encoder = new TextEncoder()
  const data = encoder.encode(userId)
  const hashBuffer = await crypto.subtle.digest('SHA-256', data)
  const hashArray = Array.from(new Uint8Array(hashBuffer))
  return hashArray.map((b) => b.toString(16).padStart(2, '0')).join('')
}

export enum AnalyticsEvent {
  LoginCompleted = 'login_completed',
  LogoutCompleted = 'logout_completed',
  ChallengeLinkClicked = 'challenge_link_clicked',
  LeaderboardTabChanged = 'leaderboard_tab_changed',
  PointsHistoryOpened = 'points_history_opened',
  HowToGetPointsOpened = 'how_to_get_points_opened',
  AchievementClicked = 'achievement_clicked',
  LanguageChanged = 'language_changed',
  ColorModeChanged = 'color_mode_changed',
  ConsentAccepted = 'consent_accepted',
  ConsentRejected = 'consent_rejected',
}

export function getAgeGroup(age?: number | null) {
  if (typeof age != 'number') {
    return 'UNKNOWN'
  }
  if (age < 10) {
    return '< 10'
  } else if (age <= 12) {
    return '10 - 12'
  } else if (age <= 18) {
    return '13 - 18'
  } else if (age <= 25) {
    return '19 - 25'
  } else if (age <= 36) {
    return '26 - 36'
  } else if (age <= 50) {
    return '37 - 50'
  } else if (age <= 64) {
    return '51 - 64'
  } else {
    return '65+'
  }
}
