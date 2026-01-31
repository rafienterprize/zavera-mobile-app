/**
 * Property Tests for Product Filtering System
 * Properties 1 & 2: Filter Results Consistency and Sort Order Correctness
 */
import { describe, it, expect } from 'vitest';
import * as fc from 'fast-check';

// Types
interface Product {
  id: number;
  name: string;
  price: number;
  subcategory?: string;
  stock: number;
}

interface ProductFilters {
  sizes: string[];
  priceRange: { min: number; max: number } | null;
  subcategory: string | null;
}

type SortOption = 'newest' | 'price-low' | 'price-high' | 'name';

// Filter logic (extracted from CategoryPage)
function filterProducts(products: Product[], filters: ProductFilters): Product[] {
  let result = [...products];

  // Apply subcategory filter
  if (filters.subcategory) {
    result = result.filter(
      (p) => p.subcategory?.toLowerCase() === filters.subcategory?.toLowerCase()
    );
  }

  // Apply price filter
  if (filters.priceRange) {
    result = result.filter(
      (p) =>
        p.price >= filters.priceRange!.min &&
        p.price <= filters.priceRange!.max
    );
  }

  return result;
}

// Sort logic (extracted from CategoryPage)
function sortProducts(products: Product[], sortBy: SortOption): Product[] {
  const result = [...products];
  result.sort((a, b) => {
    switch (sortBy) {
      case 'price-low':
        return a.price - b.price;
      case 'price-high':
        return b.price - a.price;
      case 'name':
        return a.name.localeCompare(b.name);
      default:
        return 0;
    }
  });
  return result;
}

// Arbitraries
const productArb = fc.record({
  id: fc.integer({ min: 1, max: 10000 }),
  name: fc.string({ minLength: 1, maxLength: 50 }),
  price: fc.integer({ min: 1000, max: 10000000 }),
  subcategory: fc.option(fc.constantFrom('Dress', 'Tops', 'Bottoms', 'Outerwear', 'Accessories'), { nil: undefined }),
  stock: fc.integer({ min: 0, max: 100 }),
});

const productsArb = fc.array(productArb, { minLength: 0, maxLength: 50 });

const priceRangeArb = fc.option(
  fc.record({
    min: fc.integer({ min: 0, max: 500000 }),
    max: fc.integer({ min: 500001, max: 10000000 }),
  }),
  { nil: null }
);

const filtersArb = fc.record({
  sizes: fc.array(fc.constantFrom('XS', 'S', 'M', 'L', 'XL', 'XXL'), { maxLength: 6 }),
  priceRange: priceRangeArb,
  subcategory: fc.option(fc.constantFrom('Dress', 'Tops', 'Bottoms', 'Outerwear', 'Accessories'), { nil: null }),
});

const sortOptionArb = fc.constantFrom<SortOption>('newest', 'price-low', 'price-high', 'name');

describe('Property 1: Filter Results Consistency', () => {
  it('all filtered products should match subcategory filter when applied', () => {
    fc.assert(
      fc.property(productsArb, filtersArb, (products, filters) => {
        const filtered = filterProducts(products, filters);
        
        if (filters.subcategory) {
          return filtered.every(
            (p) => p.subcategory?.toLowerCase() === filters.subcategory?.toLowerCase()
          );
        }
        return true;
      }),
      { numRuns: 100 }
    );
  });

  it('all filtered products should be within price range when applied', () => {
    fc.assert(
      fc.property(productsArb, filtersArb, (products, filters) => {
        const filtered = filterProducts(products, filters);
        
        if (filters.priceRange) {
          return filtered.every(
            (p) => p.price >= filters.priceRange!.min && p.price <= filters.priceRange!.max
          );
        }
        return true;
      }),
      { numRuns: 100 }
    );
  });

  it('filtered results should be a subset of original products', () => {
    fc.assert(
      fc.property(productsArb, filtersArb, (products, filters) => {
        const filtered = filterProducts(products, filters);
        return filtered.every((fp) => products.some((p) => p.id === fp.id));
      }),
      { numRuns: 100 }
    );
  });

  it('clearing all filters should return all products', () => {
    fc.assert(
      fc.property(productsArb, (products) => {
        const emptyFilters: ProductFilters = { sizes: [], priceRange: null, subcategory: null };
        const filtered = filterProducts(products, emptyFilters);
        return filtered.length === products.length;
      }),
      { numRuns: 100 }
    );
  });
});

describe('Property 2: Sort Order Correctness', () => {
  it('price-low sort should order products by ascending price', () => {
    fc.assert(
      fc.property(productsArb, (products) => {
        const sorted = sortProducts(products, 'price-low');
        for (let i = 1; i < sorted.length; i++) {
          if (sorted[i].price < sorted[i - 1].price) {
            return false;
          }
        }
        return true;
      }),
      { numRuns: 100 }
    );
  });

  it('price-high sort should order products by descending price', () => {
    fc.assert(
      fc.property(productsArb, (products) => {
        const sorted = sortProducts(products, 'price-high');
        for (let i = 1; i < sorted.length; i++) {
          if (sorted[i].price > sorted[i - 1].price) {
            return false;
          }
        }
        return true;
      }),
      { numRuns: 100 }
    );
  });

  it('name sort should order products alphabetically', () => {
    fc.assert(
      fc.property(productsArb, (products) => {
        const sorted = sortProducts(products, 'name');
        for (let i = 1; i < sorted.length; i++) {
          if (sorted[i].name.localeCompare(sorted[i - 1].name) < 0) {
            return false;
          }
        }
        return true;
      }),
      { numRuns: 100 }
    );
  });

  it('sorting should preserve all products (no loss)', () => {
    fc.assert(
      fc.property(productsArb, sortOptionArb, (products, sortBy) => {
        const sorted = sortProducts(products, sortBy);
        return sorted.length === products.length;
      }),
      { numRuns: 100 }
    );
  });
});
