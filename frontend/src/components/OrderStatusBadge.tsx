import type { OrderStatus } from '../api/types'

export default function OrderStatusBadge({ status }: { status: OrderStatus }) {
  return status === 'CREATED' ? (
    <span className="inline-flex items-center rounded-full bg-blue-100 px-2 py-0.5 text-xs font-medium text-blue-700">
      Created
    </span>
  ) : (
    <span className="inline-flex items-center rounded-full bg-red-100 px-2 py-0.5 text-xs font-medium text-red-700">
      Cancelled
    </span>
  )
}
