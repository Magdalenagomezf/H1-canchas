import type { ReactNode } from 'react';
import { GoogleOAuthProvider } from '@react-oauth/google';
import { GOOGLE_CLIENT_ID } from '@/lib/google';

/** Loads Google Identity Services only when a client ID is configured; otherwise renders children as-is. */
export function GoogleProvider({ children }: { children: ReactNode }) {
  return GOOGLE_CLIENT_ID ? (
    <GoogleOAuthProvider clientId={GOOGLE_CLIENT_ID} locale="es">
      {children}
    </GoogleOAuthProvider>
  ) : (
    <>{children}</>
  );
}
