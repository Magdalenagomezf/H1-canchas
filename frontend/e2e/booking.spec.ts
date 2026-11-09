import { expect, test, type APIRequestContext } from '@playwright/test';
import {
  argDay, bookingsOn, cleanupSpace, createSpace, loginAsAdmin, TEST_PASSWORD, uniquePhone, uniqueTag, type TestSpace,
} from './helpers';

const FAKE_CHECKOUT = 'https://sandbox.mercadopago.test/checkout/e2e';

let admin: APIRequestContext;
let space: TestSpace;
const tomorrow = argDay(1);

test.beforeEach(async () => {
  admin = await loginAsAdmin();
  space = await createSpace(admin);
});

test.afterEach(async () => {
  await cleanupSpace(admin, space, [tomorrow]);
  await admin.dispose();
});

test('a new customer registers and books a slot up to the Mercado Pago checkout', async ({ page }) => {
  // Creating the preference calls the real Mercado Pago API, which CI has no
  // credentials for: answer it here and check that the app sends the user there.
  await page.route('**/bookings/*/payments', (route) =>
    route.fulfill({
      json: { preference_id: 'e2e', init_point: FAKE_CHECKOUT, sandbox_init_point: FAKE_CHECKOUT },
    }),
  );
  const name = `E2E Cliente ${uniqueTag()}`;

  await page.goto('/login');
  await page.getByRole('tab', { name: 'Registrarme' }).click();
  await page.getByLabel('Nombre completo').fill(name);
  await page.getByLabel('Teléfono').fill(uniquePhone());
  await page.getByLabel('Email').fill(`e2e-${uniqueTag()}@example.com`);
  await page.getByLabel('Contraseña', { exact: true }).fill(TEST_PASSWORD);
  await page.getByRole('button', { name: 'Crear cuenta' }).click();
  await expect(page).toHaveURL('/');

  await page.goto(`/canchas/${space.id}`);
  // Second day in the strip = tomorrow, so the slot is never in the past.
  await page.getByRole('group', { name: 'Seleccioná el día' }).getByRole('button').nth(1).click();
  await page.getByRole('group', { name: 'Turnos' }).getByRole('button', { name: space.slotLabel }).click();
  await page.getByRole('button', { name: 'Reservar turno' }).filter({ visible: true }).click();

  await expect(page.getByRole('link', { name: 'Pagar con Mercado Pago' })).toHaveAttribute('href', FAKE_CHECKOUT);

  const bookings = await bookingsOn(admin, tomorrow, space.id);
  expect(bookings).toHaveLength(1);
  expect(bookings[0]).toMatchObject({ status: 'pending', customer: { name } });
});
