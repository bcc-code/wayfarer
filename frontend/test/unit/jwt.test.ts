import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { isTokenExpired } from '../../app/utils/jwt'

function makeToken(payload: Record<string, unknown>): string {
  const encode = (obj: Record<string, unknown>) =>
    Buffer.from(JSON.stringify(obj))
      .toString('base64')
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=+$/, '')
  return `${encode({ alg: 'HS256', typ: 'JWT' })}.${encode(payload)}.signature`
}

describe('isTokenExpired', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2024-06-15T12:00:00'))
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  const nowSeconds = () => Math.floor(Date.now() / 1000)

  it('returns false for a token expiring well in the future', () => {
    const token = makeToken({ exp: nowSeconds() + 24 * 60 * 60 })
    expect(isTokenExpired(token)).toBe(false)
  })

  it('returns true for a token that expired in the past', () => {
    const token = makeToken({ exp: nowSeconds() - 60 })
    expect(isTokenExpired(token)).toBe(true)
  })

  it('returns true for a token expiring within the clock-skew margin', () => {
    const token = makeToken({ exp: nowSeconds() + 10 })
    expect(isTokenExpired(token)).toBe(true)
  })

  it('returns true when the exp claim is missing', () => {
    const token = makeToken({ sub: 'US01ARZ3NDEKTSV4RRFFQ69G5FAV' })
    expect(isTokenExpired(token)).toBe(true)
  })

  it('returns true when the exp claim is not a number', () => {
    const token = makeToken({ exp: 'tomorrow' })
    expect(isTokenExpired(token)).toBe(true)
  })

  it('returns true for null, undefined, and empty tokens', () => {
    expect(isTokenExpired(null)).toBe(true)
    expect(isTokenExpired(undefined)).toBe(true)
    expect(isTokenExpired('')).toBe(true)
  })

  it('returns true for malformed tokens', () => {
    expect(isTokenExpired('not-a-jwt')).toBe(true)
    expect(isTokenExpired('a.b.c')).toBe(true)
    expect(isTokenExpired('onlyonepart.')).toBe(true)
  })

  it('handles base64url payloads (with - and _ characters)', () => {
    // Payload chosen so the JSON encodes to base64 containing + and /
    const token = makeToken({
      exp: nowSeconds() + 24 * 60 * 60,
      name: '???>>>???>>>',
    })
    expect(isTokenExpired(token)).toBe(false)
  })
})
