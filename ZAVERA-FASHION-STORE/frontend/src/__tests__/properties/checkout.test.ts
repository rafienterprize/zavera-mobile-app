/**
 * Property Tests for Checkout Form
 * Properties 10 & 11: Form Validation and Pre-fill
 */
import { describe, it, expect } from 'vitest';
import * as fc from 'fast-check';

// Types
interface User {
  id: number;
  name: string;
  email: string;
  phone?: string;
}

interface CheckoutFormData {
  name: string;
  email: string;
  phone: string;
  address: string;
}

interface ValidationError {
  field: string;
  message: string;
}

// Validation logic
function validateEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}

function validatePhone(phone: string): boolean {
  // Indonesian phone format: starts with 08 or +62, 10-13 digits
  const phoneRegex = /^(\+62|62|0)8[1-9][0-9]{7,10}$/;
  return phoneRegex.test(phone.replace(/[\s-]/g, ''));
}

function validateName(name: string): boolean {
  return name.trim().length >= 2;
}

function validateAddress(address: string): boolean {
  return address.trim().length >= 10;
}

function validateCheckoutForm(data: CheckoutFormData): ValidationError[] {
  const errors: ValidationError[] = [];

  if (!validateName(data.name)) {
    errors.push({ field: 'name', message: 'Nama harus minimal 2 karakter' });
  }

  if (!validateEmail(data.email)) {
    errors.push({ field: 'email', message: 'Format email tidak valid' });
  }

  if (!validatePhone(data.phone)) {
    errors.push({ field: 'phone', message: 'Format nomor telepon tidak valid' });
  }

  if (!validateAddress(data.address)) {
    errors.push({ field: 'address', message: 'Alamat harus minimal 10 karakter' });
  }

  return errors;
}

// Pre-fill logic
function prefillCheckoutForm(user: User | null): Partial<CheckoutFormData> {
  if (!user) {
    return {};
  }

  return {
    name: user.name || '',
    email: user.email || '',
    phone: user.phone || '',
  };
}

function hasFieldError(errors: ValidationError[], field: string): boolean {
  return errors.some((e) => e.field === field);
}

function getFieldError(errors: ValidationError[], field: string): string | null {
  const error = errors.find((e) => e.field === field);
  return error?.message || null;
}

// Arbitraries
const validEmailArb = fc.emailAddress();
const invalidEmailArb = fc.oneof(
  fc.constant('invalid'),
  fc.constant('no@domain'),
  fc.constant('@nodomain.com'),
  fc.constant('spaces in@email.com')
);

const validPhoneArb = fc.oneof(
  fc.stringMatching(/^08[1-9][0-9]{8,10}$/),
  fc.stringMatching(/^\+628[1-9][0-9]{8,10}$/)
);

const invalidPhoneArb = fc.oneof(
  fc.constant('123'),
  fc.constant('abcdefghij'),
  fc.constant('0712345678') // doesn't start with 08
);

const validNameArb = fc.string({ minLength: 2, maxLength: 100 }).filter((s) => s.trim().length >= 2);
const invalidNameArb = fc.oneof(fc.constant(''), fc.constant(' '), fc.constant('A'));

const validAddressArb = fc.string({ minLength: 10, maxLength: 200 }).filter((s) => s.trim().length >= 10);
const invalidAddressArb = fc.string({ minLength: 0, maxLength: 9 });

const userArb = fc.record({
  id: fc.integer({ min: 1, max: 10000 }),
  name: validNameArb,
  email: validEmailArb,
  phone: fc.option(validPhoneArb, { nil: undefined }),
});

