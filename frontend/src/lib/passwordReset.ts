/**
 * Reads the reset token from the URL fragment (`#token=...`). The email link
 * carries it in the fragment so it never reaches server logs or Referer headers.
 */
export function readResetToken(hash: string): string | null {
  if (!hash.startsWith('#')) return null;
  const token = new URLSearchParams(hash.slice(1)).get('token');
  return token ? token : null;
}

// Same rule as the backend (validateNewPassword), which counts UTF-8 bytes
// like Go's len(): at least 6 and at most 72, the bcrypt limit. "ñ" or an
// accented letter is 2 bytes, so the messages avoid promising a character count
// for the maximum.
const MIN_BYTES = 6;
const MAX_BYTES = 72;

/** Returns the message to show for the new password, or null when it is valid. */
export function newPasswordError(password: string, repeat: string): string | null {
  const bytes = new TextEncoder().encode(password).length;
  if (bytes < MIN_BYTES) return `La contraseña debe tener al menos ${MIN_BYTES} caracteres`;
  if (bytes > MAX_BYTES) return 'La contraseña es demasiado larga';
  if (password !== repeat) return 'Las contraseñas no coinciden';
  return null;
}
