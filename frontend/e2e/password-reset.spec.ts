import { expect, test } from '@playwright/test';
import { registerCustomer } from './helpers';

// The email link cannot be followed here (the token only travels by email and
// the database keeps its hash), so these tests cover both screens up to the
// backend's answer. Redeeming a real token is covered by the backend tests.

test('the login page links to the forgot password form, which always confirms', async ({ page }) => {
  const customer = await registerCustomer('E2E Forgot');

  await page.goto('/login');
  await page.getByRole('link', { name: '¿Olvidaste tu contraseña?' }).click();
  await expect(page).toHaveURL(/\/recuperar-contrasena$/);

  await page.getByLabel('Email').fill(customer.email);
  await page.getByRole('button', { name: 'Enviar link' }).click();

  await expect(page.getByRole('status')).toContainText('Si el email corresponde a una cuenta');
});

test('a reset link with an invalid token offers to ask for a new one', async ({ page }) => {
  await page.goto('/restablecer-contrasena#token=not-a-real-token');
  // The token leaves the address bar as soon as the page reads it.
  await expect(page).toHaveURL(/\/restablecer-contrasena$/);

  await page.getByLabel('Nueva contraseña').fill('nueva-clave');
  await page.getByLabel('Repetí la contraseña').fill('nueva-clave');
  await page.getByRole('button', { name: 'Guardar contraseña' }).click();

  await expect(page.getByRole('alert')).toContainText('no es válido o ya venció');
  // A retry after a lost response also lands here, so the page points to the login too.
  await expect(page.getByRole('alert')).toContainText('Si ya cambiaste la contraseña con este link');
  await page.getByRole('link', { name: 'Pedir un link nuevo' }).click();
  await expect(page).toHaveURL(/\/recuperar-contrasena$/);
});

test('the reset form checks that both passwords match before sending', async ({ page }) => {
  let sent = false;
  page.on('request', (req) => { if (req.url().includes('/auth/password-reset/confirm')) sent = true; });

  await page.goto('/restablecer-contrasena#token=some-token');
  await page.getByLabel('Nueva contraseña').fill('nueva-clave');
  await page.getByLabel('Repetí la contraseña').fill('otra-clave');
  await page.getByRole('button', { name: 'Guardar contraseña' }).click();

  await expect(page.getByRole('alert')).toHaveText('Las contraseñas no coinciden');
  expect(sent).toBe(false);
});

test('a reset page opened without a token shows the invalid link message', async ({ page }) => {
  await page.goto('/restablecer-contrasena');
  await expect(page.getByRole('alert')).toContainText('no es válido o ya venció');
});
