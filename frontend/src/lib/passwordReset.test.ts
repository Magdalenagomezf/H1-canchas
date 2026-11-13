import { describe, expect, it } from 'vitest';
import { newPasswordError, readResetToken } from './passwordReset';

describe('readResetToken', () => {
  it('reads the token from the URL fragment', () => {
    expect(readResetToken('#token=abc-DEF_123')).toBe('abc-DEF_123');
  });

  it('ignores other fragment parameters', () => {
    expect(readResetToken('#foo=1&token=abc')).toBe('abc');
  });

  it.each(['', '#', '#token=', '#other=abc', 'token=abc'])('returns null for %j', (hash) => {
    expect(readResetToken(hash)).toBeNull();
  });
});

describe('newPasswordError', () => {
  it('accepts a valid password typed twice', () => {
    expect(newPasswordError('secreta1', 'secreta1')).toBeNull();
  });

  it('asks for at least 6 characters', () => {
    expect(newPasswordError('12345', '12345')).toBe('La contraseña debe tener al menos 6 caracteres');
  });

  it('rejects more than 72 bytes, like the backend', () => {
    expect(newPasswordError('a'.repeat(72), 'a'.repeat(72))).toBeNull();
    expect(newPasswordError('a'.repeat(73), 'a'.repeat(73))).toBe('La contraseña no puede superar los 72 caracteres');
    // 37 x "ñ" are 37 characters but 74 bytes.
    expect(newPasswordError('ñ'.repeat(37), 'ñ'.repeat(37))).toBe('La contraseña no puede superar los 72 caracteres');
  });

  it('asks both passwords to match', () => {
    expect(newPasswordError('secreta1', 'secreta2')).toBe('Las contraseñas no coinciden');
  });
});
