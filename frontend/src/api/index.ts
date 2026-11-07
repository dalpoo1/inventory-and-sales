// Public API surface. Components import `api` from '../api' and `ApiError`
// from '../api/client' (or here); domain types come from '../api/types'.
import { ordersApi } from './orders'
import { productsApi } from './products'

export const api = {
  ...productsApi,
  ...ordersApi,
}

export * from './client'
export * from './types'
