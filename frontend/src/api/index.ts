import { request } from './client'
import type {
  CreateOrderInput,
  Order,
  Paginated,
  Product,
  ProductCreateInput,
  ProductStatusFilter,
  ProductUpdateInput,
  StockAdjustInput,
  StockMovement,
  Summary,
} from './types'

function qs(params: Record<string, string | number | undefined>): string {
  const sp = new URLSearchParams()
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== '') sp.set(k, String(v))
  }
  const s = sp.toString()
  return s ? `?${s}` : ''
}

export const api = {
  listProducts(params: { search?: string; status?: ProductStatusFilter; page?: number; limit?: number }) {
    return request<Paginated<Product>>(`/api/products${qs(params)}`)
  },

  getProduct(id: number) {
    return request<Product>(`/api/products/${id}`)
  },

  createProduct(input: ProductCreateInput) {
    return request<Product>('/api/products', { method: 'POST', body: JSON.stringify(input) })
  },

  updateProduct(id: number, input: ProductUpdateInput) {
    return request<Product>(`/api/products/${id}`, { method: 'PUT', body: JSON.stringify(input) })
  },

  deleteProduct(id: number) {
    return request<void>(`/api/products/${id}`, { method: 'DELETE' })
  },

  adjustStock(id: number, input: StockAdjustInput) {
    return request<Product>(`/api/products/${id}/stock`, { method: 'POST', body: JSON.stringify(input) })
  },

  listMovements(id: number) {
    return request<{ data: StockMovement[] }>(`/api/products/${id}/movements`)
  },

  getSummary() {
    return request<Summary>('/api/summary')
  },

  listOrders(params: { page?: number; limit?: number }) {
    return request<Paginated<Order>>(`/api/orders${qs(params)}`)
  },

  createOrder(input: CreateOrderInput, idempotencyKey?: string) {
    return request<Order>('/api/orders', {
      method: 'POST',
      body: JSON.stringify(input),
      headers: idempotencyKey ? { 'Idempotency-Key': idempotencyKey } : undefined,
    })
  },

  getOrder(id: number) {
    return request<Order>(`/api/orders/${id}`)
  },

  cancelOrder(id: number) {
    return request<Order>(`/api/orders/${id}/cancel`, { method: 'POST' })
  },
}
