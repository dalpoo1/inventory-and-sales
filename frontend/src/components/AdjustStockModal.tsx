import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { api } from '../api'
import { ApiError } from '../api/client'
import type { MovementType, Product } from '../api/types'
import { validateStockAdjust } from '../utils/validation'
import Modal from './Modal'
import Spinner from './Spinner'

export default function AdjustStockModal({
  product,
  onClose,
  onSuccess,
}: {
  product: Product
  onClose: () => void
  onSuccess: () => void
}) {
  const [type, setType] = useState<MovementType>('in')
  const [qty, setQty] = useState('')
  const [note, setNote] = useState('')
  const [fieldErrors, setFieldErrors] = useState<ReturnType<typeof validateStockAdjust>>({})
  const [serverError, setServerError] = useState<string | null>(null)

  const mutation = useMutation({
    mutationFn: () => api.adjustStock(product.id, { type, qty: Number(qty), note: note || undefined }),
    onSuccess: () => onSuccess(),
    onError: (err) => {
      if (err instanceof ApiError && err.code === 'INSUFFICIENT_STOCK') {
        // Surface the 409 in the UI with the available quantity — never console-only.
        setServerError(`Insufficient stock — only ${err.available} available.`)
      } else if (err instanceof ApiError) {
        setServerError(err.message || 'Failed to adjust stock.')
      } else {
        setServerError('Network error. Please try again.')
      }
    },
  })

  const submit = () => {
    const errors = validateStockAdjust(type, Number(qty))
    setFieldErrors(errors)
    if (Object.keys(errors).length > 0) return
    setServerError(null)
    mutation.mutate()
  }

  return (
    <Modal title={`Adjust stock — ${product.sku}`} onClose={onClose}>
      <div className="space-y-4">
        <div>
          <label className="mb-1 block text-sm font-medium text-gray-700">Direction</label>
          <select
            value={type}
            onChange={(e) => setType(e.target.value as MovementType)}
            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
          >
            <option value="in">Stock in (+)</option>
            <option value="out">Stock out (−)</option>
          </select>
          {fieldErrors.type && <p className="mt-1 text-xs text-red-600">{fieldErrors.type}</p>}
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium text-gray-700">Quantity</label>
          <input
            type="number"
            value={qty}
            onChange={(e) => setQty(e.target.value)}
            placeholder="0"
            min={1}
            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
          />
          {fieldErrors.qty && <p className="mt-1 text-xs text-red-600">{fieldErrors.qty}</p>}
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium text-gray-700">Note (optional)</label>
          <input
            type="text"
            value={note}
            onChange={(e) => setNote(e.target.value)}
            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
          />
        </div>

        {serverError && (
          <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
            {serverError}
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
            {mutation.isPending ? 'Saving…' : 'Save'}
          </button>
        </div>
      </div>
    </Modal>
  )
}
