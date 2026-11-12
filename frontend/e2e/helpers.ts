import { request, type APIRequestContext, type APIResponse } from '@playwright/test';
import { toArgISO } from '../src/lib/dates';

export const API_URL = process.env.E2E_API_URL ?? 'http://localhost:8080';

export const TEST_PASSWORD = 'e2e-password-123';

/** A unique tag so runs never collide with each other or with leftovers in QA. */
export const uniqueTag = () => `${Date.now()}${Math.floor(Math.random() * 1000)}`;

/**
 * A random Argentine mobile number in E.164 (+5493515xxxxxx). The subscriber
 * part must not start with 0, or the backend rejects it as invalid.
 */
export const uniquePhone = () => `+5493515${String(Math.floor(Math.random() * 1_000_000)).padStart(6, '0')}`;

/** A day in Argentina as YYYY-MM-DD, offset from today. */
export const argDay = (offsetDays: number) => toArgISO(new Date(Date.now() + offsetDays * 86_400_000));

export function requireEnv(name: string): string {
  const value = process.env[name];
  if (!value) throw new Error(`${name} is required to run the E2E tests`);
  return value;
}

export async function apiAs(token?: string): Promise<APIRequestContext> {
  return request.newContext({
    baseURL: API_URL,
    extraHTTPHeaders: token ? { Authorization: `Bearer ${token}` } : {},
  });
}

async function ok<T>(res: APIResponse): Promise<T> {
  if (!res.ok()) throw new Error(`${res.url()} -> ${res.status()}: ${await res.text()}`);
  return (await res.json()) as T;
}

/** Logs in with phone and password and returns an API context that sends the user's token. */
export async function loginAs(phone: string, password: string): Promise<APIRequestContext> {
  const anon = await apiAs();
  const { token } = await ok<{ token: string }>(await anon.post('/auth/login', { data: { phone, password } }));
  await anon.dispose();
  return apiAs(token);
}

export async function loginAsAdmin(): Promise<APIRequestContext> {
  return loginAs(requireEnv('E2E_ADMIN_PHONE'), requireEnv('E2E_ADMIN_PASSWORD'));
}

export async function registerCustomer(name: string) {
  const anon = await apiAs();
  const phone = uniquePhone();
  await ok(
    await anon.post('/auth/register', {
      data: { name, phone, password: TEST_PASSWORD, email: `e2e-${uniqueTag()}@example.com` },
    }),
  );
  await anon.dispose();
  return { name, phone, password: TEST_PASSWORD };
}

export interface TestSpace {
  id: number;
  name: string;
  slotLabel: string;
}

/** Creates an active space with one slot (no fixed time, so it never counts as past). */
export async function createSpace(admin: APIRequestContext): Promise<TestSpace> {
  const tag = uniqueTag();
  const name = `E2E ${tag}`;
  const space = await ok<{ id: number }>(
    await admin.post('/spaces', {
      data: { name, type: 'cancha_padel', description: 'Creado por los tests E2E', price_per_slot: 10000 },
    }),
  );
  const slotLabel = `E2E turno ${tag}`;
  await ok(await admin.post(`/spaces/${space.id}/slots`, { data: { label: slotLabel } }));
  return { id: space.id, name, slotLabel };
}

interface AdminBooking {
  id: number;
  status: string;
  space: { id: number };
  customer: { name: string; phone: string | null };
}

export async function bookingsOn(admin: APIRequestContext, date: string, spaceId: number): Promise<AdminBooking[]> {
  const rows = await ok<AdminBooking[]>(await admin.get(`/admin/bookings?date=${date}`));
  return rows.filter((b) => b.space.id === spaceId);
}

/** Cancels the space's active bookings on the given dates and deactivates the space. */
export async function cleanupSpace(admin: APIRequestContext, space: TestSpace, dates: string[]) {
  for (const date of dates) {
    for (const b of await bookingsOn(admin, date, space.id)) {
      if (b.status === 'pending' || b.status === 'confirmed') {
        await admin.patch(`/bookings/${b.id}/cancel`);
      }
    }
  }
  await admin.delete(`/spaces/${space.id}`);
}
