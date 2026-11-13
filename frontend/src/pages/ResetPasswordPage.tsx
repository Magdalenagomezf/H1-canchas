import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import { confirmPasswordReset } from '../api/auth';
import { ApiError } from '../api/client';
import { AuthShell, Field, FormFooter, PasswordInput } from '@/components/auth/AuthForm';
import { buttonVariants } from '@/components/ui/button';
import { MONO } from '@/components/landing/ui';
import { newPasswordError, readResetToken } from '@/lib/passwordReset';
import { cn } from '@/lib/utils';

const linkButton = cn(buttonVariants({ variant: 'court' }), 'h-14 w-full justify-center text-base font-semibold');


function InvalidLink() {
  return (
    <div role="alert" className="flex flex-col gap-6">
      <p className="text-lg leading-relaxed text-paper">
        El link para restablecer la contraseña no es válido o ya venció. Pedí uno nuevo.
      </p>
      <Link to="/recuperar-contrasena" className={linkButton}>
        Pedir un link nuevo
      </Link>
    </div>
  );
}

export default function ResetPasswordPage() {
  const [token] = useState(() => readResetToken(window.location.hash));

  // Once read, drop the fragment from the address bar so the token does not
  // stay visible or get copied along with the URL. Done in an effect, not in
  // the initializer above, because StrictMode runs initializers twice.
  useEffect(() => {
    if (window.location.hash) {
      window.history.replaceState(window.history.state, '', window.location.pathname + window.location.search);
    }
  }, []);
  const [password, setPassword] = useState('');
  const [repeat, setRepeat] = useState('');
  const [show, setShow] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: () => confirmPasswordReset(token ?? '', password),
  });

  const linkInvalid = !token || (mutation.error instanceof ApiError && mutation.error.code === 'invalid_reset_token');

  const handleSubmit = () => {
    const err = newPasswordError(password, repeat);
    setFormError(err);
    if (!err) mutation.mutate();
  };

  let content;
  if (mutation.isSuccess) {
    content = (
      <div role="status" className="flex flex-col gap-6">
        <p className="text-lg leading-relaxed text-paper">Listo, tu contraseña cambió. Ya podés iniciar sesión con la nueva.</p>
        <Link to="/login" className={linkButton}>
          Iniciar sesión
        </Link>
      </div>
    );
  } else if (linkInvalid) {
    content = <InvalidLink />;
  } else {
    content = (
      <form onSubmit={(e) => { e.preventDefault(); handleSubmit(); }} className="flex flex-col gap-8">
        <Field id="reset-password" label="Nueva contraseña">
          <PasswordInput
            id="reset-password"
            value={password}
            onChange={setPassword}
            show={show}
            onToggle={() => setShow((v) => !v)}
            autoComplete="new-password"
          />
        </Field>
        <Field id="reset-repeat" label="Repetí la contraseña">
          <PasswordInput
            id="reset-repeat"
            value={repeat}
            onChange={setRepeat}
            show={show}
            onToggle={() => setShow((v) => !v)}
            autoComplete="new-password"
          />
        </Field>
        <FormFooter
          error={formError ?? mutation.error?.message}
          isPending={mutation.isPending}
          idle="Guardar contraseña"
          pending="Guardando..."
        />
      </form>
    );
  }

  return (
    <AuthShell lines={['NUEVA', 'CONTRASEÑA']} text="Elegí una contraseña nueva para tu cuenta">
      {content}
      <Link to="/login" className={cn(MONO, 'mt-10 inline-block text-concrete transition-colors hover:text-paper')}>
        ← Volver a iniciar sesión
      </Link>
    </AuthShell>
  );
}
