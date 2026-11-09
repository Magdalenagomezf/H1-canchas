// Dates and times are always shown in Argentina time, wherever the browser is.
export const TZ = 'America/Argentina/Buenos_Aires';

/** Calendar day in Argentina as YYYY-MM-DD. */
export function toArgISO(d: Date): string {
  return d.toLocaleDateString('en-CA', { timeZone: TZ });
}

/** Minutes since midnight in Argentina for the given instant (default: now). */
export function argMinutesOf(d: Date = new Date()): number {
  const [h, m] = d.toLocaleTimeString('en-GB', { timeZone: TZ, hour12: false }).split(':');
  return Number(h) * 60 + Number(m);
}

/**
 * Minutes since midnight of a slot time. The API sends "HH:MM:SS" or, when a
 * TIME without zone is serialized as a full datetime, "0000-01-01THH:MM:SSZ".
 * Either way the time is Argentina local time; the "Z" is not a real UTC offset.
 */
export function timeToMinutes(time: string): number {
  const clock = time.includes('T') ? time.split('T')[1] : time;
  const [h, m] = clock.split(':');
  return Number(h) * 60 + Number(m);
}

/**
 * Whether a slot can no longer be booked: it is today and already started.
 * Slots without a start time never count as past. The API omits the field
 * for them, so it can arrive as undefined, not only null.
 */
export function isSlotPast(
  startTime: string | null | undefined,
  date: string,
  today: string,
  nowMinutes: number,
): boolean {
  if (date !== today || startTime == null) return false;
  return timeToMinutes(startTime) <= nowMinutes;
}

export function formatDuration(minutes: number): string {
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  if (h === 0) return `${m} min`;
  if (m === 0) return h === 1 ? '1 hora' : `${h} horas`;
  return `${h} h ${m} min`;
}

/** Moves a YYYY-MM-DD date by the given number of days. */
export function shiftDate(iso: string, days: number): string {
  // Noon keeps the result on the right day across DST and timezone offsets.
  const d = new Date(iso + 'T12:00:00');
  d.setDate(d.getDate() + days);
  return toArgISO(d);
}
