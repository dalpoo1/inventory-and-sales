// Pure formatting helpers (unit-tested).

export function formatIDR(amount: number): string {
  return 'Rp ' + amount.toLocaleString('id-ID')
}

export function formatTimeAgo(iso: string | null, now: number = Date.now()): string {
  if (!iso) return '—'
  const then = new Date(iso).getTime()
  if (Number.isNaN(then)) return '—'

  const diff = Math.max(0, now - then)
  const sec = Math.floor(diff / 1000)
  if (sec < 60) return 'just now'

  const min = Math.floor(sec / 60)
  if (min < 60) return `${min}m ago`

  const hr = Math.floor(min / 60)
  if (hr < 24) return `${hr}h ago`

  const day = Math.floor(hr / 24)
  if (day < 30) return `${day}d ago`

  const mo = Math.floor(day / 30)
  if (mo < 12) return `${mo}mo ago`

  return `${Math.floor(mo / 12)}y ago`
}
