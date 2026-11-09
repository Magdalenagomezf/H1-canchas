import { expect, test, type APIRequestContext } from '@playwright/test';
import {
  argDay, bookingsOn, cleanupSpace, createSpace, loginAsAdmin, requireEnv, uniqueTag, type TestSpace,
} from './helpers';

let admin: APIRequestContext;
let space: TestSpace;
const today = argDay(0);

test.beforeEach(async () => {
  admin = await loginAsAdmin();
  space = await createSpace(admin);
});

test.afterEach(async () => {
  await cleanupSpace(admin, space, [today]);
  await admin.dispose();
});

test('staff creates a manual booking from the panel', async ({ page }) => {
  const customerName = `E2E Manual ${uniqueTag()}`;

  await page.goto('/login');
  await page.getByLabel('Teléfono').fill(requireEnv('E2E_ADMIN_PHONE'));
  await page.getByLabel('Contraseña', { exact: true }).fill(requireEnv('E2E_ADMIN_PASSWORD'));
  await page.getByRole('button', { name: 'Ingresar' }).click();
  await expect(page).toHaveURL('/');

  await page.goto('/panel');
  await page.getByRole('button', { name: 'Nueva reserva' }).click();
  const dialog = page.getByRole('dialog');
  await dialog.getByLabel('Espacio').selectOption({ label: space.name });
  await dialog.getByRole('button', { name: space.slotLabel }).click();
  await dialog.getByLabel('Nombre completo').fill(customerName);
  await dialog.getByLabel('Teléfono').fill('3515550000');
  await dialog.getByRole('button', { name: 'Confirmar reserva' }).click();

  await expect(dialog).toBeHidden();
  await expect(page.getByText(customerName)).toBeVisible();

  const bookings = await bookingsOn(admin, today, space.id);
  expect(bookings).toHaveLength(1);
  expect(bookings[0]).toMatchObject({ status: 'confirmed', customer: { name: customerName } });
});
