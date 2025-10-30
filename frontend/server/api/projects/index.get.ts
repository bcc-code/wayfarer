import type { Project } from '#shared/types'

export default defineEventHandler(async () => {
  return [
    {
      id: '8f14de93-43ad-4292-b237-a2646ed1cc7b',
      name: 'Summer Camp 2025',
      description:
        'Annual summer camp program with multiple week-long sessions featuring outdoor activities, workshops, and team building exercises',
      logo: 'https://placehold.co/64/f59e0b/white?text=SC',
      color: '#f59e0b',
      startDate: '2025-06-15',
      endDate: '2025-08-20',
    },
    {
      id: '42b5c9d2-843b-4089-8c5b-278c58019b91',
      name: 'Easter Retreat',
      description:
        'Spring break retreat with worship sessions, community service projects, and recreational activities for youth and families',
      logo: 'https://placehold.co/64/8b5cf6/white?text=ER',
      color: '#8b5cf6',
      startDate: '2025-04-17',
      endDate: '2025-04-21',
    },
    {
      id: '4a03dd42-b192-4db3-9fa5-de3e4a4d8acb',
      name: 'New Years Conference',
      description:
        'Multi-day conference celebrating the new year with keynote speakers, breakout sessions, and celebration events',
      logo: 'https://placehold.co/64/3b82f6/white?text=NY',
      color: '#3b82f6',
      startDate: '2025-12-28',
      endDate: '2026-01-02',
    },
    {
      id: 'a7c8e3f9-2b4d-4e8a-9c3d-5f7e1a2b3c4d',
      name: 'Fall Leadership Retreat',
      logo: 'https://placehold.co/64/f97316/white?text=FL',
      color: '#f97316',
      startDate: '2025-10-10',
      endDate: '2025-10-15',
    },
    {
      id: 'b9d2f4e6-3c5a-4f7b-8d9e-6a1c2b3d4e5f',
      name: 'Winter Sports Camp',
      description:
        'Action-packed winter camp featuring skiing, snowboarding lessons, and indoor activities for all skill levels',
      startDate: '2026-01-25',
      endDate: '2026-02-01',
    },
  ] as Project[]
})
