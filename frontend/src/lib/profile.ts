export type ProfileField = 'name' | 'email' | 'phone';

/** Only same-site paths are accepted as `next`, so the page can't redirect elsewhere. */
export function safeNext(next: string | null): string | null {
  if (!next || !next.startsWith('/')) return null;
  // "//host" and "/\host" are both read by browsers as another site.
  if (next.startsWith('//') || next.startsWith('/\\')) return null;
  return next;
}

/**
 * Profile errors (400 validation, 409 duplicates) all start with the field
 * they refer to, e.g. "el teléfono ya está registrado".
 */
export function fieldOfError(message: string): ProfileField | null {
  const m = message.toLowerCase();
  if (m.startsWith('el nombre')) return 'name';
  if (m.startsWith('el email')) return 'email';
  if (m.startsWith('el teléfono')) return 'phone';
  return null;
}
