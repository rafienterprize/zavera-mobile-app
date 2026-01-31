"use client";

import Link from "next/link";
import { ProductCategory } from "@/types";

interface BreadcrumbItem {
  label: string;
  href?: string;
}

interface BreadcrumbProps {
  items: BreadcrumbItem[];
}

const CATEGORY_LABELS: Record<ProductCategory, string> = {
  wanita: "Wanita",
  pria: "Pria",
  anak: "Anak",
  sports: "Sports",
  luxury: "Luxury",
  beauty: "Beauty",
};

export function getCategoryLabel(category: ProductCategory): string {
  return CATEGORY_LABELS[category] || category;
}

export default function Breadcrumb({ items }: BreadcrumbProps) {
  return (
    <nav className="flex items-center text-sm text-gray-500 mb-4" aria-label="Breadcrumb">
      <ol className="flex items-center flex-wrap gap-1">
        {items.map((item, index) => (
          <li key={index} className="flex items-center">
            {index > 0 && (
              <svg
                className="w-4 h-4 mx-2 text-gray-300"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 5l7 7-7 7"
                />
              </svg>
            )}
            {item.href ? (
              <Link
                href={item.href}
                className="hover:text-primary transition-colors"
              >
                {item.label}
              </Link>
            ) : (
              <span className="text-gray-900 font-medium truncate max-w-[200px]">
                {item.label}
              </span>
            )}
          </li>
        ))}
      </ol>
    </nav>
  );
}

// Helper function to generate breadcrumb items for a product
export function getProductBreadcrumbs(
  productName: string,
  category?: ProductCategory
): BreadcrumbItem[] {
  const items: BreadcrumbItem[] = [{ label: "Home", href: "/" }];

  if (category) {
    items.push({
      label: getCategoryLabel(category),
      href: `/${category}`,
    });
  }

  items.push({ label: productName });

  return items;
}
