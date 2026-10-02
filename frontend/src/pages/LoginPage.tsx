import { useState, type ReactNode } from 'react';
import { Link, useNavigate, useLocation, type Location } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import { ArrowRight, Eye, EyeOff } from 'lucide-react';
import { motion, AnimatePresence, useReducedMotion } from 'framer-motion';
import { login, register } from '../api/auth';
import { useAuth } from '../hooks/useAuth';
import { buttonVariants } from '@/components/ui/button';
import { LandingNav } from '@/components/landing/LandingNav';
import { LightLine } from '@/components/landing/LightLine';
import { SectionLabel } from '@/components/landing/SectionLabel';
import { CONTAINER, EASE, GRID, MONO } from '@/components/landing/ui';
import { cn } from '@/lib/utils';

const inputClass =
  'w-full border-0 border-b border-paper/25 bg-transparent py-3 text-lg text-paper placeholder:text-paper/30 outline-none transition-colors duration-300 focus:border-paper focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-paper';

const TITLES = [
  { lines: ['BIENVENIDO'], text: 'Ingresá para gestionar tus reservas' },
  { lines: ['CREÁ TU', 'CUENTA'], text: 'Es gratis y lleva menos de un minuto' },
] as const;

function Field({ id, label, children }: { id: string; label: string; children: ReactNode }) {
  return (
    <div className="flex flex-col gap-1">
      <label htmlFor={id} className={cn(MONO, 'text-concrete')}>
        {label}
      </label>
      {children}
    </div>
  );
}

function PasswordInput({
  id,
  value,
  onChange,
  show,
  onToggle,
}: {
  id: string;
  value: string;
  onChange: (v: string) => void;
  show: boolean;
  onToggle: () => void;
}) {
  return (
    <div className="relative">
      <input
        id={id}
        type={show ? 'text' : 'password'}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        required
        className={cn(inputClass, 'pr-11')}
      />
      <button
        type="button"
        onClick={onToggle}
        aria-label={show ? 'Ocultar contraseña' : 'Mostrar contraseña'}
        className="absolute right-0 top-1/2 flex size-10 -translate-y-1/2 items-center justify-center text-concrete transition-colors hover:text-paper"
      >
        {show ? <EyeOff size={18} /> : <Eye size={18} />}
      </button>
    </div>
  );
}

function FormFooter({
  error,
  isPending,
  idle,
  pending,
}: {
  error?: string;
  isPending: boolean;
  idle: string;
  pending: string;
}) {
  return (
    <>
      {error && (
        <p role="alert" className={cn(MONO, 'normal-case tracking-normal text-red-300')}>
          {error}
        </p>
      )}
      <button
        type="submit"
        disabled={isPending}
        className={cn(
          buttonVariants({ variant: 'court' }),
          'mt-2 h-14 w-full justify-center gap-3 text-base font-semibold disabled:cursor-not-allowed disabled:opacity-60',
        )}
      >
        {isPending ? pending : idle}
        <ArrowRight size={18} className="transition-transform duration-300 group-hover:translate-x-1" />
      </button>
    </>
  );
}

