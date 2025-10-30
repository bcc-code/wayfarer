export interface Challenge {
  id: string
  name: string
  description: string
  image?: string
  url?: string
  publishedAt?: string
  userCompletedAt?: string
}

export interface Project {
  id: string
  name: string
  description?: string
  logo?: string
  color?: string
  startDate: string
  endDate: string
}
