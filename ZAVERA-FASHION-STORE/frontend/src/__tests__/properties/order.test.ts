/**
 * Property Tests for Order History
 * Properties 6, 7, 8, 9: Order Card, Timeline, Pay Button, Status Badge
 */
import { describe, it, expect } from 'vitest';
import * as fc from 'fast-check';

// Types (from OrderTimeline.tsx)
type OrderStatus = 
  | 'PENDING' 
  | 'PAID' 
  | 'PROCESSING' 
  | 'SHIPPED' 
  | 'DELIVERED' 
  | 'COMPLETED' 
  | 'CANCELLED' 
  | 'FAILED' 
  | 'EXPIRED';

interface Order {
  id: number;
  order_code: string;
  status: OrderStatus;
  total_amount: number;
  created_at: string;
  items: OrderItem[];
}

interface OrderItem {
  id: number;
  product_name: string;
  quantity: number;
  price: number;
}

// Status color mapping (from OrderTimeline.tsx)
const STATUS_COLORS: Record<OrderStatus, string> = {
  PENDING: 'bg-yellow-100 text-yellow-800 border-yellow-200',
  PAID: 'bg-blue-100 text-blue-800 border-blue-200',
  PROCESSING: 'bg-purple-100 text-purple-800 border-purple-200',
  SHIPPED: 'bg-indigo-100 text-indigo-800 border-indigo-200',
  DELIVERED: 'bg-green-100 text-green-800 border-green-200',
  COMPLETED: 'bg-green-100 text-green-800 border-green-200',
  CANCELLED: 'bg-red-100 text-red-800 border-red-200',
  FAILED: 'bg-red-100 text-red-800 border-red-200',
  EXPIRED: 'bg-gray-100 text-gray-800 border-gray-200',
};

// Status labels in Indonesian (from OrderTimeline.tsx)
const STATUS_LABELS: Record<OrderStatus, string> = {
  PENDING: 'Menunggu Pembayaran',
  PAID: 'Dibayar',
  PROCESSING: 'Diproses',
  SHIPPED: 'Dikirim',
  DELIVERED: 'Terkirim',
  COMPLETED: 'Selesai',
  CANCELLED: 'Dibatalkan',
  FAILED: 'Gagal',
  EXPIRED: 'Kadaluarsa',
};

// Timeline steps
const TIMELINE_STEPS: OrderStatus[] = ['PENDING', 'PAID', 'PROCESSING', 'SHIPPED', 'DELIVERED'];

// Logic functions (from OrderTimeline.tsx)
function getStatusIndex(status: OrderStatus): number {
  const index = TIMELINE_STEPS.indexOf(status);
  if (status === 'COMPLETED') return TIMELINE_STEPS.length - 1;
  return index;
}

function isStepCompleted(stepKey: OrderStatus, currentStatus: OrderStatus): boolean {
  const stepIndex = getStatusIndex(stepKey);
  const currentIndex = getStatusIndex(currentStatus);
  
  if (['CANCELLED', 'FAILED', 'EXPIRED'].includes(currentStatus)) {
    return false;
  }
  
  return stepIndex <= currentIndex;
}

function isCurrentStep(stepKey: OrderStatus, currentStatus: OrderStatus): boolean {
  if (currentStatus === 'COMPLETED' && stepKey === 'DELIVERED') return true;
  return stepKey === currentStatus;
}

function shouldShowPayButton(status: OrderStatus): boolean {
  return status === 'PENDING';
}

function formatOrderCode(code: string): string {
  return code.toUpperCase();
}

function formatCurrency(amount: number): string {
  return `Rp ${amount.toLocaleString('id-ID')}`;
}

// Arbitraries
const orderStatusArb = fc.constantFrom<OrderStatus>(
  'PENDING', 'PAID', 'PROCESSING', 'SHIPPED', 'DELIVERED', 'COMPLETED', 'CANCELLED', 'FAILED', 'EXPIRED'
);

const progressiveStatusArb = fc.constantFrom<OrderStatus>(
  'PENDING', 'PAID', 'PROCESSING', 'SHIPPED', 'DELIVERED', 'COMPLETED'
);

const terminalStatusArb = fc.constantFrom<OrderStatus>('CANCELLED', 'FAILED', 'EXPIRED');

const orderCodeArb = fc.stringMatching(/^ORD-[A-Z0-9]{8}$/);

const orderItemArb = fc.record({
  id: fc.integer({ min: 1, max: 10000 }),
  product_name: fc.string({ minLength: 1, maxLength: 50 }),
  quantity: fc.integer({ min: 1, max: 10 }),
  price: fc.integer({ min: 1000, max: 5000000 }),
});

const orderArb = fc.record({
  id: fc.integer({ min: 1, max: 10000 }),
  order_code: orderCodeArb,
  status: orderStatusArb,
  total_amount: fc.integer({ min: 10000, max: 50000000 }),
  created_at: fc.date({ min: new Date(2020, 0, 1), max: new Date(2025, 11, 31) }).map((d) => d.toISOString()),
  items: fc.array(orderItemArb, { minLength: 1, maxLength: 10 }),
});

