/**
 * Reads the reset token from the URL fragment (`#token=...`). The email link
 * carries it in the fragment so it never reaches server logs or Referer headers.
 */
export function readResetToken(hash: string): string | null {
  if (!hash.startsWith('#')) return null;
  const token = new URLSearchParams(hash.slice(1)).get('token');
  return token ? token : null;
}

// Same rule as the backend (validateNewPassword): 6 characters minimum and
// 72 bytes maximum, the bcrypt limit.
const MIN_LENGTH = 6;
const MAX_BYTES = 72;

/** Returns the message to show for the new password, or null when it is valid. */
export function newPasswordError(password: string, repeat: string): string | null {
  if (password.length < MIN_LENGTH) return `La contraseña debe tener al menos ${MIN_LENGTH} caracteres`;
  if (new TextEncoder().encode(password).length > MAX_BYTES) {
    return `La contraseña no puede superar los ${MAX_BYTES} caracteres`;
  }
  if (password !== repeat) return 'Las contraseñas no coinciden';
  return null;
}
