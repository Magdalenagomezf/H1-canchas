import { expect, test, type Page } from '@playwright/test';
import { registerCustomer } from './helpers';

// The email link cannot be followed here (the token only travels by email and
// the database keeps its hash), so these tests cover both screens up to the
// backend's answer, stubbing it where a working link is needed. Redeeming a
// real token is covered by the backend tests.

test('the login page links to the forgot password form, which always confirms', async ({ page }) => {
  const customer = await registerCustomer('E2E Forgot');

  await page.goto('/login');
  await page.getByRole('link', { name: '¿Olvidaste tu contraseña?' }).click();
  await expect(page).toHaveURL(/\/recuperar-contrasena$/);

  await page.getByLabel('Email').fill(customer.email);
  await page.getByRole('button', { name: 'Enviar link' }).click();

  await expect(page.getByRole('status')).toContainText('Si el email corresponde a una cuenta');
});

// The real backend says any made-up token is invalid, so the tests that need
// the form answer the on-load check themselves, as if the link still worked.
async function linkStillWorks(page: Page) {
  await page.route('**/auth/password-reset/check', (route) => route.fulfill({ status: 204 }));
}

test('a used or invalid link shows its message as soon as the page opens', async ({ page }) => {
  await page.goto('/restablecer-contrasena#token=not-a-real-token');
  // The token leaves the address bar as soon as the page reads it.
  await expect(page).toHaveURL(/\/restablecer-contrasena$/);

  // No form to fill in for nothing: the page asked the backend on load.
  await expect(page.getByRole('alert')).toContainText('no es válido o ya venció');
  await expect(page.getByLabel('Nueva contraseña')).toHaveCount(0);
  // A retry after a lost response also lands here, so the page points to the login too.
  await expect(page.getByRole('alert')).toContainText('Si ya cambiaste la contraseña con este link');
  await page.getByRole('link', { name: 'Pedir un link nuevo' }).click();
  await expect(page).toHaveURL(/\/recuperar-contrasena$/);
});

test('a link that stops working before saving still shows the invalid link message', async ({ page }) => {
  await linkStillWorks(page);
  // The check passes, but the real confirm rejects the made-up token, as when it expires meanwhile.
  await page.goto('/restablecer-contrasena#token=not-a-real-token');
  await page.getByLabel('Nueva contraseña').fill('nueva-clave');
  await page.getByLabel('Repetí la contraseña').fill('nueva-clave');
  await page.getByRole('button', { name: 'Guardar contraseña' }).click();

  await expect(page.getByRole('alert')).toContainText('no es válido o ya venció');
});

test('the reset form checks that both passwords match before sending', async ({ page }) => {
  await linkStillWorks(page);
  let sent = false;
  page.on('request', (req) => { if (req.url().includes('/auth/password-reset/confirm')) sent = true; });

  await page.goto('/restablecer-contrasena#token=some-token');
  await page.getByLabel('Nueva contraseña').fill('nueva-clave');
  await page.getByLabel('Repetí la contraseña').fill('otra-clave');
  await page.getByRole('button', { name: 'Guardar contraseña' }).click();

  await expect(page.getByRole('alert')).toHaveText('Las contraseñas no coinciden');
  expect(sent).toBe(false);
});

test('a successful reset shows the success message and a link to log in', async ({ page }) => {
  await linkStillWorks(page);
  let body: unknown;
  await page.route('**/auth/password-reset/confirm', (route) => {
    body = route.request().postDataJSON();
    return route.fulfill({ status: 204 });
  });

  await page.goto('/restablecer-contrasena#token=some-token');
  await page.getByLabel('Nueva contraseña').fill('nueva-clave');
  await page.getByLabel('Repetí la contraseña').fill('nueva-clave');
  await page.getByRole('button', { name: 'Guardar contraseña' }).click();

  await expect(page.getByRole('status')).toContainText('Listo, tu contraseña cambió');
  expect(body).toEqual({ token: 'some-token', password: 'nueva-clave' });
  await page.getByRole('link', { name: 'Iniciar sesión', exact: true }).click();
  await expect(page).toHaveURL(/\/login$/);
});

test('a reset page opened without a token shows the invalid link message', async ({ page }) => {
  await page.goto('/restablecer-contrasena');
  await expect(page.getByRole('alert')).toContainText('no es válido o ya venció');
});
