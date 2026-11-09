import { expect, test } from '@playwright/test';
import { registerCustomer } from './helpers';

test('a registered customer logs in with phone and password', async ({ page }) => {
  const customer = await registerCustomer('E2E Login');

  await page.goto('/login');
  await page.getByLabel('Teléfono').fill(customer.phone);
  await page.getByLabel('Contraseña', { exact: true }).fill(customer.password);
  await page.getByRole('button', { name: 'Ingresar' }).click();

  await expect(page).toHaveURL('/');
  await expect(page.getByRole('link', { name: 'Mis reservas' }).first()).toBeVisible();
  expect(await page.evaluate(() => sessionStorage.getItem('token'))).toBeTruthy();
});

test('a wrong password shows an error and keeps the user out', async ({ page }) => {
  const customer = await registerCustomer('E2E Wrong Password');

  await page.goto('/login');
  await page.getByLabel('Teléfono').fill(customer.phone);
  await page.getByLabel('Contraseña', { exact: true }).fill('not-the-password');
  await page.getByRole('button', { name: 'Ingresar' }).click();

  await expect(page.getByRole('alert')).toBeVisible();
  await expect(page).toHaveURL(/\/login$/);
  expect(await page.evaluate(() => sessionStorage.getItem('token'))).toBeNull();
});
