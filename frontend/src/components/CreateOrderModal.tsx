import { useRef, useState } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api } from '../api'
import { ApiError } from '../api/client'
import { queryKeys } from '../api/queryKeys'
import type { CreateOrderInput, Order, Product } from '../api/types'
import { formatIDR } from '../utils/format'
import { errorMessage } from '../utils/errorMessage'
import { validateOrderDraft, type OrderItemDraft } from '../utils/validation'
import Modal from './Modal'
import Spinner from './Spinner'

function newIdempotencyKey(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  return `${Date.now()}-${Math.random().toString(36).slice(2)}`
}

export default function CreateOrderModal({
  onClose,
  onCreated,
}: {
  onClose: () => void
  onCreated?: (order: Order) => void
}) {
  const [customer, setCustomer] = useState('')
  const [items, setItems] = useState<OrderItemDraft[]>([{ sku: '', qty: 1 }])
  const [fieldErrors, setFieldErrors] = useState<ReturnType<typeof validateOrderDraft>>({})
  const [serverItemErrors, setServerItemErrors] = useState<Record<string, string>>({})
  const [serverMessage, setServerMessage] = useState<string | null>(null)
  const [created, setCreated] = useState<Order | null>(null)
  const idemKey = useRef<string | null>(null)

  // Selectable products for the SKU dropdown (all products, active or not).
  const productsQuery = useQuery({
    queryKey: queryKeys.products.options,
    queryFn: () => api.listProducts({ limit: 100 }),
  })
  const products = productsQuery.data?.data ?? []
  const productBySku = new Map<string, Product>(products.map((p) => [p.sku, p]))

  const mutation = useMutation({
    mutationFn: () => {
      if (!idemKey.current) idemKey.current = newIdempotencyKey()
      const input: CreateOrderInput = {
        customer: customer.trim(),
        items: items.map((it) => ({ sku: it.sku.trim(), qty: it.qty })),
      }
      return api.createOrder(input, idemKey.current)
    },
    onSuccess: (order) => {
      idemKey.current = null
      setCreated(order)
      onCreated?.(order)
    },
    onError: (err) => {
      if (err instanceof ApiError) {
        if (err.code === 'INSUFFICIENT_STOCK' && err.sku) {
          const sku = err.sku
          setServerItemErrors((prev) => ({
            ...prev,
            [sku]: `Insufficient stock — only ${err.available} available`,
          }))
        } else {
          setServerMessage(err.message || 'Failed to create order.')
        }
      } else {
        setServerMessage('Network error. Please try again.')
      }
    },
  })

  const resetIdem = () => {
    idemKey.current = null
  }

  const updateItem = (i: number, patch: Partial<OrderItemDraft>) => {
    resetIdem()
    setServerItemErrors({})
    setItems((prev) => prev.map((it, idx) => (idx === i ? { ...it, ...patch } : it)))
  }
  const addItem = () => {
    resetIdem()
    setItems((prev) => [...prev, { sku: '', qty: 1 }])
  }
  const removeItem = (i: number) => {
    resetIdem()
    setItems((prev) => prev.filter((_, idx) => idx !== i))
  }

  const submit = () => {
    const errors = validateOrderDraft(customer, items)
    setFieldErrors(errors)
    setServerMessage(null)
    setServerItemErrors({})
    if (errors.customer || errors.items || errors.itemErrors) return
    mutation.mutate()
  }

  if (created) {
    return (
      <Modal title="Order created" onClose={onClose}>
        <div className="space-y-4">
          <div className="rounded-md border border-green-200 bg-green-50 p-4">
            <p className="text-sm text-green-800">
              Order <span className="font-semibold">#{created.id}</span> for{' '}
              <span className="font-semibold">{created.customer}</span>
            </p>
            <p className="mt-2 text-2xl font-bold text-green-900">{formatIDR(created.total)}</p>
          </div>

          <ul className="divide-y divide-gray-100 text-sm">
            {created.items.map((it) => (
              <li key={it.id} className="flex justify-between py-2">
                <span>
                  {it.sku} × {it.qty}
                </span>
                <span className="text-gray-600">{formatIDR(it.unitPrice * it.qty)}</span>
              </li>
            ))}
          </ul>

          <div className="flex justify-end gap-2">
            <button
              onClick={onClose}
              className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
            >
              Close
            </button>
            <Link
              to={`/orders/${created.id}`}
              onClick={onClose}
              className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
            >
              View order
            </Link>
          </div>
        </div>
      </Modal>
    )
  }

  return (
    <Modal title="Create order" onClose={onClose}>
      <div className="space-y-4">
        <div>
          <label className="mb-1 block text-sm font-medium text-gray-700">Customer</label>
          <input
            type="text"
            value={customer}
            onChange={(e) => {
              resetIdem()
              setCustomer(e.target.value)
            }}
            placeholder="Customer name"
            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
          />
          {fieldErrors.customer && (
            <p className="mt-1 text-xs text-red-600">{fieldErrors.customer}</p>
          )}
        </div>

        <div>
          <div className="mb-1 flex items-center justify-between">
            <label className="block text-sm font-medium text-gray-700">Items</label>
            <button onClick={addItem} className="text-sm font-medium text-blue-600 hover:underline">
              + Add item
            </button>
          </div>

          {productsQuery.isLoading && (
            <p className="mb-2 flex items-center gap-2 text-xs text-gray-500">
              <Spinner /> Loading products…
            </p>
          )}
          {productsQuery.isError && (
            <p className="mb-2 text-xs text-red-600">
              Couldn't load products — {errorMessage(productsQuery.error)}
            </p>
          )}

          <div className="space-y-2">
            {items.map((item, i) => {
              const selected = item.sku ? productBySku.get(item.sku) : undefined
              return (
                <div key={i} className="rounded-md border border-gray-200 p-2">
                  <div className="flex gap-2">
                    <select
                      value={item.sku}
                      onChange={(e) => updateItem(i, { sku: e.target.value })}
                      className="w-1/2 rounded-md border border-gray-300 px-2 py-1.5 text-sm"
                    >
                      <option value="">Select product…</option>
                      {products.map((p) => (
                        <option key={p.sku} value={p.sku}>
                          {p.sku} — {p.name} (stock {p.stock})
                          {p.isActive ? '' : ' · inactive'}
                        </option>
                      ))}
                    </select>
                    <input
                      type="number"
                      value={item.qty}
                      min={1}
                      onChange={(e) => updateItem(i, { qty: Number(e.target.value) })}
                      placeholder="Qty"
                      className="w-1/3 rounded-md border border-gray-300 px-2 py-1.5 text-sm"
                    />
                    <button
                      onClick={() => removeItem(i)}
                      disabled={items.length === 1}
                      className="rounded-md border border-gray-300 px-2 text-sm text-gray-500 hover:bg-gray-50 disabled:opacity-40"
                      aria-label="Remove item"
                    >
                      ✕
                    </button>
                  </div>
                  {selected && (
                    <p className="mt-1 text-xs text-gray-500">
                      {formatIDR(selected.price)} each · {selected.stock} in stock
                      {selected.isActive ? '' : ' · inactive'}
                    </p>
                  )}
                  {(fieldErrors.itemErrors?.[i]?.sku || serverItemErrors[item.sku.trim()]) && (
                    <p className="mt-1 text-xs text-red-600">
                      {serverItemErrors[item.sku.trim()] ?? fieldErrors.itemErrors?.[i]?.sku}
                    </p>
                  )}
                  {fieldErrors.itemErrors?.[i]?.qty && (
                    <p className="mt-1 text-xs text-red-600">{fieldErrors.itemErrors[i].qty}</p>
                  )}
                </div>
              )
            })}
          </div>
          {fieldErrors.items && <p className="mt-1 text-xs text-red-600">{fieldErrors.items}</p>}
        </div>

        {serverMessage && (
          <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
            {serverMessage}
          </div>
        )}

        <div className="flex justify-end gap-2 pt-2">
          <button
            onClick={onClose}
            className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            onClick={submit}
            disabled={mutation.isPending}
            className="inline-flex items-center gap-2 rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-60"
          >
            {mutation.isPending && <Spinner />}
            {mutation.isPending ? 'Creating…' : 'Create order'}
          </button>
        </div>
      </div>
    </Modal>
  )
}
