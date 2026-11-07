// Product, stock-adjustment, movement, and summary endpoints.
import { http } from './client'
import type {
  Paginated,
  Product,
  ProductCreateInput,
  ProductStatusFilter,
  ProductUpdateInput,
  StockAdjustInput,
  StockMovement,
  Summary,
} from './types'

export interface ProductListParams {
  search?: string
  status?: ProductStatusFilter
  page?: number
  limit?: number
}

export const productsApi = {
  listProducts(params: ProductListParams = {}) {
    return http.get<Paginated<Product>>('/products', { params })
  },

  getProduct(id: number) {
    return http.get<Product>(`/products/${id}`)
  },

  createProduct(input: ProductCreateInput) {
    return http.post<Product>('/products', input)
  },

  updateProduct(id: number, input: ProductUpdateInput) {
    return http.put<Product>(`/products/${id}`, input)
  },

  deleteProduct(id: number) {
    return http.delete<void>(`/products/${id}`)
  },

  adjustStock(id: number, input: StockAdjustInput) {
    return http.post<Product>(`/products/${id}/stock`, input)
  },

  listMovements(id: number) {
    return http.get<{ data: StockMovement[] }>(`/products/${id}/movements`)
  },

  getSummary() {
    return http.get<Summary>('/summary')
  },
}
