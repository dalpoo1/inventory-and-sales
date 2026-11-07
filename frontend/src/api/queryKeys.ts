// Centralised TanStack Query keys. Keeping keys in one place prevents typos and
// makes invalidations (partial-key matching) safe to change in a single spot.
import type { ProductStatusFilter } from './types'

export const queryKeys = {
  products: {
    all: ['products'],
    list: (params: { search: string; status: ProductStatusFilter; page: number }) => [
      'products',
      params,
    ],
    options: ['products', 'options'],
  },
  summary: ['summary'],
  orders: {
    all: ['orders'],
    list: (page: number) => ['orders', page],
    detail: (id: number) => ['order', id],
  },
}
