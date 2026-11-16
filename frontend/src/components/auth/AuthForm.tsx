import type { ReactNode } from 'react';
import { ArrowRight, Eye, EyeOff } from 'lucide-react';
import { motion, useReducedMotion } from 'framer-motion';
import { buttonVariants } from '@/components/ui/button';
import { LandingNav } from '@/components/landing/LandingNav';
import { LightLine } from '@/components/landing/LightLine';
import { SectionLabel } from '@/components/landing/SectionLabel';
import { CONTAINER, EASE, GRID, MONO } from '@/components/landing/ui';
import { cn } from '@/lib/utils';

// Form pieces shared by the login, forgot password and reset password pages.

export const inputClass =
  'w-full border-0 border-b border-paper/25 bg-transparent py-3 text-lg text-paper placeholder:text-paper/30 outline-none transition-colors duration-300 focus:border-paper focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-paper';

export function Field({ id, label, children }: { id: string; label: string; children: ReactNode }) {
  return (
    <div className="flex flex-col gap-1">
      <label htmlFor={id} className={cn(MONO, 'text-concrete')}>
        {label}
      </label>
      {children}
    </div>
  );
}

export function PasswordInput({
  id,
  value,
  onChange,
  show,
  onToggle,
  autoComplete,
}: {
  id: string;
  value: string;
  onChange: (v: string) => void;
  show: boolean;
  onToggle: () => void;
  autoComplete?: string;
}) {
  return (
    <div className="relative">
      <input
        id={id}
        type={show ? 'text' : 'password'}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        autoComplete={autoComplete}
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

export function FormFooter({
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

/** Page frame of the access screens: title on the left, form on the right. */
export function AuthShell({ lines, text, children }: { lines: readonly string[]; text: string; children: ReactNode }) {
  const reduce = useReducedMotion();
  const y = reduce ? 0 : 12;

  return (
    <div className="landing">
      <div className="on-dark min-h-[100svh] bg-night text-paper">
        <LandingNav solid hideAccountLink />

        <main className="relative pt-16 md:pt-20">
          <div
            aria-hidden="true"
            className="pointer-events-none absolute inset-x-0 top-16 h-[160px] bg-gradient-to-b from-light/10 to-transparent md:top-20 md:h-[200px]"
          />
          <LightLine className="absolute inset-x-0 top-16 md:top-20" delay={0.3} />

          <div className={cn(CONTAINER, GRID, 'relative py-14 md:py-20 lg:min-h-[calc(100svh-5rem)] lg:items-center lg:py-24')}>
            <div className="col-span-12 lg:col-span-6">
              <SectionLabel number="H1" label="Acceso" />
              <h1 className="mt-10 font-arch font-expanded text-[clamp(2.25rem,min(8vw,12vh),6.5rem)] font-extrabold uppercase leading-[0.9] tracking-[-0.02em] text-paper md:mt-14">
                {lines.map((line, i) => (
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
                className="mt-6 max-w-md text-base leading-relaxed text-paper/85 md:text-lg"
                initial={{ opacity: 0, y }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.6, ease: EASE, delay: 0.35 }}
              >
                {text}
              </motion.p>
            </div>

            <motion.div
              className="col-span-12 mt-14 lg:col-span-5 lg:col-start-8 lg:mt-0"
              initial={{ opacity: 0, y }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.7, ease: EASE, delay: 0.3 }}
            >
              {children}
            </motion.div>
          </div>
        </main>
      </div>
    </div>
  );
}
