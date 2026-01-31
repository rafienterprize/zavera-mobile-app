/**
 * Property Tests for Cart Total Calculation
 * Property 3: Cart Total Calculation
 */
import { describe, it, expect } from 'vitest';
import * as fc from 'fast-check';

// Types
interface CartItem {
  id: number;
  name: string;
  price: number;
  quantity: number;
}

// Cart total calculation logic
function calculateCartTotal(items: CartItem[], shippingCost: number = 0): number {
  const subtotal = items.reduce((sum, item) => sum + item.price * item.quantity, 0);
  return subtotal + shippingCost;
}

function calculateSubtotal(items: CartItem[]): number {
  return items.reduce((sum, item) => sum + item.price * item.quantity, 0);
}

// Arbitraries
const cartItemArb = fc.record({
  id: fc.integer({ min: 1, max: 10000 }),
  name: fc.string({ minLength: 1, maxLength: 50 }),
  price: fc.integer({ min: 1000, max: 5000000 }),
  quantity: fc.integer({ min: 1, max: 10 }),
});

const cartItemsArb = fc.array(cartItemArb, { minLength: 0, maxLength: 20 });
const shippingCostArb = fc.integer({ min: 0, max: 100000 });

describe('Property 3: Cart Total Calculation', () => {
  it('total should equal sum of (price × quantity) for all items plus shipping', () => {
    fc.assert(
      fc.property(cartItemsArb, shippingCostArb, (items, shipping) => {
        const total = calculateCartTotal(items, shipping);
        const expectedSubtotal = items.reduce((sum, item) => sum + item.price * item.quantity, 0);
        return total === expectedSubtotal + shipping;
      }),
      { numRuns: 100 }
    );
  });

  it('subtotal should be non-negative', () => {
    fc.assert(
      fc.property(cartItemsArb, (items) => {
        const subtotal = calculateSubtotal(items);
        return subtotal >= 0;
      }),
      { numRuns: 100 }
    );
  });

  it('empty cart should have zero subtotal', () => {
    const subtotal = calculateSubtotal([]);
    expect(subtotal).toBe(0);
  });

  it('total should be at least equal to shipping cost', () => {
    fc.assert(
      fc.property(cartItemsArb, shippingCostArb, (items, shipping) => {
        const total = calculateCartTotal(items, shipping);
        return total >= shipping;
      }),
      { numRuns: 100 }
    );
  });

  it('adding an item should increase total by price × quantity', () => {
    fc.assert(
      fc.property(cartItemsArb, cartItemArb, (existingItems, newItem) => {
        const totalBefore = calculateCartTotal(existingItems, 0);
        const totalAfter = calculateCartTotal([...existingItems, newItem], 0);
        return totalAfter === totalBefore + newItem.price * newItem.quantity;
      }),
      { numRuns: 100 }
    );
  });

  it('removing an item should decrease total by price × quantity', () => {
    fc.assert(
      fc.property(
        fc.array(cartItemArb, { minLength: 1, maxLength: 20 }),
        (items) => {
          const totalBefore = calculateCartTotal(items, 0);
          const removedItem = items[0];
          const remainingItems = items.slice(1);
          const totalAfter = calculateCartTotal(remainingItems, 0);
          return totalAfter === totalBefore - removedItem.price * removedItem.quantity;
        }
      ),
      { numRuns: 100 }
    );
  });
});