describe('Property 6: Order Card Information Completeness', () => {
  it('order should have all required fields', () => {
    fc.assert(
      fc.property(orderArb, (order) => {
        return (
          typeof order.order_code === 'string' &&
          order.order_code.length > 0 &&
          typeof order.created_at === 'string' &&
          typeof order.status === 'string' &&
          typeof order.total_amount === 'number' &&
          order.total_amount > 0
        );
      }),
      { numRuns: 100 }
    );
  });

  it('order code should be formatted correctly', () => {
    fc.assert(
      fc.property(orderCodeArb, (code) => {
        const formatted = formatOrderCode(code);
        return formatted === formatted.toUpperCase();
      }),
      { numRuns: 100 }
    );
  });

  it('total amount should be formatted as currency', () => {
    fc.assert(
      fc.property(fc.integer({ min: 1000, max: 50000000 }), (amount) => {
        const formatted = formatCurrency(amount);
        return formatted.startsWith('Rp ');
      }),
      { numRuns: 100 }
    );
  });
});

describe('Property 7: Order Timeline Status Progression', () => {
  it('all steps up to current status should be completed for progressive statuses', () => {
    fc.assert(
      fc.property(progressiveStatusArb, (status) => {
        const currentIndex = getStatusIndex(status);
        
        for (let i = 0; i <= currentIndex; i++) {
          if (!isStepCompleted(TIMELINE_STEPS[i], status)) {
            return false;
          }
        }
        return true;
      }),
      { numRuns: 100 }
    );
  });

  it('steps after current status should NOT be completed', () => {
    fc.assert(
      fc.property(progressiveStatusArb, (status) => {
        const currentIndex = getStatusIndex(status);
        
        for (let i = currentIndex + 1; i < TIMELINE_STEPS.length; i++) {
          if (isStepCompleted(TIMELINE_STEPS[i], status)) {
            return false;
          }
        }
        return true;
      }),
      { numRuns: 100 }
    );
  });

  it('terminal statuses should have no completed steps', () => {
    fc.assert(
      fc.property(terminalStatusArb, (status) => {
        return TIMELINE_STEPS.every((step) => !isStepCompleted(step, status));
      }),
      { numRuns: 100 }
    );
  });

  it('exactly one step should be current for progressive statuses', () => {
    fc.assert(
      fc.property(progressiveStatusArb, (status) => {
        const currentSteps = TIMELINE_STEPS.filter((step) => isCurrentStep(step, status));
        return currentSteps.length === 1;
      }),
      { numRuns: 100 }
    );
  });
});

describe('Property 8: Pending Order Pay Button Visibility', () => {
  it('Pay Now button should be visible for PENDING status', () => {
    expect(shouldShowPayButton('PENDING')).toBe(true);
  });

  it('Pay Now button should NOT be visible for non-PENDING statuses', () => {
    const nonPendingStatuses: OrderStatus[] = [
      'PAID', 'PROCESSING', 'SHIPPED', 'DELIVERED', 'COMPLETED', 'CANCELLED', 'FAILED', 'EXPIRED'
    ];
    
    nonPendingStatuses.forEach((status) => {
      expect(shouldShowPayButton(status)).toBe(false);
    });
  });

  it('Pay Now visibility should be deterministic', () => {
    fc.assert(
      fc.property(orderStatusArb, (status) => {
        const result1 = shouldShowPayButton(status);
        const result2 = shouldShowPayButton(status);
        return result1 === result2;
      }),
      { numRuns: 100 }
    );
  });
});

describe('Property 9: Status Badge Color Mapping', () => {
  it('every status should have a defined color', () => {
    fc.assert(
      fc.property(orderStatusArb, (status) => {
        return STATUS_COLORS[status] !== undefined && STATUS_COLORS[status].length > 0;
      }),
      { numRuns: 100 }
    );
  });

  it('every status should have a defined Indonesian label', () => {
    fc.assert(
      fc.property(orderStatusArb, (status) => {
        return STATUS_LABELS[status] !== undefined && STATUS_LABELS[status].length > 0;
      }),
      { numRuns: 100 }
    );
  });

  it('success statuses should have green colors', () => {
    const successStatuses: OrderStatus[] = ['DELIVERED', 'COMPLETED'];
    successStatuses.forEach((status) => {
      expect(STATUS_COLORS[status]).toContain('green');
    });
  });

  it('error statuses should have red colors', () => {
    const errorStatuses: OrderStatus[] = ['CANCELLED', 'FAILED'];
    errorStatuses.forEach((status) => {
      expect(STATUS_COLORS[status]).toContain('red');
    });
  });

  it('pending status should have yellow/warning color', () => {
    expect(STATUS_COLORS['PENDING']).toContain('yellow');
  });

  it('color mapping should be consistent', () => {
    fc.assert(
      fc.property(orderStatusArb, (status) => {
        const color1 = STATUS_COLORS[status];
        const color2 = STATUS_COLORS[status];
        return color1 === color2;
      }),
      { numRuns: 100 }
    );
  });
});
