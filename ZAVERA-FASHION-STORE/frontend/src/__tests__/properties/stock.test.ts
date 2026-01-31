/**
 * Property Tests for Low Stock Indicator
 * Property 5: Low Stock Indicator Display
 */
import { describe, it, expect } from 'vitest';
import * as fc from 'fast-check';

// Constants
const LOW_STOCK_THRESHOLD = 10;

// Types
interface Product {
  id: number;
  name: string;
  stock: number;
}

// Low stock logic
function shouldShowLowStockIndicator(stock: number): boolean {
  return stock > 0 && stock < LOW_STOCK_THRESHOLD;
}

function getLowStockMessage(stock: number): string | null {
  if (shouldShowLowStockIndicator(stock)) {
    return `Sisa ${stock}`;
  }
  return null;
}

function isOutOfStock(stock: number): boolean {
  return stock <= 0;
}

// Arbitraries
const stockArb = fc.integer({ min: 0, max: 100 });
const lowStockArb = fc.integer({ min: 1, max: LOW_STOCK_THRESHOLD - 1 });
const normalStockArb = fc.integer({ min: LOW_STOCK_THRESHOLD, max: 100 });
const outOfStockArb = fc.constant(0);

describe('Property 5: Low Stock Indicator Display', () => {
  it('should show low stock indicator when stock < threshold and > 0', () => {
    fc.assert(
      fc.property(lowStockArb, (stock) => {
        return shouldShowLowStockIndicator(stock) === true;
      }),
      { numRuns: 100 }
    );
  });

  it('should NOT show low stock indicator when stock >= threshold', () => {
    fc.assert(
      fc.property(normalStockArb, (stock) => {
        return shouldShowLowStockIndicator(stock) === false;
      }),
      { numRuns: 100 }
    );
  });

  it('should NOT show low stock indicator when stock is 0', () => {
    expect(shouldShowLowStockIndicator(0)).toBe(false);
  });

  it('low stock message should contain the stock number', () => {
    fc.assert(
      fc.property(lowStockArb, (stock) => {
        const message = getLowStockMessage(stock);
        return message !== null && message.includes(stock.toString());
      }),
      { numRuns: 100 }
    );
  });

  it('low stock message should be null for normal stock', () => {
    fc.assert(
      fc.property(normalStockArb, (stock) => {
        return getLowStockMessage(stock) === null;
      }),
      { numRuns: 100 }
    );
  });

  it('out of stock should be true only when stock <= 0', () => {
    fc.assert(
      fc.property(stockArb, (stock) => {
        return isOutOfStock(stock) === (stock <= 0);
      }),
      { numRuns: 100 }
    );
  });

  it('low stock and out of stock should be mutually exclusive', () => {
    fc.assert(
      fc.property(stockArb, (stock) => {
        const isLow = shouldShowLowStockIndicator(stock);
        const isOut = isOutOfStock(stock);
        // Cannot be both low stock and out of stock
        return !(isLow && isOut);
      }),
      { numRuns: 100 }
    );
  });
});
