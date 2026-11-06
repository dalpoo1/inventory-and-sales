import { describe, expect, it } from 'vitest'
import { validateOrderDraft, validateProduct, validateStockAdjust } from './validation'

describe('validateOrderDraft', () => {
  it('requires a customer and at least one item', () => {
    const errors = validateOrderDraft('', [])
    expect(errors.customer).toBeDefined()
    expect(errors.items).toBeDefined()
  })

  it('flags duplicate SKUs on the later item', () => {
    const errors = validateOrderDraft('Acme', [
      { sku: 'SKU-1', qty: 1 },
      { sku: 'SKU-1', qty: 2 },
    ])
    expect(errors.itemErrors?.[1]?.sku).toContain('Duplicate')
  })

  it('flags non-positive quantity', () => {
    const errors = validateOrderDraft('Acme', [{ sku: 'SKU-1', qty: 0 }])
    expect(errors.itemErrors?.[0]?.qty).toBeDefined()
  })

  it('accepts a valid order', () => {
    const errors = validateOrderDraft('Acme', [{ sku: 'SKU-1', qty: 2 }])
    expect(Object.keys(errors)).toHaveLength(0)
  })
})

describe('validateStockAdjust', () => {
  it('rejects an invalid direction and non-positive quantity', () => {
    const errors = validateStockAdjust('sideways', 0)
    expect(errors.type).toBeDefined()
    expect(errors.qty).toBeDefined()
  })

  it('accepts a valid adjustment', () => {
    expect(Object.keys(validateStockAdjust('out', 5))).toHaveLength(0)
  })
})

describe('validateProduct', () => {
  it('requires sku/name and a non-negative price', () => {
    const errors = validateProduct('', '', -1)
    expect(errors.sku).toBeDefined()
    expect(errors.name).toBeDefined()
    expect(errors.price).toBeDefined()
  })

  it('accepts a valid product', () => {
    expect(Object.keys(validateProduct('SKU-1', 'Widget', 100))).toHaveLength(0)
  })
})
