import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { ref } from 'vue'
import useGroupedProjects from '../../app/composables/useGroupedProjects'

type Project = { id: string; startDate: string; endDate: string }

describe('useGroupedProjects', () => {
  beforeEach(() => {
    // Mock current date to 2024-06-15 at noon
    // Using noon avoids timezone/midnight boundary issues
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2024-06-15T12:00:00'))
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  describe('initialization', () => {
    it('should handle undefined projects', () => {
      const grouped = useGroupedProjects(undefined)

      expect(grouped.currentProjects.value).toEqual([])
      expect(grouped.pastProjects.value).toEqual([])
      expect(grouped.futureProjects.value).toEqual([])
    })

    it('should handle empty array', () => {
      const grouped = useGroupedProjects([])

      expect(grouped.currentProjects.value).toEqual([])
      expect(grouped.pastProjects.value).toEqual([])
      expect(grouped.futureProjects.value).toEqual([])
    })
  })

  describe('current projects', () => {
    it('should identify projects that are currently active', () => {
      const projects: Project[] = [
        {
          id: '1',
          startDate: '2024-06-01',
          endDate: '2024-06-30',
        },
        {
          id: '2',
          startDate: '2024-06-10',
          endDate: '2024-06-20',
        },
      ]

      const grouped = useGroupedProjects(projects)

      expect(grouped.currentProjects.value).toHaveLength(2)
      expect(grouped.currentProjects.value.map((p) => p.id)).toEqual([
        '1',
        '2',
      ])
    })

    it('should include projects that started today', () => {
      const projects: Project[] = [
        {
          id: '1',
          startDate: '2024-06-15',
          endDate: '2024-06-30',
        },
      ]

      const grouped = useGroupedProjects(projects)

      expect(grouped.currentProjects.value).toHaveLength(1)
      expect(grouped.currentProjects.value[0].id).toBe('1')
    })

    it('should not include projects that ended at midnight today when now is later', () => {
      // Project ends at '2024-06-15' (midnight), but now is noon
      // So noon > midnight means project has ended
      const projects: Project[] = [
        {
          id: '1',
          startDate: '2024-06-01',
          endDate: '2024-06-15',
        },
      ]

      const grouped = useGroupedProjects(projects)

      expect(grouped.currentProjects.value).toHaveLength(0)
      expect(grouped.pastProjects.value).toHaveLength(1)
    })
  })

  describe('past projects', () => {
    it('should identify projects that have ended', () => {
      const projects: Project[] = [
        {
          id: '1',
          startDate: '2024-05-01',
          endDate: '2024-05-31',
        },
        {
          id: '2',
          startDate: '2024-04-01',
          endDate: '2024-04-30',
        },
      ]

      const grouped = useGroupedProjects(projects)

      expect(grouped.pastProjects.value).toHaveLength(2)
      expect(grouped.pastProjects.value.map((p) => p.id)).toEqual(['1', '2'])
    })

    it('should include projects that ended at midnight today', () => {
      // Project ends at '2024-06-15' (midnight), now is noon
      // So this is a past project
      const projects: Project[] = [
        {
          id: '1',
          startDate: '2024-06-01',
          endDate: '2024-06-15',
        },
      ]

      const grouped = useGroupedProjects(projects)

      expect(grouped.pastProjects.value).toHaveLength(1)
      expect(grouped.pastProjects.value[0].id).toBe('1')
    })

    it('should not include current projects in past', () => {
      const projects: Project[] = [
        {
          id: '1',
          startDate: '2024-06-01',
          endDate: '2024-06-30',
        },
        {
          id: '2',
          startDate: '2024-05-01',
          endDate: '2024-05-31',
        },
      ]

      const grouped = useGroupedProjects(projects)

      expect(grouped.pastProjects.value).toHaveLength(1)
      expect(grouped.pastProjects.value[0].id).toBe('2')
    })
  })

  describe('future projects', () => {
    it('should identify projects that have not started yet', () => {
      const projects: Project[] = [
        {
          id: '1',
          startDate: '2024-07-01',
          endDate: '2024-07-31',
        },
        {
          id: '2',
          startDate: '2024-08-01',
          endDate: '2024-08-31',
        },
      ]

      const grouped = useGroupedProjects(projects)

      expect(grouped.futureProjects.value).toHaveLength(2)
      expect(grouped.futureProjects.value.map((p) => p.id)).toEqual(['1', '2'])
    })

    it('should not include projects that started at midnight today', () => {
      // Project starts at '2024-06-15' (midnight), now is noon
      // So this is a current project, not future
      const projects: Project[] = [
        {
          id: '1',
          startDate: '2024-06-15',
          endDate: '2024-06-30',
        },
      ]

      const grouped = useGroupedProjects(projects)

      expect(grouped.futureProjects.value).toHaveLength(0)
      expect(grouped.currentProjects.value).toHaveLength(1)
    })

    it('should not include current projects in future', () => {
      const projects: Project[] = [
        {
          id: '1',
          startDate: '2024-06-01',
          endDate: '2024-06-30',
        },
        {
          id: '2',
          startDate: '2024-07-01',
          endDate: '2024-07-31',
        },
      ]

      const grouped = useGroupedProjects(projects)

      expect(grouped.futureProjects.value).toHaveLength(1)
      expect(grouped.futureProjects.value[0].id).toBe('2')
    })
  })

  describe('mixed projects', () => {
    it('should correctly categorize a mix of past, current, and future projects', () => {
      const projects: Project[] = [
        { id: 'past1', startDate: '2024-04-01', endDate: '2024-04-30' },
        { id: 'past2', startDate: '2024-05-01', endDate: '2024-05-31' },
        { id: 'current1', startDate: '2024-06-01', endDate: '2024-06-30' },
        { id: 'current2', startDate: '2024-06-10', endDate: '2024-06-20' },
        { id: 'future1', startDate: '2024-07-01', endDate: '2024-07-31' },
        { id: 'future2', startDate: '2024-08-01', endDate: '2024-08-31' },
      ]

      const grouped = useGroupedProjects(projects)

      expect(grouped.pastProjects.value.map((p) => p.id)).toEqual([
        'past1',
        'past2',
      ])
      expect(grouped.currentProjects.value.map((p) => p.id)).toEqual([
        'current1',
        'current2',
      ])
      expect(grouped.futureProjects.value.map((p) => p.id)).toEqual([
        'future1',
        'future2',
      ])
    })

    it('should ensure each project appears in exactly one category', () => {
      const projects: Project[] = [
        { id: '1', startDate: '2024-04-01', endDate: '2024-04-30' },
        { id: '2', startDate: '2024-06-01', endDate: '2024-06-30' },
        { id: '3', startDate: '2024-07-01', endDate: '2024-07-31' },
      ]

      const grouped = useGroupedProjects(projects)

      const allIds = [
        ...grouped.pastProjects.value.map((p) => p.id),
        ...grouped.currentProjects.value.map((p) => p.id),
        ...grouped.futureProjects.value.map((p) => p.id),
      ]

      // Check no duplicates
      expect(allIds).toEqual(['1', '2', '3'])
      expect(new Set(allIds).size).toBe(3)
    })
  })

  describe('reactivity', () => {
    it('should work with a ref', () => {
      const projectsRef = ref<Project[]>([
        { id: '1', startDate: '2024-06-01', endDate: '2024-06-30' },
      ])

      const grouped = useGroupedProjects(projectsRef)

      expect(grouped.currentProjects.value).toHaveLength(1)

      // Update the ref
      projectsRef.value = [
        { id: '1', startDate: '2024-06-01', endDate: '2024-06-30' },
        { id: '2', startDate: '2024-07-01', endDate: '2024-07-31' },
      ]

      expect(grouped.currentProjects.value).toHaveLength(1)
      expect(grouped.futureProjects.value).toHaveLength(1)
    })

    it('should work with a getter function', () => {
      let projects: Project[] = [
        { id: '1', startDate: '2024-06-01', endDate: '2024-06-30' },
      ]

      const grouped = useGroupedProjects(() => projects)

      expect(grouped.currentProjects.value).toHaveLength(1)

      // Update the projects
      projects = [
        { id: '1', startDate: '2024-06-01', endDate: '2024-06-30' },
        { id: '2', startDate: '2024-05-01', endDate: '2024-05-31' },
      ]

      expect(grouped.currentProjects.value).toHaveLength(1)
      expect(grouped.pastProjects.value).toHaveLength(1)
    })

    it('should react to changes from undefined to populated', () => {
      const projectsRef = ref<Project[] | undefined>(undefined)
      const grouped = useGroupedProjects(projectsRef)

      expect(grouped.currentProjects.value).toEqual([])

      projectsRef.value = [
        { id: '1', startDate: '2024-06-01', endDate: '2024-06-30' },
      ]

      expect(grouped.currentProjects.value).toHaveLength(1)
    })
  })

  describe('edge cases', () => {
    it('should handle projects with same start and end date at midnight', () => {
      // Project is only valid at '2024-06-15' midnight, now is noon
      // So this project has already ended
      const projects: Project[] = [
        { id: '1', startDate: '2024-06-15', endDate: '2024-06-15' },
      ]

      const grouped = useGroupedProjects(projects)

      expect(grouped.pastProjects.value).toHaveLength(1)
      expect(grouped.pastProjects.value[0].id).toBe('1')
    })

    it('should handle projects active for the entire day', () => {
      // Project runs for the whole day
      const projects: Project[] = [
        {
          id: '1',
          startDate: '2024-06-15T00:00:00',
          endDate: '2024-06-15T23:59:59',
        },
      ]

      const grouped = useGroupedProjects(projects)

      expect(grouped.currentProjects.value).toHaveLength(1)
      expect(grouped.currentProjects.value[0].id).toBe('1')
    })

    it('should handle projects spanning multiple years', () => {
      const projects: Project[] = [
        { id: '1', startDate: '2023-01-01', endDate: '2025-12-31' },
      ]

      const grouped = useGroupedProjects(projects)

      expect(grouped.currentProjects.value).toHaveLength(1)
    })

    it('should handle date strings with time components', () => {
      const projects: Project[] = [
        {
          id: '1',
          startDate: '2024-06-01T00:00:00Z',
          endDate: '2024-06-30T23:59:59Z',
        },
      ]

      const grouped = useGroupedProjects(projects)

      expect(grouped.currentProjects.value).toHaveLength(1)
    })
  })

  describe('custom types', () => {
    it('should work with objects that have additional properties', () => {
      type ExtendedProject = {
        id: string
        name: string
        startDate: string
        endDate: string
        description: string
      }

      const projects: ExtendedProject[] = [
        {
          id: '1',
          name: 'Project A',
          startDate: '2024-06-01',
          endDate: '2024-06-30',
          description: 'Current project',
        },
        {
          id: '2',
          name: 'Project B',
          startDate: '2024-05-01',
          endDate: '2024-05-31',
          description: 'Past project',
        },
      ]

      const grouped = useGroupedProjects(projects)

      expect(grouped.currentProjects.value[0].name).toBe('Project A')
      expect(grouped.pastProjects.value[0].description).toBe('Past project')
    })
  })
})
