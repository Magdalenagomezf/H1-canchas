import { useState, type ReactNode } from 'react';
import { Link, useLocation, useNavigate, useSearchParams } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { ArrowRight } from 'lucide-react';
import { getMe, updateMe, type UpdateMePayload } from '../api/users';
import { useAuth } from '../hooks/useAuth';
import type { User } from '../types';
import { buttonVariants } from '@/components/ui/button';
import { LandingNav } from '@/components/landing/LandingNav';
import { LightLine } from '@/components/landing/LightLine';
import { SectionLabel } from '@/components/landing/SectionLabel';
import { CONTAINER, GRID, MONO } from '@/components/landing/ui';
import { cn } from '@/lib/utils';

const inputClass =
  'w-full border-0 border-b border-paper/25 bg-transparent py-3 text-lg text-paper placeholder:text-paper/30 outline-none transition-colors duration-300 focus:border-paper focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-paper';

type FieldName = 'name' | 'email' | 'phone';

export interface ProfileLocationState {
  /** Message shown above the form, e.g. why the user was sent here. */
  notice?: string;
  /** Booking selection to restore when returning to `next`. */
  pendingBooking?: { date: string; slotId: number };
}

/** Only same-site paths are accepted as `next`, so the page can't redirect elsewhere. */
function safeNext(next: string | null): string | null {
  return next && next.startsWith('/') && !next.startsWith('//') ? next : null;
}

/**
 * Profile errors (400 validation, 409 duplicates) all start with the field
 * they refer to, e.g. "el teléfono ya está registrado".
 */
function fieldOfError(message: string): FieldName | null {
  const m = message.toLowerCase();
  if (m.startsWith('el nombre')) return 'name';
  if (m.startsWith('el email')) return 'email';
  if (m.startsWith('el teléfono')) return 'phone';
  return null;
}

function Field({ id, label, error, children }: { id: string; label: string; error?: string; children: ReactNode }) {
  return (
    <div className="flex flex-col gap-1">
      <label htmlFor={id} className={cn(MONO, 'text-concrete')}>
        {label}
      </label>
      {children}
      {error && (
        <p id={`${id}-error`} role="alert" className={cn(MONO, 'mt-1 normal-case tracking-normal text-red-300')}>
          {error}
        </p>
      )}
    </div>
  );
}

