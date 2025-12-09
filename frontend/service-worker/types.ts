/**
 * Push notification payload sent from the server
 */
export interface PushPayload {
  /** Notification title */
  title: string
  /** Notification body text */
  body: string
  /** Icon URL for the notification */
  icon?: string
  /** Badge URL (small monochrome icon) */
  badge?: string
  /** Image URL to display in the notification */
  image?: string
  /** Unique notification ID */
  notificationId: string
  /** Type of notification for routing purposes */
  type: NotificationType
  /** URL to open when notification is clicked */
  url?: string
  /** Additional data to pass through */
  data?: Record<string, unknown>
  /** Timestamp of the notification */
  timestamp?: number
  /** Tag for grouping/replacing notifications */
  tag?: string
}

/**
 * Types of notifications the app can send
 */
export type NotificationType =
  | 'achievement_unlocked'
  | 'challenge_available'
  | 'generic'

/**
 * Data attached to the notification for click handling
 */
export interface NotificationData {
  /** URL to navigate to */
  url: string
  /** Notification ID for tracking */
  notificationId: string
  /** Notification type */
  type: NotificationType
  /** Any additional data */
  [key: string]: unknown
}
