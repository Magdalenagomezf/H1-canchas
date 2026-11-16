import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import { requestPasswordReset } from '../api/auth';
import { AuthShell, Field, FormFooter, inputClass } from '@/components/auth/AuthForm';
import { MONO } from '@/components/landing/ui';
import { cn } from '@/lib/utils';

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('');
  const mutation = useMutation({ mutationFn: () => requestPasswordReset(email) });

  return (
    <AuthShell lines={['RECUPERÁ TU', 'CONTRASEÑA']} text="Te mandamos un link a tu email para que elijas una nueva">
      {mutation.isSuccess ? (
        // Same text whether the email has an account or not: it must never
        // reveal which emails are registered. How long the link lasts is said
        // in the email itself, next to the button.
        <div role="status" className="flex flex-col gap-4">
          <p className="text-2xl font-semibold leading-snug text-paper">¡Listo! Revisá tu email</p>
          <p className="text-lg leading-relaxed text-paper/85">
            Si está registrado, en unos minutos te llega el link.
          </p>
        </div>
      ) : (
        <form
          onSubmit={(e) => { e.preventDefault(); mutation.mutate(); }}
          className="flex flex-col gap-8"
        >
          <Field id="forgot-email" label="Email">
            <input
              id="forgot-email"
              type="email"
              autoComplete="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              className={inputClass}
            />
          </Field>
          <FormFooter
            error={mutation.error?.message}
            isPending={mutation.isPending}
            idle="Enviar link"
            pending="Enviando..."
          />
        </form>
      )}

      <Link to="/login" className={cn(MONO, 'mt-10 inline-block text-concrete transition-colors hover:text-paper')}>
        ← Volver a iniciar sesión
      </Link>
    </AuthShell>
  );
}
