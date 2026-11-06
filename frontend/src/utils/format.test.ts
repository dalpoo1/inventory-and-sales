import { describe, expect, it } from 'vitest'
import { formatIDR, formatTimeAgo } from './format'

describe('formatIDR', () => {
  it('uses dot thousands separator with an "Rp " prefix', () => {
    expect(formatIDR(1234567)).toBe('Rp 1.234.567')
    expect(formatIDR(75000)).toBe('Rp 75.000')
    expect(formatIDR(0)).toBe('Rp 0')
  })
})

describe('formatTimeAgo', () => {
  const now = new Date('2026-11-06T12:00:00Z').getTime()

  it('returns an em dash for null/absent timestamps', () => {
    expect(formatTimeAgo(null, now)).toBe('—')
  })

  it('returns "just now" under a minute', () => {
    expect(formatTimeAgo(new Date(now - 30_000).toISOString(), now)).toBe('just now')
  })

  it('returns minutes', () => {
    expect(formatTimeAgo(new Date(now - 5 * 60_000).toISOString(), now)).toBe('5m ago')
  })

  it('returns hours', () => {
    expect(formatTimeAgo(new Date(now - 2 * 3_600_000).toISOString(), now)).toBe('2h ago')
  })

  it('returns days', () => {
    expect(formatTimeAgo(new Date(now - 3 * 86_400_000).toISOString(), now)).toBe('3d ago')
  })
})
