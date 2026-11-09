import { describe, expect, it } from 'vitest';
import { fieldOfError, safeNext } from './profile';

describe('safeNext', () => {
  it('accepts same-site paths', () => {
    expect(safeNext('/mis-reservas')).toBe('/mis-reservas');
    expect(safeNext('/espacios/3?fecha=2026-11-10')).toBe('/espacios/3?fecha=2026-11-10');
  });

  it.each([null, '', 'https://evil.com', '//evil.com', '/\\evil.com', 'mis-reservas'])(
    'rejects %j',
    (next) => {
      expect(safeNext(next)).toBeNull();
    },
  );
});

describe('fieldOfError', () => {
  it.each([
    ['El nombre es requerido', 'name'],
    ['el email no es válido', 'email'],
    ['El teléfono ya está registrado', 'phone'],
  ] as const)('%s -> %s', (message, field) => {
    expect(fieldOfError(message)).toBe(field);
  });

  it('returns null for errors that are not about a field', () => {
    expect(fieldOfError('Algo salió mal')).toBeNull();
  });
});
