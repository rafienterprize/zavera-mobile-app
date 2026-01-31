/**
 * Property Tests for Breadcrumb Navigation
 * Property 4: Breadcrumb Category Match
 */
import { describe, it, expect } from 'vitest';
import * as fc from 'fast-check';

// Types
type ProductCategory = 'wanita' | 'pria' | 'anak' | 'sports' | 'luxury' | 'beauty';

interface BreadcrumbItem {
  label: string;
  href?: string;
}

// Category labels mapping (from Breadcrumb.tsx)
const CATEGORY_LABELS: Record<ProductCategory, string> = {
  wanita: 'Wanita',
  pria: 'Pria',
  anak: 'Anak',
  sports: 'Sports',
  luxury: 'Luxury',
  beauty: 'Beauty',
};

// Breadcrumb generation logic (from Breadcrumb.tsx)
function getCategoryLabel(category: ProductCategory): string {
  return CATEGORY_LABELS[category] || category;
}

function getProductBreadcrumbs(
  productName: string,
  category?: ProductCategory
): BreadcrumbItem[] {
  const items: BreadcrumbItem[] = [{ label: 'Home', href: '/' }];

  if (category) {
    items.push({
      label: getCategoryLabel(category),
      href: `/${category}`,
    });
  }

  items.push({ label: productName });

  return items;
}

// Arbitraries
const categoryArb = fc.constantFrom<ProductCategory>('wanita', 'pria', 'anak', 'sports', 'luxury', 'beauty');
const productNameArb = fc.string({ minLength: 1, maxLength: 100 });

describe('Property 4: Breadcrumb Category Match', () => {
  it('breadcrumb should include category when product has a category', () => {
    fc.assert(
      fc.property(productNameArb, categoryArb, (productName, category) => {
        const breadcrumbs = getProductBreadcrumbs(productName, category);
        const categoryLabel = getCategoryLabel(category);
        return breadcrumbs.some((item) => item.label === categoryLabel);
      }),
      { numRuns: 100 }
    );
  });

  it('breadcrumb should always start with Home', () => {
    fc.assert(
      fc.property(productNameArb, fc.option(categoryArb, { nil: undefined }), (productName, category) => {
        const breadcrumbs = getProductBreadcrumbs(productName, category ?? undefined);
        return breadcrumbs[0].label === 'Home' && breadcrumbs[0].href === '/';
      }),
      { numRuns: 100 }
    );
  });

  it('breadcrumb should always end with product name', () => {
    fc.assert(
      fc.property(productNameArb, fc.option(categoryArb, { nil: undefined }), (productName, category) => {
        const breadcrumbs = getProductBreadcrumbs(productName, category ?? undefined);
        const lastItem = breadcrumbs[breadcrumbs.length - 1];
        return lastItem.label === productName && lastItem.href === undefined;
      }),
      { numRuns: 100 }
    );
  });

  it('breadcrumb with category should have 3 items', () => {
    fc.assert(
      fc.property(productNameArb, categoryArb, (productName, category) => {
        const breadcrumbs = getProductBreadcrumbs(productName, category);
        return breadcrumbs.length === 3;
      }),
      { numRuns: 100 }
    );
  });

  it('breadcrumb without category should have 2 items', () => {
    fc.assert(
      fc.property(productNameArb, (productName) => {
        const breadcrumbs = getProductBreadcrumbs(productName, undefined);
        return breadcrumbs.length === 2;
      }),
      { numRuns: 100 }
    );
  });

  it('category href should match category slug', () => {
    fc.assert(
      fc.property(productNameArb, categoryArb, (productName, category) => {
        const breadcrumbs = getProductBreadcrumbs(productName, category);
        const categoryItem = breadcrumbs.find((item) => item.label === getCategoryLabel(category));
        return categoryItem?.href === `/${category}`;
      }),
      { numRuns: 100 }
    );
  });
});
