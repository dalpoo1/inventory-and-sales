import type { Product } from '../api/types'
import { formatIDR, formatTimeAgo } from '../utils/format'
import StatusBadge from './StatusBadge'

export default function ProductTable({
  products,
  deletingId,
  onAdjust,
  onEdit,
  onDelete,
}: {
  products: Product[]
  deletingId: number | null
  onAdjust: (p: Product) => void
  onEdit: (p: Product) => void
  onDelete: (p: Product) => void
}) {
  return (
    <div className="overflow-x-auto rounded-lg border border-gray-200 bg-white">
      <table className="min-w-full divide-y divide-gray-200 text-sm">
        <thead className="bg-gray-50 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
          <tr>
            <th className="px-4 py-3">SKU</th>
            <th className="px-4 py-3">Name</th>
            <th className="px-4 py-3 text-right">Price</th>
            <th className="px-4 py-3 text-right">Stock</th>
            <th className="px-4 py-3">Status</th>
            <th className="px-4 py-3">Last movement</th>
            <th className="px-4 py-3 text-right">Actions</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-100">
          {products.map((p) => (
            <tr key={p.id} className="hover:bg-gray-50">
              <td className="px-4 py-3 font-mono text-xs text-gray-700">{p.sku}</td>
              <td className="px-4 py-3 text-gray-900">{p.name}</td>
              <td className="px-4 py-3 text-right text-gray-900">{formatIDR(p.price)}</td>
              <td className="px-4 py-3 text-right">
                <span className={p.stock <= 10 ? 'font-semibold text-red-600' : 'text-gray-900'}>
                  {p.stock}
                </span>
              </td>
              <td className="px-4 py-3">
                <StatusBadge isActive={p.isActive} />
              </td>
              <td className="px-4 py-3 text-gray-500">{formatTimeAgo(p.lastMovementAt)}</td>
              <td className="px-4 py-3">
                <div className="flex justify-end gap-2">
                  <button
                    onClick={() => onEdit(p)}
                    className="rounded-md border border-gray-300 px-3 py-1.5 text-xs font-medium text-gray-700 hover:bg-gray-50"
                  >
                    Edit
                  </button>
                  <button
                    onClick={() => onAdjust(p)}
                    className="rounded-md border border-gray-300 px-3 py-1.5 text-xs font-medium text-gray-700 hover:bg-gray-50"
                  >
                    Adjust
                  </button>
                  <button
                    onClick={() => onDelete(p)}
                    disabled={deletingId === p.id}
                    className="rounded-md border border-red-200 px-3 py-1.5 text-xs font-medium text-red-600 hover:bg-red-50 disabled:opacity-50"
                  >
                    {deletingId === p.id ? 'Deleting…' : 'Delete'}
                  </button>
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
