import { useState } from 'react';
import { useNavigate, useLocation, type Location } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import { Eye, EyeOff, Phone, Lock, User } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { login, register } from '../api/auth';
import { useAuth } from '../hooks/useAuth';
import { cn } from '@/lib/utils';

const inputClass =
  'w-full py-3 bg-surface border-[1.5px] border-black/10 rounded-lg text-sm font-medium text-ink placeholder:text-ink-2/60 outline-none transition-all duration-normal focus:border-primary focus:ring-2 focus:ring-primary/[0.13]';

export default function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { setAuth } = useAuth();

  const state = location.state as { from?: Location; tab?: string; pendingBooking?: { date: string; slotId: number } } | null;
  const redirectTo = state?.from
    ? state.from.pathname + (state.from.search ?? '')
    : '/';
  const pendingBooking = state?.pendingBooking;

  const [tab, setTab] = useState<0 | 1>(state?.tab === 'register' ? 1 : 0);

  const [loginPhone, setLoginPhone] = useState('');
  const [loginPassword, setLoginPassword] = useState('');
  const [showLoginPass, setShowLoginPass] = useState(false);

  const [regName, setRegName] = useState('');
  const [regPhone, setRegPhone] = useState('');
  const [regPassword, setRegPassword] = useState('');
  const [showRegPass, setShowRegPass] = useState(false);

  const loginMutation = useMutation({
    mutationFn: () => login(loginPhone, loginPassword),
    onSuccess: ({ user, token }) => {
      setAuth(user, token);
      navigate(redirectTo, { replace: true, state: pendingBooking ? { pendingBooking } : undefined });
    },
  });

  const registerMutation = useMutation({
    mutationFn: () => register(regName, regPhone, regPassword),
    onSuccess: ({ user, token }) => {
      setAuth(user, token);
      navigate(redirectTo, { replace: true, state: pendingBooking ? { pendingBooking } : undefined });
    },
  });

  const handleTabChange = (next: 0 | 1) => {
    setTab(next);
    loginMutation.reset();
    registerMutation.reset();
  };

  const error = tab === 0 ? loginMutation.error?.message : registerMutation.error?.message;
  const isPending = tab === 0 ? loginMutation.isPending : registerMutation.isPending;

  return (
    <div className="animate-fade-up min-h-[calc(100vh-58px)] bg-bg flex items-center justify-center px-4 py-16">
      <div className="w-full max-w-md">

        {/* Header */}
        <div className="text-center mb-8">
          <h1 className="font-serif text-4xl text-ink tracking-tight mb-1.5">
            {tab === 0 ? 'Bienvenido' : 'Crear cuenta'}
          </h1>
          <p className="text-sm text-ink-2">
            {tab === 0
              ? 'Ingresá para gestionar tus reservas'
              : 'Es gratis y lleva menos de un minuto'}
          </p>
        </div>

        {/* Card */}
        <div className="bg-white rounded-xl border-[1.5px] border-black/[0.07] shadow-md p-6">

          {/* Tab switcher */}
          <div className="flex bg-surface rounded-lg p-1 mb-6">
            {(['Iniciar sesión', 'Registrarme'] as const).map((label, i) => (
              <button
                key={label}
                onClick={() => handleTabChange(i as 0 | 1)}
                className={cn(
                  'flex-1 py-2 text-sm font-semibold rounded-md transition-all duration-normal',
                  tab === i
                    ? 'bg-white text-ink shadow-sm'
                    : 'text-ink-2 hover:text-ink',
                )}
              >
                {label}
              </button>
            ))}
          </div>

          <AnimatePresence mode="wait">
            {tab === 0 ? (
              <motion.form
                key="login"
                initial={{ opacity: 0, x: -10 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: 10 }}
                transition={{ duration: 0.18 }}
                onSubmit={(e) => { e.preventDefault(); loginMutation.mutate(); }}
                className="flex flex-col gap-4"
              >
                {/* Phone */}
                <div className="relative">
                  <Phone size={15} className="absolute left-3.5 top-1/2 -translate-y-1/2 text-ink-2/50 pointer-events-none" />
                  <input
                    type="tel"
                    placeholder="Teléfono"
                    value={loginPhone}
                    onChange={(e) => setLoginPhone(e.target.value)}
                    required
                    className={cn(inputClass, 'pl-10 pr-4')}
                  />
                </div>

                {/* Password */}
                <div className="relative">
                  <Lock size={15} className="absolute left-3.5 top-1/2 -translate-y-1/2 text-ink-2/50 pointer-events-none" />
                  <input
                    type={showLoginPass ? 'text' : 'password'}
                    placeholder="Contraseña"
                    value={loginPassword}
                    onChange={(e) => setLoginPassword(e.target.value)}
                    required
                    className={cn(inputClass, 'pl-10 pr-11')}
                  />
                  <button
                    type="button"
                    onClick={() => setShowLoginPass((v) => !v)}
                    className="absolute right-3.5 top-1/2 -translate-y-1/2 text-ink-2/50 hover:text-ink-2 transition-colors"
                  >
                    {showLoginPass ? <EyeOff size={15} /> : <Eye size={15} />}
                  </button>
                </div>

                {error && <p className="text-xs text-status-cancelled font-medium">{error}</p>}

                <button
                  type="submit"
                  disabled={isPending}
                  className="bg-primary text-white font-bold rounded-lg px-6 py-3 text-sm transition-all duration-normal ease-smooth hover:bg-primary-dark hover:shadow-glow active:scale-[0.98] disabled:opacity-60 disabled:cursor-not-allowed mt-1"
                >
                  {isPending ? 'Ingresando...' : 'Ingresar'}
                </button>
              </motion.form>
            ) : (
              <motion.form
                key="register"
                initial={{ opacity: 0, x: 10 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -10 }}
                transition={{ duration: 0.18 }}
                onSubmit={(e) => { e.preventDefault(); registerMutation.mutate(); }}
                className="flex flex-col gap-4"
              >
                {/* Name */}
                <div className="relative">
                  <User size={15} className="absolute left-3.5 top-1/2 -translate-y-1/2 text-ink-2/50 pointer-events-none" />
                  <input
                    type="text"
                    placeholder="Nombre completo"
                    value={regName}
                    onChange={(e) => setRegName(e.target.value)}
                    required
                    className={cn(inputClass, 'pl-10 pr-4')}
                  />
                </div>

                {/* Phone */}
                <div className="relative">
                  <Phone size={15} className="absolute left-3.5 top-1/2 -translate-y-1/2 text-ink-2/50 pointer-events-none" />
                  <input
                    type="tel"
                    placeholder="Teléfono"
                    value={regPhone}
                    onChange={(e) => setRegPhone(e.target.value)}
                    required
                    className={cn(inputClass, 'pl-10 pr-4')}
                  />
                </div>

                {/* Password */}
                <div className="relative">
                  <Lock size={15} className="absolute left-3.5 top-1/2 -translate-y-1/2 text-ink-2/50 pointer-events-none" />
                  <input
                    type={showRegPass ? 'text' : 'password'}
                    placeholder="Contraseña"
                    value={regPassword}
                    onChange={(e) => setRegPassword(e.target.value)}
                    required
                    className={cn(inputClass, 'pl-10 pr-11')}
                  />
                  <button
                    type="button"
                    onClick={() => setShowRegPass((v) => !v)}
                    className="absolute right-3.5 top-1/2 -translate-y-1/2 text-ink-2/50 hover:text-ink-2 transition-colors"
                  >
                    {showRegPass ? <EyeOff size={15} /> : <Eye size={15} />}
                  </button>
                </div>

                {error && <p className="text-xs text-status-cancelled font-medium">{error}</p>}

                <button
                  type="submit"
                  disabled={isPending}
                  className="bg-primary text-white font-bold rounded-lg px-6 py-3 text-sm transition-all duration-normal ease-smooth hover:bg-primary-dark hover:shadow-glow active:scale-[0.98] disabled:opacity-60 disabled:cursor-not-allowed mt-1"
                >
                  {isPending ? 'Creando cuenta...' : 'Crear cuenta'}
                </button>
              </motion.form>
            )}
          </AnimatePresence>
        </div>
      </div>
    </div>
  );
}