export default function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { setAuth } = useAuth();
  const reduce = useReducedMotion();

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
  const [regEmail, setRegEmail] = useState('');
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
    mutationFn: () => register(regName, regPhone, regPassword, regEmail),
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

  const title = TITLES[tab];
  const y = reduce ? 0 : 12;
  const swap = { duration: 0.25, ease: EASE };

  return (
    <div className="landing">
      <div className="on-dark min-h-[100svh] bg-night text-paper">
        <LandingNav solid hideAccountLink />

        <main className="relative pt-16 md:pt-20">
          {/* Slab LED with its soft warm wash, same as the hero */}
          <div
            aria-hidden="true"
            className="pointer-events-none absolute inset-x-0 top-16 h-[160px] bg-gradient-to-b from-light/10 to-transparent md:top-20 md:h-[200px]"
          />
          <LightLine className="absolute inset-x-0 top-16 md:top-20" delay={0.3} />

          <div className={cn(CONTAINER, GRID, 'relative py-14 md:py-20 lg:min-h-[calc(100svh-5rem)] lg:items-center lg:py-24')}>
            <div className="col-span-12 lg:col-span-6">
              <SectionLabel number="H1" label="Acceso" />

              <h1
                key={tab}
                className="mt-10 font-arch font-expanded text-[clamp(2.25rem,min(8vw,12vh),6.5rem)] font-extrabold uppercase leading-[0.9] tracking-[-0.02em] text-paper md:mt-14"
              >
                {title.lines.map((line, i) => (
                  <motion.span
                    key={line}
                    className="block"
                    initial={{ opacity: 0, y: reduce ? 0 : 28 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.7, ease: EASE, delay: 0.15 + i * 0.08 }}
                  >
                    {line}
                  </motion.span>
                ))}
              </h1>
              <motion.p
                key={`sub-${tab}`}
                className="mt-6 max-w-md text-base leading-relaxed text-paper/85 md:text-lg"
                initial={{ opacity: 0, y }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.6, ease: EASE, delay: 0.35 }}
              >
                {title.text}
              </motion.p>
            </div>

            <motion.div
              className="col-span-12 mt-14 lg:col-span-5 lg:col-start-8 lg:mt-0"
              initial={{ opacity: 0, y }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.7, ease: EASE, delay: 0.3 }}
            >
              <div role="tablist" aria-label="Acceso" className="mb-10 flex gap-8">
                {(['Iniciar sesión', 'Registrarme'] as const).map((label, i) => (
                  <button
                    key={label}
                    type="button"
                    role="tab"
                    id={`auth-tab-${i}`}
                    aria-selected={tab === i}
                    aria-controls="auth-panel"
                    onClick={() => handleTabChange(i as 0 | 1)}
                    className={cn(
                      MONO,
                      'relative pb-3 transition-colors duration-300',
                      tab === i ? 'text-paper' : 'text-concrete hover:text-paper',
                    )}
                  >
                    {label}
                    {tab === i && (
                      <motion.span
                        layoutId="auth-tab-led"
                        aria-hidden="true"
                        className="absolute inset-x-0 bottom-0 h-[2px] bg-light"
                        transition={{ duration: reduce ? 0 : 0.35, ease: EASE }}
                      />
                    )}
                  </button>
                ))}
              </div>

              <div id="auth-panel" role="tabpanel" aria-labelledby={`auth-tab-${tab}`}>
                <AnimatePresence mode="wait">
                  {tab === 0 ? (
                    <motion.form
                      key="login"
                      initial={{ opacity: 0, y }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -y }}
                      transition={swap}
                      onSubmit={(e) => { e.preventDefault(); loginMutation.mutate(); }}
                      className="flex flex-col gap-8"
                    >
                      <Field id="login-phone" label="Teléfono">
                        <input
                          id="login-phone"
                          type="tel"
                          autoComplete="tel"
                          value={loginPhone}
                          onChange={(e) => setLoginPhone(e.target.value)}
                          required
                          className={inputClass}
                        />
                      </Field>
                      <Field id="login-password" label="Contraseña">
                        <PasswordInput
                          id="login-password"
                          value={loginPassword}
                          onChange={setLoginPassword}
                          show={showLoginPass}
                          onToggle={() => setShowLoginPass((v) => !v)}
                        />
                      </Field>
                      <FormFooter error={error} isPending={isPending} idle="Ingresar" pending="Ingresando..." />
                    </motion.form>
                  ) : (
                    <motion.form
                      key="register"
                      initial={{ opacity: 0, y }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -y }}
                      transition={swap}
                      onSubmit={(e) => { e.preventDefault(); registerMutation.mutate(); }}
                      className="flex flex-col gap-8"
                    >
                      <Field id="reg-name" label="Nombre completo">
                        <input
                          id="reg-name"
                          type="text"
                          autoComplete="name"
                          value={regName}
                          onChange={(e) => setRegName(e.target.value)}
                          required
                          className={inputClass}
                        />
                      </Field>
                      <Field id="reg-phone" label="Teléfono">
                        <input
                          id="reg-phone"
                          type="tel"
                          autoComplete="tel"
                          value={regPhone}
                          onChange={(e) => setRegPhone(e.target.value)}
                          required
                          className={inputClass}
                        />
                      </Field>
                      <Field id="reg-email" label="Email">
                        <input
                          id="reg-email"
                          type="email"
                          autoComplete="email"
                          value={regEmail}
                          onChange={(e) => setRegEmail(e.target.value)}
                          required
                          className={inputClass}
                        />
                      </Field>
                      <Field id="reg-password" label="Contraseña">
                        <PasswordInput
                          id="reg-password"
                          value={regPassword}
                          onChange={setRegPassword}
                          show={showRegPass}
                          onToggle={() => setShowRegPass((v) => !v)}
                        />
                      </Field>
                      <FormFooter error={error} isPending={isPending} idle="Crear cuenta" pending="Creando cuenta..." />
                    </motion.form>
                  )}
                </AnimatePresence>
              </div>

              <Link
                to="/"
                className={cn(MONO, 'mt-10 inline-block text-concrete transition-colors hover:text-paper')}
              >
                ← Volver al inicio
              </Link>
            </motion.div>
          </div>
        </main>
      </div>
    </div>
  );
}
