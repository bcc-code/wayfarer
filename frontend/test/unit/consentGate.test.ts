import { describe, it, expect } from 'vitest'
import { ConsentManagementType } from '../../app/api/generated'
import {
  blockingConsentIds,
  unansweredLocalConsents,
} from '../../app/utils/consentGate'

/**
 * The one rule deciding whether the consent screen must be shown. Both the
 * route middleware and the consent page read it, so the edge cases are pinned
 * here rather than once per caller.
 */

function consent(
  id: string,
  isRequired: boolean,
  managementType = ConsentManagementType.Local,
) {
  return { id, isRequired, managementType }
}

const REMOTE = ConsentManagementType.Remote

function status(
  pending: ReturnType<typeof consent>[] = [],
  rejected: ReturnType<typeof consent>[] = [],
) {
  return {
    pendingConsents: pending,
    rejectedConsents: rejected.map((c) => ({ consent: c })),
  }
}

describe('blockingConsentIds', () => {
  it('is empty when there is no status yet', () => {
    expect(blockingConsentIds(undefined).size).toBe(0)
  })

  it('blocks on a required consent that was never answered', () => {
    expect([...blockingConsentIds(status([consent('CN1', true)]))]).toEqual([
      'CN1',
    ])
  })

  it('blocks on a required consent that was REJECTED', () => {
    // The case `pendingConsents` cannot express: the user acted, so it is no
    // longer pending, but it is still not accepted.
    expect([...blockingConsentIds(status([], [consent('CN1', true)]))]).toEqual(
      ['CN1'],
    )
  })

  it('does not block on an accepted one', () => {
    // An accepted consent appears in neither list the gate reads.
    expect(blockingConsentIds(status()).size).toBe(0)
  })

  it('does not block on optional consents, answered either way', () => {
    const s = status([consent('CN1', false)], [consent('CN2', false)])
    expect(blockingConsentIds(s).size).toBe(0)
  })

  it('does not block on a required REMOTE consent', () => {
    // It cannot be accepted in the app, so gating on it would strand the user.
    const s = status(
      [consent('CN1', true, REMOTE)],
      [consent('CN2', true, REMOTE)],
    )
    expect(blockingConsentIds(s).size).toBe(0)
  })

  it('collects every blocking consent, from both lists', () => {
    const s = status(
      [consent('CN1', true), consent('CN2', false)],
      [consent('CN3', true), consent('CN4', false)],
    )
    expect([...blockingConsentIds(s)].sort()).toEqual(['CN1', 'CN3'])
  })
})

describe('unansweredLocalConsents', () => {
  it('is empty when there is no status yet', () => {
    expect(unansweredLocalConsents(undefined)).toEqual([])
  })

  it('returns pending local consents whether required or not', () => {
    const s = status([consent('CN1', true), consent('CN2', false)])
    expect(unansweredLocalConsents(s).map((c) => c.id)).toEqual(['CN1', 'CN2'])
  })

  it('excludes remote consents, which are answered elsewhere', () => {
    const s = status([consent('CN1', false, REMOTE)])
    expect(unansweredLocalConsents(s)).toEqual([])
  })

  it('ignores rejected consents — answered is answered', () => {
    const s = status([], [consent('CN1', false)])
    expect(unansweredLocalConsents(s)).toEqual([])
  })
})
