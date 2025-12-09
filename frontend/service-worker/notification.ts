import type { NotificationType, PushPayload } from './types'

/**
 * Default notification icons based on type
 */
const typeIcons: Record<NotificationType, string> = {
  achievement_unlocked: '/pwa-192x192.png',
  challenge_available: '/pwa-192x192.png',
  generic: '/pwa-192x192.png',
}

/**
 * Get the default icon for a notification type
 */
export function getNotificationIcon(type: NotificationType): string {
  return typeIcons[type] || typeIcons.generic
}

/**
 * Build a URL path based on notification type
 */
export function getNotificationUrl(payload: PushPayload): string {
  if (payload.url) {
    return payload.url
  }

  // Default URL routing based on notification type
  switch (payload.type) {
    case 'achievement_unlocked':
      return `/?achievementId=${payload.data?.achievementId}`
    case 'challenge_available':
      return '/challenges'
    default:
      return '/'
  }
}
