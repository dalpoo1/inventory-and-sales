// Pure validation logic (unit-tested), mirrors the backend rules so the UI
// catches obvious mistakes before a round-trip.

export interface OrderItemDraft {
  sku: string
  qty: number
}

export interface OrderValidation {
  customer?: string
  items?: string
  itemErrors?: Record<number, { sku?: string; qty?: string }>
}

export function validateOrderDraft(customer: string, items: OrderItemDraft[]): OrderValidation {
  const errors: OrderValidation = {}
  if (!customer.trim()) errors.customer = 'Customer name is required'
  if (items.length === 0) errors.items = 'Add at least one item'

  const seen = new Set<string>()
  const itemErrors: Record<number, { sku?: string; qty?: string }> = {}

  items.forEach((item, i) => {
    const e: { sku?: string; qty?: string } = {}
    const sku = item.sku.trim()
    if (!sku) {
      e.sku = 'SKU is required'
    } else if (seen.has(sku)) {
      e.sku = `Duplicate SKU: ${sku}`
    } else {
      seen.add(sku)
    }
    if (!Number.isInteger(item.qty) || item.qty <= 0) {
      e.qty = 'Quantity must be a positive integer'
    }
    if (e.sku || e.qty) itemErrors[i] = e
  })

  if (Object.keys(itemErrors).length > 0) errors.itemErrors = itemErrors
  return errors
}

export interface StockAdjustValidation {
  type?: string
  qty?: string
}

export function validateStockAdjust(type: string, qty: number): StockAdjustValidation {
  const errors: StockAdjustValidation = {}
  if (type !== 'in' && type !== 'out') errors.type = 'Direction is required'
  if (!Number.isInteger(qty) || qty <= 0) errors.qty = 'Quantity must be a positive integer'
  return errors
}

export interface ProductValidation {
  sku?: string
  name?: string
  price?: string
}

export function validateProduct(sku: string, name: string, price: number): ProductValidation {
  const errors: ProductValidation = {}
  if (!sku.trim()) errors.sku = 'SKU is required'
  if (!name.trim()) errors.name = 'Name is required'
  if (!Number.isInteger(price) || price < 0) errors.price = 'Price must be a non-negative integer'
  return errors
}