function ProfileForm({ me }: { me: User }) {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const { updateUser } = useAuth();
  const queryClient = useQueryClient();

  const state = location.state as ProfileLocationState | null;
  const next = safeNext(searchParams.get('next'));

  const [name, setName] = useState(me.name);
  const [email, setEmail] = useState(me.email ?? '');
  const [phone, setPhone] = useState(me.phone ?? '');
  const [saved, setSaved] = useState(false);

  // Only the fields that changed are sent: an absent field is left untouched by the backend.
  const changes: UpdateMePayload = {};
  if (name.trim() !== me.name) changes.name = name;
  if (email.trim() !== (me.email ?? '')) changes.email = email;
  if (phone.trim() !== (me.phone ?? '')) changes.phone = phone;
  const hasChanges = Object.keys(changes).length > 0;

  const mutation = useMutation({
    mutationFn: () => updateMe(changes),
    onSuccess: (user) => {
      updateUser(user);
      queryClient.setQueryData(['me'], user);
      if (next) {
        navigate(next, { replace: true, state: state?.pendingBooking ? { pendingBooking: state.pendingBooking } : undefined });
        return;
      }
      setName(user.name);
      setEmail(user.email ?? '');
      setPhone(user.phone ?? '');
      setSaved(true);
    },
  });

  const errorMessage = mutation.error?.message;
  const errorField = errorMessage ? fieldOfError(errorMessage) : null;
  const fieldError = (field: FieldName) => (errorField === field ? errorMessage : undefined);

  const edit = (setter: (v: string) => void) => (value: string) => {
    setter(value);
    setSaved(false);
    if (mutation.isError) mutation.reset();
  };

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (hasChanges) mutation.mutate();
      }}
      className="flex flex-col gap-8"
    >
      {state?.notice && (
        <p role="status" className="border-l-2 border-light pl-4 text-base text-paper">
          {state.notice}
        </p>
      )}

      <Field id="profile-name" label="Nombre completo" error={fieldError('name')}>
        <input
          id="profile-name"
          type="text"
          autoComplete="name"
          value={name}
          onChange={(e) => edit(setName)(e.target.value)}
          aria-invalid={errorField === 'name'}
          aria-describedby={errorField === 'name' ? 'profile-name-error' : undefined}
          required
          maxLength={100}
          className={inputClass}
        />
      </Field>
      <Field id="profile-email" label="Email" error={fieldError('email')}>
        <input
          id="profile-email"
          type="email"
          autoComplete="email"
          value={email}
          onChange={(e) => edit(setEmail)(e.target.value)}
          aria-invalid={errorField === 'email'}
          aria-describedby={errorField === 'email' ? 'profile-email-error' : undefined}
          className={inputClass}
        />
      </Field>
      <Field id="profile-phone" label="Teléfono" error={fieldError('phone')}>
        <input
          id="profile-phone"
          type="tel"
          autoComplete="tel"
          value={phone}
          onChange={(e) => edit(setPhone)(e.target.value)}
          aria-invalid={errorField === 'phone'}
          aria-describedby={errorField === 'phone' ? 'profile-phone-error' : undefined}
          placeholder="Ej. 383 412 3456"
          className={inputClass}
        />
      </Field>

      {errorMessage && !errorField && (
        <p role="alert" className={cn(MONO, 'normal-case tracking-normal text-red-300')}>
          {errorMessage}
        </p>
      )}
      {saved && (
        <p role="status" className={cn(MONO, 'normal-case tracking-normal text-concrete')}>
          Cambios guardados
        </p>
      )}

      <button
        type="submit"
        disabled={!hasChanges || mutation.isPending}
        className={cn(
          buttonVariants({ variant: 'court' }),
          'mt-2 h-14 w-full justify-center gap-3 text-base font-semibold disabled:cursor-not-allowed disabled:opacity-60',
        )}
      >
        {mutation.isPending ? 'Guardando...' : 'Guardar cambios'}
        <ArrowRight size={18} />
      </button>
    </form>
  );
}

export default function ProfilePage() {
  const { data: me, isLoading, error } = useQuery({ queryKey: ['me'], queryFn: getMe });

  return (
    <div className="landing">
      <div className="on-dark min-h-[100svh] bg-night text-paper">
        <LandingNav solid />

        <main className="relative pt-16 md:pt-20">
          <LightLine className="absolute inset-x-0 top-16 md:top-20" delay={0.3} />

          <div className={cn(CONTAINER, GRID, 'relative py-14 md:py-20 lg:items-start lg:py-24')}>
            <div className="col-span-12 lg:col-span-6">
              <SectionLabel number="H1" label="Cuenta" />
              <h1 className="mt-10 font-arch font-expanded text-[clamp(2.25rem,min(8vw,12vh),6.5rem)] font-extrabold uppercase leading-[0.9] tracking-[-0.02em] text-paper md:mt-14">
                <span className="block">MI</span>
                <span className="block">PERFIL</span>
              </h1>
              <p className="mt-6 max-w-md text-base leading-relaxed text-paper/85 md:text-lg">
                Tus datos de contacto. Usamos el teléfono para avisarte sobre tus reservas.
              </p>
            </div>

            <div className="col-span-12 mt-14 lg:col-span-5 lg:col-start-8 lg:mt-0">
              {isLoading && (
                <p aria-busy="true" className={cn(MONO, 'text-concrete')}>
                  Cargando...
                </p>
              )}
              {error && (
                <p role="alert" className={cn(MONO, 'normal-case tracking-normal text-red-300')}>
                  {error.message}
                </p>
              )}
              {me && <ProfileForm me={me} />}

              <Link to="/" className={cn(MONO, 'mt-10 inline-block text-concrete transition-colors hover:text-paper')}>
                ← Volver al inicio
              </Link>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}
