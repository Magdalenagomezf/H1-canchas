import { createContext, useContext, useState, useCallback, type ReactNode } from 'react';
import type { User } from '../types';
import queryClient from '../lib/queryClient';

interface AuthContextValue {
  user: User | null;
  token: string | null;
  setAuth: (user: User, token: string) => void;
  logout: () => void;
  isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextValue | null>(null);

function getStoredUser(): User | null {
  try {
    const raw = sessionStorage.getItem('user');
    return raw ? (JSON.parse(raw) as User) : null;
  } catch {
    return null;
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(getStoredUser);
  const [token, setToken] = useState<string | null>(() => sessionStorage.getItem('token'));

  const setAuth = useCallback((newUser: User, newToken: string) => {
    sessionStorage.setItem('token', newToken);
    sessionStorage.setItem('user', JSON.stringify(newUser));
    setUser(newUser);
    setToken(newToken);
  }, []);

  const logout = useCallback(() => {
    sessionStorage.removeItem('token');
    sessionStorage.removeItem('user');
    queryClient.clear();
    setUser(null);
    setToken(null);
  }, []);

  return (
    <AuthContext.Provider value={{ user, token, setAuth, logout, isAuthenticated: !!token }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used inside AuthProvider');
  return ctx;
}
