import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link, useParams } from 'react-router-dom'
import { api } from '../api'
import { errorMessage } from '../utils/errorMessage'
import { formatIDR } from '../utils/format'
import EmptyState from '../components/EmptyState'
import ErrorState from '../components/ErrorState'
import OrderStatusBadge from '../components/OrderStatusBadge'
import Spinner from '../components/Spinner'

export default function OrderDetailPage() {
  const { id } = useParams()
  const orderId = Number(id)
  const valid = Number.isInteger(orderId) && orderId > 0

  const queryClient = useQueryClient()
  const [confirming, setConfirming] = useState(false)
  const [actionError, setActionError] = useState<string | null>(null)

  const orderQuery = useQuery({
    queryKey: ['order', orderId],
    queryFn: () => api.getOrder(orderId),
    enabled: valid,
  })

  const cancelMutation = useMutation({
    mutationFn: () => api.cancelOrder(orderId),
    onSuccess: () => {
      setConfirming(false)
      setActionError(null)
      // Stock is restored server-side; refresh products + summary + orders.
      queryClient.invalidateQueries({ queryKey: ['order', orderId] })
      queryClient.invalidateQueries({ queryKey: ['orders'] })
      queryClient.invalidateQueries({ queryKey: ['products'] })
      queryClient.invalidateQueries({ queryKey: ['summary'] })
    },
    onError: (err) => setActionError(errorMessage(err)),
  })

  if (!valid) {
    return <EmptyState message="Invalid order id." />
  }

  if (orderQuery.isLoading) {
    return (
      <div className="flex items-center justify-center gap-2 py-12 text-gray-500">
        <Spinner /> Loading order…
      </div>
    )
  }

  if (orderQuery.isError) {
    return <ErrorState message={errorMessage(orderQuery.error)} onRetry={() => orderQuery.refetch()} />
  }

  const order = orderQuery.data
  if (!order) return <EmptyState message="Order not found." />

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <Link to="/orders" className="text-sm text-blue-600 hover:underline">
            ← Orders
          </Link>
          <h1 className="text-2xl font-semibold text-gray-900">Order #{order.id}</h1>
          <p className="text-sm text-gray-500">
            {order.customer} · {new Date(order.createdAt).toLocaleString('id-ID')}
          </p>
        </div>
        <OrderStatusBadge status={order.status} />
      </div>

      {actionError && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
          {actionError}
        </div>
      )}

      <div className="overflow-x-auto rounded-lg border border-gray-200 bg-white">
        <table className="min-w-full divide-y divide-gray-200 text-sm">
          <thead className="bg-gray-50 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
            <tr>
              <th className="px-4 py-3">SKU</th>
              <th className="px-4 py-3 text-right">Unit price</th>
              <th className="px-4 py-3 text-right">Qty</th>
              <th className="px-4 py-3 text-right">Subtotal</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {order.items.map((it) => (
              <tr key={it.id}>
                <td className="px-4 py-3 font-mono text-xs text-gray-700">{it.sku}</td>
                <td className="px-4 py-3 text-right text-gray-700">{formatIDR(it.unitPrice)}</td>
                <td className="px-4 py-3 text-right text-gray-700">{it.qty}</td>
                <td className="px-4 py-3 text-right font-medium text-gray-900">
                  {formatIDR(it.unitPrice * it.qty)}
                </td>
              </tr>
            ))}
          </tbody>
          <tfoot className="bg-gray-50">
            <tr>
              <td colSpan={3} className="px-4 py-3 text-right font-medium text-gray-900">
                Total
              </td>
              <td className="px-4 py-3 text-right text-lg font-bold text-gray-900">
                {formatIDR(order.total)}
              </td>
            </tr>
          </tfoot>
        </table>
      </div>

      {order.status === 'CREATED' && (
        <div className="flex justify-end">
          {!confirming ? (
            <button
              onClick={() => setConfirming(true)}
              className="rounded-md border border-red-300 px-4 py-2 text-sm font-medium text-red-600 hover:bg-red-50"
            >
              Cancel order
            </button>
          ) : (
            <div className="flex items-center gap-2 rounded-md border border-red-200 bg-red-50 p-3">
              <span className="text-sm text-red-700">Cancel this order and restore stock?</span>
              <button
                onClick={() => cancelMutation.mutate()}
                disabled={cancelMutation.isPending}
                className="inline-flex items-center gap-2 rounded-md bg-red-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-60"
              >
                {cancelMutation.isPending && <Spinner />}
                {cancelMutation.isPending ? 'Cancelling…' : 'Yes, cancel'}
              </button>
              <button
                onClick={() => setConfirming(false)}
                className="rounded-md border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-50"
              >
                Keep order
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
