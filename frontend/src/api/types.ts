// Shared domain types for the whole frontend. No `any` on domain data.

export interface Product {
  id: number
  sku: string
  name: string
  price: number
  stock: number
  isActive: boolean
  createdAt: string
  lastMovementAt: string | null
}

export type MovementType = 'in' | 'out'

export interface StockMovement {
  id: number
  productId: number
  type: MovementType
  qty: number
  note: string | null
  createdAt: string
}

export type OrderStatus = 'CREATED' | 'CANCELLED'

export interface OrderItem {
  id: number
  sku: string
  qty: number
  unitPrice: number
}

export interface Order {
  id: number
  customer: string
  status: OrderStatus
  total: number
  createdAt: string
  cancelledAt: string | null
  items: OrderItem[]
}

export interface Summary {
  totalProducts: number
  activeProducts: number
  totalStockValue: number
  lowStockCount: number
}

export interface Paginated<T> {
  data: T[]
  page: number
  limit: number
  total: number
}

export type ProductStatusFilter = 'all' | 'active' | 'inactive' | 'low'

export interface StockAdjustInput {
  type: MovementType
  qty: number
  note?: string
}

export interface CreateOrderItemInput {
  sku: string
  qty: number
}

export interface CreateOrderInput {
  customer: string
  items: CreateOrderItemInput[]
}

export interface ProductCreateInput {
  sku: string
  name: string
  price: number
}

export interface ProductUpdateInput {
  sku?: string
  name?: string
  price?: number
  isActive?: boolean
}
