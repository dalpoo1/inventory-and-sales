import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../api'
import { ApiError } from '../api/client'
import type { Product, ProductStatusFilter } from '../api/types'
import { useDebounce } from '../hooks/useDebounce'
import { errorMessage } from '../utils/errorMessage'
import AdjustStockModal from '../components/AdjustStockModal'
import EmptyState from '../components/EmptyState'
import ErrorState from '../components/ErrorState'
import ProductFormModal from '../components/ProductFormModal'
import ProductTable from '../components/ProductTable'
import Spinner from '../components/Spinner'
import SummaryCards from '../components/SummaryCards'

const FILTERS: { value: ProductStatusFilter; label: string }[] = [
  { value: 'all', label: 'All' },
  { value: 'active', label: 'Active' },
  { value: 'inactive', label: 'Inactive' },
  { value: 'low', label: 'Low stock' },
]

export default function ProductsPage() {
  const [search, setSearch] = useState('')
  const debouncedSearch = useDebounce(search, 300)
  const [status, setStatus] = useState<ProductStatusFilter>('all')
  const [page, setPage] = useState(1)

  const [adjusting, setAdjusting] = useState<Product | null>(null)
  const [showCreate, setShowCreate] = useState(false)
  const [deletingId, setDeletingId] = useState<number | null>(null)
  const [pageError, setPageError] = useState<string | null>(null)

  const queryClient = useQueryClient()

  const productsQuery = useQuery({
    queryKey: ['products', { search: debouncedSearch, status, page }],
    queryFn: () => api.listProducts({ search: debouncedSearch, status, page }),
  })

  const summaryQuery = useQuery({ queryKey: ['summary'], queryFn: () => api.getSummary() })

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ['products'] })
    queryClient.invalidateQueries({ queryKey: ['summary'] })
  }

  const deleteMutation = useMutation({
    mutationFn: (id: number) => api.deleteProduct(id),
    onSuccess: () => {
      setDeletingId(null)
      setPageError(null)
      invalidate()
    },
    onError: (err) => {
      setDeletingId(null)
      if (err instanceof ApiError && err.code === 'PRODUCT_HAS_HISTORY') {
        setPageError('Cannot delete this product: it has movement history.')
      } else {
        setPageError(errorMessage(err))
      }
    },
  })

  const onDelete = (p: Product) => {
    setPageError(null)
    if (!window.confirm(`Delete ${p.sku} — ${p.name}?`)) return
    setDeletingId(p.id)
    deleteMutation.mutate(p.id)
  }

  const data = productsQuery.data
  const totalPages = data ? Math.max(1, Math.ceil(data.total / data.limit)) : 1

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold text-gray-900">Products</h1>
        <button
          onClick={() => setShowCreate(true)}
          className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          Add product
        </button>
      </div>

      <SummaryCards summary={summaryQuery.data} loading={summaryQuery.isLoading} />

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
        <input
          value={search}
          onChange={(e) => {
            setSearch(e.target.value)
            setPage(1)
          }}
          placeholder="Search SKU or name…"
          className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm sm:max-w-xs"
        />
        <div className="flex flex-wrap gap-1">
          {FILTERS.map((f) => (
            <button
              key={f.value}
              onClick={() => {
                setStatus(f.value)
                setPage(1)
              }}
              className={`rounded-md px-3 py-2 text-sm font-medium ${
                status === f.value
                  ? 'bg-blue-600 text-white'
                  : 'border border-gray-300 bg-white text-gray-700 hover:bg-gray-50'
              }`}
            >
              {f.label}
            </button>
          ))}
        </div>
      </div>

      {pageError && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
          {pageError}
        </div>
      )}

      {productsQuery.isLoading && (
        <div className="flex items-center justify-center gap-2 py-12 text-gray-500">
          <Spinner /> Loading products…
        </div>
      )}

      {productsQuery.isError && (
        <ErrorState message={errorMessage(productsQuery.error)} onRetry={() => productsQuery.refetch()} />
      )}

      {data && data.data.length === 0 && <EmptyState message="No products found." />}

      {data && data.data.length > 0 && (
        <>
          <ProductTable
            products={data.data}
            deletingId={deletingId}
            onAdjust={setAdjusting}
            onDelete={onDelete}
          />
          <div className="flex items-center justify-between text-sm text-gray-600">
            <span>
              Page {data.page} of {totalPages} · {data.total} product{data.total === 1 ? '' : 's'}
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

      {adjusting && (
        <AdjustStockModal
          product={adjusting}
          onClose={() => setAdjusting(null)}
          onSuccess={() => {
            setAdjusting(null)
            invalidate()
          }}
        />
      )}
      {showCreate && (
        <ProductFormModal
          onClose={() => setShowCreate(false)}
          onSuccess={() => {
            setShowCreate(false)
            invalidate()
          }}
        />
      )}
    </div>
  )
}
