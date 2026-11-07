import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { api } from '../api'
import { ApiError } from '../api/client'
import type { Product } from '../api/types'
import { validateProduct } from '../utils/validation'
import Modal from './Modal'
import Spinner from './Spinner'

// One modal, two modes: `product` is passed for edit, omitted for create.
export default function ProductFormModal({
  product,
  onClose,
  onSuccess,
}: {
  product?: Product | null
  onClose: () => void
  onSuccess: () => void
}) {
  const isEdit = product != null

  const [sku, setSku] = useState(product?.sku ?? '')
  const [name, setName] = useState(product?.name ?? '')
  const [price, setPrice] = useState(product ? String(product.price) : '')
  const [isActive, setIsActive] = useState(product?.isActive ?? true)
  const [fieldErrors, setFieldErrors] = useState<ReturnType<typeof validateProduct>>({})
  const [serverError, setServerError] = useState<string | null>(null)

  const mutation = useMutation({
    mutationFn: () =>
      isEdit
        ? api.updateProduct(product.id, {
            sku: sku.trim(),
            name: name.trim(),
            price: Number(price),
            isActive,
          })
        : api.createProduct({ sku: sku.trim(), name: name.trim(), price: Number(price) }),
    onSuccess: () => onSuccess(),
    onError: (err) => {
      if (err instanceof ApiError && err.fields?.sku) {
        setFieldErrors({ sku: err.fields.sku })
      } else if (err instanceof ApiError) {
        setServerError(err.message || 'Failed to save product.')
      } else {
        setServerError('Network error. Please try again.')
      }
    },
  })

  const submit = () => {
    const errors = validateProduct(sku, name, Number(price))
    setFieldErrors(errors)
    if (Object.keys(errors).length > 0) return
    setServerError(null)
    mutation.mutate()
  }

  return (
    <Modal title={isEdit ? 'Edit product' : 'Add product'} onClose={onClose}>
      <div className="space-y-4">
        <div>
          <label className="mb-1 block text-sm font-medium text-gray-700">SKU</label>
          <input
            type="text"
            value={sku}
            onChange={(e) => setSku(e.target.value)}
            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
          />
          {fieldErrors.sku && <p className="mt-1 text-xs text-red-600">{fieldErrors.sku}</p>}
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium text-gray-700">Name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
          />
          {fieldErrors.name && <p className="mt-1 text-xs text-red-600">{fieldErrors.name}</p>}
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium text-gray-700">Price (IDR)</label>
          <input
            type="number"
            value={price}
            min={0}
            onChange={(e) => setPrice(e.target.value)}
            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
          />
          {fieldErrors.price && <p className="mt-1 text-xs text-red-600">{fieldErrors.price}</p>}
        </div>

        {isEdit && (
          <label className="flex items-center gap-2 text-sm font-medium text-gray-700">
            <input
              type="checkbox"
              checked={isActive}
              onChange={(e) => setIsActive(e.target.checked)}
              className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
            />
            Active
          </label>
        )}

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
            {mutation.isPending ? 'Saving…' : isEdit ? 'Save changes' : 'Save'}
          </button>
        </div>
      </div>
    </Modal>
  )
}
