import { describe, it, expect } from 'vitest'
import {
  getProjectAdminIds,
  getChurchAdminIds,
  hasProjectAdminFor,
  hasChurchAdminFor,
  hasRoleType,
  findProjectAdminRole,
  findChurchAdminRole,
  type RoleLike,
} from '../../app/utils/permissions'
import { RoleType } from '../../app/api/generated'

describe('permissions', () => {
  const createRole = (
    role: RoleType,
    scope?: { project?: { id: string } | null; church?: { id: string } | null },
  ): RoleLike => ({
    role,
    scope: scope ?? null,
  })

  describe('getProjectAdminIds', () => {
    it('should return empty array for empty roles', () => {
      expect(getProjectAdminIds([])).toEqual([])
    })

    it('should return project IDs for project admin roles', () => {
      const roles: RoleLike[] = [
        createRole(RoleType.ProjectAdmin, { project: { id: 'proj-1' } }),
        createRole(RoleType.ProjectAdmin, { project: { id: 'proj-2' } }),
      ]

      expect(getProjectAdminIds(roles)).toEqual(['proj-1', 'proj-2'])
    })

    it('should filter out non-project-admin roles', () => {
      const roles: RoleLike[] = [
        createRole(RoleType.ProjectAdmin, { project: { id: 'proj-1' } }),
        createRole(RoleType.ChurchAdmin, { church: { id: 'church-1' } }),
        createRole(RoleType.Admin),
      ]

      expect(getProjectAdminIds(roles)).toEqual(['proj-1'])
    })

    it('should filter out project admin roles without project scope', () => {
      const roles: RoleLike[] = [
        createRole(RoleType.ProjectAdmin, { project: { id: 'proj-1' } }),
        createRole(RoleType.ProjectAdmin, { project: null }),
        createRole(RoleType.ProjectAdmin),
      ]

      expect(getProjectAdminIds(roles)).toEqual(['proj-1'])
    })
  })

  describe('getChurchAdminIds', () => {
    it('should return empty array for empty roles', () => {
      expect(getChurchAdminIds([])).toEqual([])
    })

    it('should return church IDs for church admin roles', () => {
      const roles: RoleLike[] = [
        createRole(RoleType.ChurchAdmin, { church: { id: 'church-1' } }),
        createRole(RoleType.ChurchAdmin, { church: { id: 'church-2' } }),
      ]

      expect(getChurchAdminIds(roles)).toEqual(['church-1', 'church-2'])
    })

    it('should filter out non-church-admin roles', () => {
      const roles: RoleLike[] = [
        createRole(RoleType.ChurchAdmin, { church: { id: 'church-1' } }),
        createRole(RoleType.ProjectAdmin, { project: { id: 'proj-1' } }),
        createRole(RoleType.Admin),
      ]

      expect(getChurchAdminIds(roles)).toEqual(['church-1'])
    })

    it('should filter out church admin roles without church scope', () => {
      const roles: RoleLike[] = [
        createRole(RoleType.ChurchAdmin, { church: { id: 'church-1' } }),
        createRole(RoleType.ChurchAdmin, { church: null }),
        createRole(RoleType.ChurchAdmin),
      ]

      expect(getChurchAdminIds(roles)).toEqual(['church-1'])
    })
  })

  describe('hasProjectAdminFor', () => {
    it('should return false for empty roles', () => {
      expect(hasProjectAdminFor([], 'proj-1')).toBe(false)
    })

    it('should return true if user has project admin role for the project', () => {
      const roles: RoleLike[] = [
        createRole(RoleType.ProjectAdmin, { project: { id: 'proj-1' } }),
      ]

      expect(hasProjectAdminFor(roles, 'proj-1')).toBe(true)
    })

    it('should return false if user has project admin role for a different project', () => {
      const roles: RoleLike[] = [
        createRole(RoleType.ProjectAdmin, { project: { id: 'proj-2' } }),
      ]

      expect(hasProjectAdminFor(roles, 'proj-1')).toBe(false)
    })

    it('should return false for non-project-admin roles', () => {
      const roles: RoleLike[] = [
        createRole(RoleType.Admin),
        createRole(RoleType.ChurchAdmin, { church: { id: 'church-1' } }),
      ]

      expect(hasProjectAdminFor(roles, 'proj-1')).toBe(false)
    })
  })

  describe('hasChurchAdminFor', () => {
    it('should return false for empty roles', () => {
      expect(hasChurchAdminFor([], 'church-1')).toBe(false)
    })

    it('should return true if user has church admin role for the church', () => {
      const roles: RoleLike[] = [
        createRole(RoleType.ChurchAdmin, { church: { id: 'church-1' } }),
      ]

      expect(hasChurchAdminFor(roles, 'church-1')).toBe(true)
    })

    it('should return false if user has church admin role for a different church', () => {
      const roles: RoleLike[] = [
        createRole(RoleType.ChurchAdmin, { church: { id: 'church-2' } }),
      ]

      expect(hasChurchAdminFor(roles, 'church-1')).toBe(false)
    })

    it('should return false for non-church-admin roles', () => {
      const roles: RoleLike[] = [
        createRole(RoleType.Admin),
        createRole(RoleType.ProjectAdmin, { project: { id: 'proj-1' } }),
      ]

      expect(hasChurchAdminFor(roles, 'church-1')).toBe(false)
    })
  })

  describe('hasRoleType', () => {
    it('should return false for empty roles', () => {
      expect(hasRoleType([], RoleType.Admin)).toBe(false)
    })

    it('should return true if user has the role type', () => {
      const roles: RoleLike[] = [createRole(RoleType.Admin)]

      expect(hasRoleType(roles, RoleType.Admin)).toBe(true)
    })

    it('should return false if user does not have the role type', () => {
      const roles: RoleLike[] = [createRole(RoleType.User)]

      expect(hasRoleType(roles, RoleType.Admin)).toBe(false)
    })

    it('should work with multiple roles', () => {
      const roles: RoleLike[] = [
        createRole(RoleType.User),
        createRole(RoleType.ProjectAdmin, { project: { id: 'proj-1' } }),
      ]

      expect(hasRoleType(roles, RoleType.ProjectAdmin)).toBe(true)
      expect(hasRoleType(roles, RoleType.Admin)).toBe(false)
    })
  })

  describe('findProjectAdminRole', () => {
    it('should return undefined for empty roles', () => {
      expect(findProjectAdminRole([], 'proj-1')).toBeUndefined()
    })

    it('should return the role if found', () => {
      const role = createRole(RoleType.ProjectAdmin, {
        project: { id: 'proj-1' },
      })
      const roles: RoleLike[] = [role]

      expect(findProjectAdminRole(roles, 'proj-1')).toBe(role)
    })

    it('should return undefined if not found', () => {
      const roles: RoleLike[] = [
        createRole(RoleType.ProjectAdmin, { project: { id: 'proj-2' } }),
      ]

      expect(findProjectAdminRole(roles, 'proj-1')).toBeUndefined()
    })
  })

  describe('findChurchAdminRole', () => {
    it('should return undefined for empty roles', () => {
      expect(findChurchAdminRole([], 'church-1')).toBeUndefined()
    })

    it('should return the role if found', () => {
      const role = createRole(RoleType.ChurchAdmin, {
        church: { id: 'church-1' },
      })
      const roles: RoleLike[] = [role]

      expect(findChurchAdminRole(roles, 'church-1')).toBe(role)
    })

    it('should return undefined if not found', () => {
      const roles: RoleLike[] = [
        createRole(RoleType.ChurchAdmin, { church: { id: 'church-2' } }),
      ]

      expect(findChurchAdminRole(roles, 'church-1')).toBeUndefined()
    })
  })
})
