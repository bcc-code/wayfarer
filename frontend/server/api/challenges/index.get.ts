import type { Challenge } from '#shared/types'

export default defineEventHandler(async () => {
  return [
    {
      id: '05e9fce2-9b87-4fc2-b537-f9667a288d9f',
      name: 'First Steps',
      description: 'Complete your first journey and log your travel experience',
      publishedAt: '2025-01-15T10:00:00Z',
      userCompletedAt: '2025-01-16T14:30:00Z',
    },
    {
      id: '50238b05-b32d-4cb8-9335-478646e6219f',
      name: 'Globe Trotter',
      description: 'Visit 5 different countries and share your stories',
      publishedAt: '2025-01-20T08:00:00Z',
    },
    {
      id: '373e39f1-6319-4cb4-976f-46db31323b1f',
      name: 'Photo Explorer',
      description: 'Upload 20 travel photos from your adventures',
      publishedAt: '2025-02-01T12:00:00Z',
      userCompletedAt: '2025-02-10T09:15:00Z',
    },
  ] satisfies Challenge[]
})
