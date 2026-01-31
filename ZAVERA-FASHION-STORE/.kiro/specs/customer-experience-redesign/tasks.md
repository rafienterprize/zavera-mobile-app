# Implementation Plan

- [x] 1. Product Filtering System
  - [x] 1.1 Create FilterPanel component with size, price range, and subcategory filters
    - Implement collapsible filter sections
    - Add checkbox groups for sizes (XS, S, M, L, XL)
    - Add price range slider or input fields
    - Add subcategory dropdown based on current category
    - _Requirements: 1.1_

  - [x] 1.2 Implement filter logic in CategoryPage
    - Create useProductFilters hook for filter state management
    - Filter products client-side based on active filters
    - Update URL query params for shareable filter states
    - _Requirements: 1.2_

  - [x] 1.3 Write property test for filter consistency
    - **Property 1: Filter Results Consistency**
    - **Validates: Requirements 1.2**
    - Created: `frontend/src/__tests__/properties/filter.test.ts`

  - [x] 1.4 Implement sort functionality
    - Add sort dropdown with options: Terbaru, Harga Rendah-Tinggi, Harga Tinggi-Rendah, Nama A-Z
    - Implement sort logic for each option
    - _Requirements: 1.3_

  - [x] 1.5 Write property test for sort correctness
    - **Property 2: Sort Order Correctness**
    - **Validates: Requirements 1.3**
    - Created: `frontend/src/__tests__/properties/filter.test.ts`

  - [x] 1.6 Create ActiveFilters component for filter tags
    - Display active filters as removable tags
    - Implement click-to-remove functionality
    - _Requirements: 1.4_

  - [x] 1.7 Create FilterDrawer component for mobile
    - Slide-out drawer containing FilterPanel
    - Add open/close button in category page header
    - _Requirements: 8.2_

- [x] 2. Enhanced Product Detail Page
  - [x] 2.1 Add breadcrumb navigation component
    - Show Home > Category > Product hierarchy
    - Make each level clickable
    - _Requirements: 3.1_

  - [x] 2.2 Write property test for breadcrumb category match
    - **Property 4: Breadcrumb Category Match**
    - **Validates: Requirements 3.1**
    - Created: `frontend/src/__tests__/properties/breadcrumb.test.ts`

  - [x] 2.3 Enhance size selector with better visual feedback
    - Add clear selected state styling
    - Add hover states for unselected sizes
    - Show size guide link
    - _Requirements: 3.3_

  - [x] 2.4 Add low stock indicator
    - Display "Sisa X" badge when stock < 10
    - Use subtle amber/orange styling
    - _Requirements: 3.4_

  - [x] 2.5 Write property test for low stock indicator
    - **Property 5: Low Stock Indicator Display**
    - **Validates: Requirements 3.4**
    - Created: `frontend/src/__tests__/properties/stock.test.ts`

- [x] 3. Order History Enhancement
  - [x] 3.1 Create OrderTimeline component
    - Visual step indicator showing order progression
    - Highlight current status step
    - Show completed steps with checkmarks
    - _Requirements: 6.2_

  - [x] 3.2 Write property test for timeline progression
    - **Property 7: Order Timeline Status Progression**
    - **Validates: Requirements 6.2**
    - Created: `frontend/src/__tests__/properties/order.test.ts`

  - [x] 3.3 Enhance order cards in OrderHistoryPage
    - Improve card layout with better spacing
    - Add product thumbnails in expanded view
    - Integrate OrderTimeline component
    - _Requirements: 6.1, 6.3_

  - [x] 3.4 Write property test for order card completeness
    - **Property 6: Order Card Information Completeness**
    - **Validates: Requirements 6.1**
    - Created: `frontend/src/__tests__/properties/order.test.ts`

  - [x] 3.5 Enhance status badges with consistent styling
    - Define color mapping for all statuses
    - Use Indonesian labels consistently
    - _Requirements: 6.5_

  - [x] 3.6 Write property test for status badge mapping
    - **Property 9: Status Badge Color Mapping**
    - **Validates: Requirements 6.5**
    - Created: `frontend/src/__tests__/properties/order.test.ts`

  - [x] 3.7 Add prominent Pay Now button for pending orders
    - Style as primary action button
    - Link to checkout with order code
    - _Requirements: 6.4_

  - [x] 3.8 Write property test for Pay Now visibility
    - **Property 8: Pending Order Pay Button Visibility**
    - **Validates: Requirements 6.4**
    - Created: `frontend/src/__tests__/properties/order.test.ts`

- [x] 4. Checkout and Cart Polish
  - [x] 4.1 Verify checkout form pre-fill for logged-in users
    - Ensure name, email, phone are pre-populated
    - Test with authenticated user flow
    - _Requirements: 5.1_

  - [x] 4.2 Write property test for checkout pre-fill
    - **Property 11: Checkout Form Pre-fill**
    - **Validates: Requirements 5.1**
    - Created: `frontend/src/__tests__/properties/checkout.test.ts`

  - [x] 4.3 Enhance form validation feedback
    - Add inline error messages below fields
    - Add error styling (red border, icon)
    - Clear errors on valid input
    - _Requirements: 7.2_

  - [x] 4.4 Write property test for validation error display
    - **Property 10: Form Validation Error Display**
    - **Validates: Requirements 7.2**
    - Created: `frontend/src/__tests__/properties/checkout.test.ts`

  - [x] 4.5 Verify cart total calculation
    - Ensure subtotal, shipping, and total are correct
    - Test with various cart configurations
    - _Requirements: 4.2, 4.5_

  - [x] 4.6 Write property test for cart total
    - **Property 3: Cart Total Calculation**
    - **Validates: Requirements 4.2, 4.5**
    - Created: `frontend/src/__tests__/properties/cart.test.ts`

- [x] 5. Loading States and Visual Polish
  - [x] 5.1 Verify skeleton loaders are consistent across pages
    - Check CategoryPage, ProductDetail, OrderHistory
    - Ensure shapes match actual content
    - _Requirements: 2.1_

  - [x] 5.2 Add error states with retry functionality
    - Create reusable ErrorState component
    - Add retry button that re-fetches data
    - _Requirements: 2.4_

  - [x] 5.3 Ensure all buttons show loading states during async operations
    - Add to cart button
    - Checkout button
    - Filter apply button
    - _Requirements: 7.4_

- [x] 6. Mobile Responsiveness
  - [x] 6.1 Verify 2-column product grid on mobile
    - Test at various mobile breakpoints
    - Ensure proper spacing and alignment
    - _Requirements: 8.1_

  - [x] 6.2 Test mobile navigation and category access
    - Verify hamburger menu works
    - Ensure all categories are accessible
    - _Requirements: 8.3_

  - [x] 6.3 Test mobile checkout experience
    - Verify form is usable on small screens
    - Test order summary visibility
    - _Requirements: 8.4_

- [x] 7. Checkpoint - Ensure all tests pass
  - All 56 property tests pass (6 test files)

- [x] 8. Final Integration and Testing
  - [x] 8.1 Test complete shopping flow
    - Browse category > Filter > Select product > Add to cart > Checkout > View order
    - Verify all transitions are smooth
    - _Requirements: All_

  - [x] 8.2 Test order history flow
    - View orders > Expand details > Pay pending order
    - Verify timeline and status display
    - _Requirements: 6.1-6.5_

- [x] 9. Push to GitHub
  - [x] 9.1 Commit all changes with descriptive message
    - Include summary of UX improvements
  
  - [x] 9.2 Push to remote repository main branch

- [x] 10. Final Checkpoint - Ensure all tests pass
  - All 56 property tests pass (6 test files) ✓
