import { expect, test, type APIRequestContext } from '@playwright/test';
import {
  apiAs, argDay, cleanupSpace, createSpace, loginAs, loginAsAdmin, registerCustomer, uniquePhone, type TestSpace,
} from './helpers';

// API integration tests: they call the backend directly, without a browser.
// They run before the browser E2E (project "api" in playwright.config.ts), so a
// broken API fails fast and points at the backend. They cover what the UI
// cannot show: status codes, role checks and the JSON contract the frontend reads.

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

async function slotIdOf(api: APIRequestContext, s: TestSpace): Promise<number> {
  const slots: { id: number; label: string }[] = await (await api.get(`/spaces/${s.id}/slots?date=${tomorrow}`)).json();
  const slot = slots.find((x) => x.label === s.slotLabel);
  if (!slot) throw new Error(`slot ${s.slotLabel} not found in space ${s.id}`);
  return slot.id;
}

test('public endpoints return the space and slot fields the frontend reads', async () => {
  const anon = await apiAs();

  const list = await anon.get('/spaces');
  expect(list.status()).toBe(200);
  const spaces: { id: number }[] = await list.json();
  expect(spaces.map((s) => s.id)).toContain(space.id);

  const detail = await anon.get(`/spaces/${space.id}`);
  expect(detail.status()).toBe(200);
  expect(await detail.json()).toMatchObject({
    id: space.id,
    name: space.name,
    type: 'cancha_padel',
    price_per_slot: 10000,
  });

  const slots = await anon.get(`/spaces/${space.id}/slots?date=${tomorrow}`);
  expect(slots.status()).toBe(200);
  // A new space also gets the default timed slots, besides the one createSpace adds.
  const body: Record<string, unknown>[] = await slots.json();
  const untimed = body.find((s) => s.label === space.slotLabel);
  expect(untimed).toMatchObject({ id: expect.any(Number), available: true });
  // A slot without a time omits start_time (undefined, not null): that crashed the space page once.
  expect(untimed).not.toHaveProperty('start_time');
  // Timed slots carry a clock time the frontend parses with timeToMinutes ("...T10:00:00Z" or "10:00").
  const timed = body.find((s) => s.start_time !== undefined);
  expect(timed?.start_time).toMatch(/(^|T)\d{2}:\d{2}/);
  expect(timed?.end_time).toMatch(/(^|T)\d{2}:\d{2}/);

  expect((await anon.get('/spaces/999999999')).status()).toBe(404);
  await anon.dispose();
});

test('protected routes reject a missing token, an invalid token and the wrong role', async () => {
  const anon = await apiAs();
  expect((await anon.get('/me')).status()).toBe(401);
  expect((await anon.get('/admin/bookings')).status()).toBe(401);
  await anon.dispose();

  const forged = await apiAs('not-a-valid-token');
  expect((await forged.get('/me')).status()).toBe(401);
  await forged.dispose();

  const c = await registerCustomer('API Roles');
  const customer = await loginAs(c.phone, c.password);
  expect((await customer.get('/me')).status()).toBe(200);
  expect((await customer.get('/admin/bookings')).status()).toBe(403);
  expect((await customer.get('/admin/users')).status()).toBe(403);
  expect(
    (await customer.post('/spaces', { data: { name: 'x', type: 'cancha_padel', price_per_slot: 1 } })).status(),
  ).toBe(403);
  await customer.dispose();
});

test('login with a wrong password answers 401', async () => {
  const c = await registerCustomer('API Login');
  const anon = await apiAs();
  const res = await anon.post('/auth/login', { data: { phone: c.phone, password: 'wrong-password' } });
  expect(res.status()).toBe(401);
  await anon.dispose();
});

test('a booked slot becomes unavailable and a second booking gets 409', async () => {
  const slotId = await slotIdOf(admin, space);
  const first = await registerCustomer('API Primero');
  const second = await registerCustomer('API Segundo');
  const firstApi = await loginAs(first.phone, first.password);
  const secondApi = await loginAs(second.phone, second.password);
  const booking = { space_id: space.id, slot_id: slotId, booking_date: tomorrow };

  const created = await firstApi.post('/bookings', { data: booking });
  expect(created.status()).toBe(201);
  expect(await created.json()).toMatchObject({
    id: expect.any(Number),
    booking_date: tomorrow,
    status: 'pending',
    deposit_status: 'unpaid',
    space: { id: space.id },
  });

  const slots: { id: number; available: boolean }[] =
    await (await secondApi.get(`/spaces/${space.id}/slots?date=${tomorrow}`)).json();
  expect(slots.find((s) => s.id === slotId)?.available).toBe(false);

  const conflict = await secondApi.post('/bookings', { data: booking });
  expect(conflict.status()).toBe(409);

  await firstApi.dispose();
  await secondApi.dispose();
});

test("a customer cannot see another customer's booking, staff can", async () => {
  const slotId = await slotIdOf(admin, space);
  const owner = await registerCustomer('API Duenio');
  const other = await registerCustomer('API Otro');
  const ownerApi = await loginAs(owner.phone, owner.password);
  const otherApi = await loginAs(other.phone, other.password);

  const created = await ownerApi.post('/bookings', {
    data: { space_id: space.id, slot_id: slotId, booking_date: tomorrow },
  });
  expect(created.status()).toBe(201);
  const { id } = await created.json();

  expect((await ownerApi.get(`/bookings/${id}`)).status()).toBe(200);
  expect((await otherApi.get(`/bookings/${id}`)).status()).toBe(403);
  expect((await otherApi.patch(`/bookings/${id}/cancel`)).status()).toBe(403);
  expect((await admin.get(`/bookings/${id}`)).status()).toBe(200);

  await ownerApi.dispose();
  await otherApi.dispose();
});

test('staff manual booking is confirmed and validates its input', async () => {
  const slotId = await slotIdOf(admin, space);

  const invalid = await admin.post('/bookings/manual', {
    data: { space_id: space.id, slot_id: slotId, booking_date: tomorrow, customer_phone: uniquePhone() },
  });
  expect(invalid.status()).toBe(400);

  const created = await admin.post('/bookings/manual', {
    data: {
      space_id: space.id,
      slot_id: slotId,
      booking_date: tomorrow,
      customer_name: 'API Manual',
      customer_phone: uniquePhone(),
    },
  });
  expect(created.status()).toBe(201);
  expect(await created.json()).toMatchObject({ status: 'confirmed', customer: { name: 'API Manual' } });
});
