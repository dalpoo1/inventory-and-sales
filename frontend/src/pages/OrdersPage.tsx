import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api } from '../api'
import { queryKeys } from '../api/queryKeys'
import { errorMessage } from '../utils/errorMessage'
import { formatIDR } from '../utils/format'
import CreateOrderModal from '../components/CreateOrderModal'
import EmptyState from '../components/EmptyState'
import ErrorState from '../components/ErrorState'
import OrderStatusBadge from '../components/OrderStatusBadge'
import Spinner from '../components/Spinner'

export default function OrdersPage() {
  const [page, setPage] = useState(1)
  const [showCreate, setShowCreate] = useState(false)
  const queryClient = useQueryClient()

  const ordersQuery = useQuery({
    queryKey: queryKeys.orders.list(page),
    queryFn: () => api.listOrders({ page }),
  })

  const onCreated = () => {
    queryClient.invalidateQueries({ queryKey: queryKeys.orders.all })
    queryClient.invalidateQueries({ queryKey: queryKeys.products.all })
    queryClient.invalidateQueries({ queryKey: queryKeys.summary })
  }

  const data = ordersQuery.data
  const totalPages = data ? Math.max(1, Math.ceil(data.total / data.limit)) : 1

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold text-gray-900">Orders</h1>
        <button
          onClick={() => setShowCreate(true)}
          className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          New order
        </button>
      </div>

      {ordersQuery.isLoading && (
        <div className="flex items-center justify-center gap-2 py-12 text-gray-500">
          <Spinner /> Loading orders…
        </div>
      )}

      {ordersQuery.isError && (
        <ErrorState message={errorMessage(ordersQuery.error)} onRetry={() => ordersQuery.refetch()} />
      )}

      {data && data.data.length === 0 && (
        <EmptyState message="No orders yet. Create your first order." />
      )}

      {data && data.data.length > 0 && (
        <>
          <div className="overflow-x-auto rounded-lg border border-gray-200 bg-white">
            <table className="min-w-full divide-y divide-gray-200 text-sm">
              <thead className="bg-gray-50 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
                <tr>
                  <th className="px-4 py-3">Order</th>
                  <th className="px-4 py-3">Customer</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3 text-right">Items</th>
                  <th className="px-4 py-3 text-right">Total</th>
                  <th className="px-4 py-3">Created</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {data.data.map((o) => (
                  <tr key={o.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3">
                      <Link to={`/orders/${o.id}`} className="font-medium text-blue-600 hover:underline">
                        #{o.id}
                      </Link>
                    </td>
                    <td className="px-4 py-3 text-gray-900">{o.customer}</td>
                    <td className="px-4 py-3">
                      <OrderStatusBadge status={o.status} />
                    </td>
                    <td className="px-4 py-3 text-right text-gray-700">
                      {o.items.reduce((sum, it) => sum + it.qty, 0)}
                    </td>
                    <td className="px-4 py-3 text-right font-medium text-gray-900">
                      {formatIDR(o.total)}
                    </td>
                    <td className="px-4 py-3 text-gray-500">
                      {new Date(o.createdAt).toLocaleDateString('id-ID')}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="flex items-center justify-between text-sm text-gray-600">
            <span>
              Page {data.page} of {totalPages} · {data.total} order{data.total === 1 ? '' : 's'}
            </span>
            <div className="flex gap-2">
              <button
                disabled={page <= 1}
                onClick={() => setPage((p) => p - 1)}
                className="rounded-md border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-40"
              >
                Previous
              </button>
              <button
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
                className="rounded-md border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-40"
              >
                Next
              </button>
            </div>
          </div>
        </>
      )}

      {showCreate && (
        <CreateOrderModal onClose={() => setShowCreate(false)} onCreated={onCreated} />
      )}
    </div>
  )
}
