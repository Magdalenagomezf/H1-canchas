import { useRef } from 'react';
import { motion, useInView, useReducedMotion } from 'framer-motion';
import { cn } from '@/lib/utils';
import { EASE, VIEWPORT } from './ui';

const DRAW = 1.2;

/**
 * Signature warm LED line: draws left to right once, then the glow
 * "powers on". The glow is the only shadow allowed on the landing.
 */
export function LightLine({ className, delay = 0.3 }: { className?: string; delay?: number }) {
  const reduce = useReducedMotion();
  const ref = useRef<HTMLDivElement>(null);
  // Observe the wrapper (never scaled) so the trigger does not depend on the drawn line.
  const inView = useInView(ref, VIEWPORT);
  const drawTime = reduce ? 0.4 : DRAW;

  return (
    <div ref={ref} aria-hidden="true" className={cn('relative h-[3px] w-full', className)}>
      <motion.div
        className="absolute inset-x-0 top-1/2 h-4 -translate-y-1/2 bg-light/45 blur-[10px]"
        initial={{ opacity: 0 }}
        animate={{ opacity: inView ? 1 : 0 }}
        transition={{ duration: reduce ? 0.4 : 1, ease: 'easeOut', delay: reduce ? delay : delay + DRAW * 0.8 }}
      />
      <motion.div
        className="absolute inset-0 origin-left bg-light shadow-[0_0_6px_1px_color-mix(in_srgb,var(--color-light)_60%,transparent)]"
        initial={reduce ? { opacity: 0 } : { scaleX: 0 }}
        animate={inView ? (reduce ? { opacity: 1 } : { scaleX: 1 }) : undefined}
        transition={{ duration: drawTime, ease: EASE, delay }}
      />
    </div>
  );
}
