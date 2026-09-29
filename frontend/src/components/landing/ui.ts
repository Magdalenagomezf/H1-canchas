/** Shared motion + layout constants for the landing. */
export const EASE = [0.22, 1, 0.36, 1] as const;
export const VIEWPORT = { once: true, amount: 0.2 } as const;

/** Page gutter + max width. */
export const CONTAINER = 'mx-auto w-full max-w-[1440px] px-6 md:px-10 lg:px-16';
/** 12-column grid used by every section. */
export const GRID = 'grid grid-cols-12 gap-x-6';
/** Vertical rhythm of a section. */
export const SECTION_Y = 'py-20 md:py-32 lg:py-40';
/** Mono data style: small, uppercase, wide tracking. */
export const MONO = 'font-jb text-xs uppercase tracking-[0.12em] leading-relaxed';