describe('Property 10: Form Validation Error Display', () => {
  it('valid form data should produce no errors', () => {
    fc.assert(
      fc.property(validNameArb, validEmailArb, validPhoneArb, validAddressArb, (name, email, phone, address) => {
        const errors = validateCheckoutForm({ name, email, phone, address });
        return errors.length === 0;
      }),
      { numRuns: 100 }
    );
  });

  it('invalid email should produce email error', () => {
    fc.assert(
      fc.property(validNameArb, invalidEmailArb, validPhoneArb, validAddressArb, (name, email, phone, address) => {
        const errors = validateCheckoutForm({ name, email, phone, address });
        return hasFieldError(errors, 'email');
      }),
      { numRuns: 100 }
    );
  });

  it('invalid phone should produce phone error', () => {
    fc.assert(
      fc.property(validNameArb, validEmailArb, invalidPhoneArb, validAddressArb, (name, email, phone, address) => {
        const errors = validateCheckoutForm({ name, email, phone, address });
        return hasFieldError(errors, 'phone');
      }),
      { numRuns: 100 }
    );
  });

  it('invalid name should produce name error', () => {
    fc.assert(
      fc.property(invalidNameArb, validEmailArb, validPhoneArb, validAddressArb, (name, email, phone, address) => {
        const errors = validateCheckoutForm({ name, email, phone, address });
        return hasFieldError(errors, 'name');
      }),
      { numRuns: 100 }
    );
  });

  it('invalid address should produce address error', () => {
    fc.assert(
      fc.property(validNameArb, validEmailArb, validPhoneArb, invalidAddressArb, (name, email, phone, address) => {
        const errors = validateCheckoutForm({ name, email, phone, address });
        return hasFieldError(errors, 'address');
      }),
      { numRuns: 100 }
    );
  });

  it('each error should have a non-empty message', () => {
    fc.assert(
      fc.property(invalidNameArb, invalidEmailArb, invalidPhoneArb, invalidAddressArb, (name, email, phone, address) => {
        const errors = validateCheckoutForm({ name, email, phone, address });
        return errors.every((e) => e.message.length > 0);
      }),
      { numRuns: 100 }
    );
  });

  it('validation should be deterministic', () => {
    fc.assert(
      fc.property(
        fc.string({ maxLength: 50 }),
        fc.string({ maxLength: 50 }),
        fc.string({ maxLength: 20 }),
        fc.string({ maxLength: 100 }),
        (name, email, phone, address) => {
          const errors1 = validateCheckoutForm({ name, email, phone, address });
          const errors2 = validateCheckoutForm({ name, email, phone, address });
          return JSON.stringify(errors1) === JSON.stringify(errors2);
        }
      ),
      { numRuns: 100 }
    );
  });
});

describe('Property 11: Checkout Form Pre-fill', () => {
  it('logged-in user should have name pre-filled', () => {
    fc.assert(
      fc.property(userArb, (user) => {
        const prefilled = prefillCheckoutForm(user);
        return prefilled.name === user.name;
      }),
      { numRuns: 100 }
    );
  });

  it('logged-in user should have email pre-filled', () => {
    fc.assert(
      fc.property(userArb, (user) => {
        const prefilled = prefillCheckoutForm(user);
        return prefilled.email === user.email;
      }),
      { numRuns: 100 }
    );
  });

  it('logged-in user with phone should have phone pre-filled', () => {
    fc.assert(
      fc.property(
        fc.record({
          id: fc.integer({ min: 1, max: 10000 }),
          name: validNameArb,
          email: validEmailArb,
          phone: validPhoneArb,
        }),
        (user) => {
          const prefilled = prefillCheckoutForm(user);
          return prefilled.phone === user.phone;
        }
      ),
      { numRuns: 100 }
    );
  });

  it('guest user (null) should have empty pre-fill', () => {
    const prefilled = prefillCheckoutForm(null);
    expect(Object.keys(prefilled).length).toBe(0);
  });

  it('pre-fill should not include address (user-specific)', () => {
    fc.assert(
      fc.property(userArb, (user) => {
        const prefilled = prefillCheckoutForm(user);
        return prefilled.address === undefined;
      }),
      { numRuns: 100 }
    );
  });

  it('pre-fill should be deterministic', () => {
    fc.assert(
      fc.property(userArb, (user) => {
        const prefilled1 = prefillCheckoutForm(user);
        const prefilled2 = prefillCheckoutForm(user);
        return JSON.stringify(prefilled1) === JSON.stringify(prefilled2);
      }),
      { numRuns: 100 }
    );
  });
});
