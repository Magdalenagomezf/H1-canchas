import { useEffect, useRef, useState } from 'react';
import { AnimatePresence, motion, useInView, useReducedMotion } from 'framer-motion';
import { cn } from '@/lib/utils';
import { FACTS, PLACE_IMAGES, PLACE_SLIDE_MS, PLACE_STATEMENT } from './content';
import { FadeIn } from './Reveal';
import { SectionLabel } from './SectionLabel';
import { CONTAINER, EASE, GRID, MONO, SECTION_Y, VIEWPORT } from './ui';

/**
 * Crossfading slideshow inside a frame that reveals bottom-up once.
 * Rotation only runs while the frame is on screen and never with reduced motion.
 */
function PlaceSlideshow({ className }: { className?: string }) {
  const reduce = useReducedMotion();
  const frameRef = useRef<HTMLDivElement>(null);
  // Observe the unclipped frame: a fully clipped target never reports as intersecting.
  const revealed = useInView(frameRef, VIEWPORT);
  const onScreen = useInView(frameRef);
  const [index, setIndex] = useState(0);

  // Warm the cache once the frame is near, so the first swaps don't flash the empty frame
  useEffect(() => {
    if (!onScreen) return;
    PLACE_IMAGES.forEach(({ src }) => {
      new Image().src = src;
    });
  }, [onScreen]);

  useEffect(() => {
    if (reduce || !onScreen) return;
    const id = window.setInterval(() => setIndex((i) => (i + 1) % PLACE_IMAGES.length), PLACE_SLIDE_MS);
    return () => window.clearInterval(id);
  }, [reduce, onScreen]);

  const image = PLACE_IMAGES[index];
  const hidden = reduce ? { opacity: 0 } : { clipPath: 'inset(100% 0 0 0)' };
  const shown = reduce ? { opacity: 1 } : { clipPath: 'inset(0% 0 0 0)' };

  return (
    <div ref={frameRef} className={cn('relative overflow-hidden bg-graphite', className)}>
      <motion.div
        className="absolute inset-0"
        initial={hidden}
        animate={revealed ? shown : hidden}
        transition={{ duration: 0.7, ease: EASE }}
      >
        <AnimatePresence initial={false}>
          <motion.img
            key={image.src}
            src={image.src}
            alt={image.alt}
            loading="lazy"
            decoding="async"
            className="absolute inset-0 h-full w-full object-cover"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.8, ease: 'easeOut' }}
          />
        </AnimatePresence>
      </motion.div>
    </div>
  );
}

/** 01/ The only light section: one statement, a photo slideshow and a row of facts. */
export function Place() {
  return (
    <section aria-label="El lugar" className={cn('bg-paper text-night', SECTION_Y)}>
      <div className={CONTAINER}>
        <SectionLabel number="01" label="El lugar" tone="light" />

        <div className={cn(GRID, 'mt-14 items-center gap-y-12 md:mt-24')}>
          <FadeIn className="col-span-12 lg:col-span-5">
            <p className="font-arch font-expanded text-[clamp(1.75rem,3.2vw,3rem)] font-extrabold uppercase leading-[0.95] tracking-[-0.02em] text-night">
              {PLACE_STATEMENT}
            </p>
          </FadeIn>
          <PlaceSlideshow className="col-span-12 aspect-[16/10] lg:col-span-7" />
        </div>

        <ul className="mt-20 grid grid-cols-2 gap-x-6 gap-y-10 border-t border-night/15 pt-10 md:mt-28 md:grid-cols-4">
          {FACTS.map((fact, i) => (
            <li key={fact.label}>
              <FadeIn delay={i * 0.08}>
                <span className="block font-arch font-expanded text-[clamp(2rem,4vw,3.5rem)] font-bold leading-none tracking-[-0.02em] text-night">
                  {fact.value}
                </span>
                <span className={cn(MONO, 'mt-3 block text-graphite')}>{fact.label}</span>
              </FadeIn>
            </li>
          ))}
        </ul>
      </div>
    </section>
  );
}
