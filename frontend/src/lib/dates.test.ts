import { describe, expect, it } from 'vitest';
import { argMinutesOf, formatDuration, shiftDate, timeToMinutes, toArgISO } from './dates';

describe('toArgISO', () => {
  it('uses the Argentina calendar day, not the UTC one', () => {
    // 02:30 UTC on the 10th is still 23:30 of the 9th in Argentina (UTC-3).
    expect(toArgISO(new Date('2026-11-10T02:30:00Z'))).toBe('2026-11-09');
    expect(toArgISO(new Date('2026-11-10T03:30:00Z'))).toBe('2026-11-10');
  });
});

describe('argMinutesOf', () => {
  it('returns minutes since midnight in Argentina', () => {
    expect(argMinutesOf(new Date('2026-11-09T21:15:00Z'))).toBe(18 * 60 + 15);
    expect(argMinutesOf(new Date('2026-11-10T03:00:00Z'))).toBe(0);
  });
});

describe('timeToMinutes', () => {
  it('parses a plain HH:MM:SS time', () => {
    expect(timeToMinutes('18:30:00')).toBe(18 * 60 + 30);
  });

  it('parses a TIME serialized as a full datetime, ignoring the Z', () => {
    expect(timeToMinutes('0000-01-01T18:30:00Z')).toBe(18 * 60 + 30);
  });
});

describe('formatDuration', () => {
  it.each([
    [30, '30 min'],
    [60, '1 hora'],
    [120, '2 horas'],
    [90, '1 h 30 min'],
  ])('%i minutes -> %s', (minutes, expected) => {
    expect(formatDuration(minutes)).toBe(expected);
  });
});

describe('shiftDate', () => {
  it.each([
    ['2026-11-09', 1, '2026-11-10'],
    ['2026-12-31', 1, '2027-01-01'],
    ['2026-03-01', -1, '2026-02-28'],
  ])('%s %+i day(s) -> %s', (iso, days, expected) => {
    expect(shiftDate(iso, days)).toBe(expected);
  });
});
