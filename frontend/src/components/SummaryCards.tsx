import type { Summary } from '../api/types'
import { formatIDR } from '../utils/format'

export default function SummaryCards({
  summary,
  loading,
}: {
  summary: Summary | undefined
  loading: boolean
}) {
  const cards = [
    { label: 'Total products', value: summary ? String(summary.totalProducts) : undefined },
    { label: 'Active products', value: summary ? String(summary.activeProducts) : undefined },
    {
      label: 'Total stock value',
      value: summary ? formatIDR(summary.totalStockValue) : undefined,
    },
    { label: 'Low stock (≤ 10)', value: summary ? String(summary.lowStockCount) : undefined },
  ]

  return (
    <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
      {cards.map((c) => (
        <div key={c.label} className="rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
          <p className="text-xs font-medium uppercase tracking-wide text-gray-500">{c.label}</p>
          <p className="mt-1 text-2xl font-semibold text-gray-900">
            {loading ? '—' : (c.value ?? '—')}
          </p>
        </div>
      ))}
    </div>
  )
}
