import { useRef, useState, type ReactNode } from 'react';
import { motion, useInView, useReducedMotion, useScroll, useTransform } from 'framer-motion';
import { cn } from '@/lib/utils';
import { EASE, VIEWPORT } from './ui';

interface ImageRevealProps {
  src?: string;
  alt: string;
  /** Classes for the frame (size / aspect ratio). */
  className?: string;
  /** Hero image: no lazy loading, high fetch priority. */
  eager?: boolean;
  /** Subtle scroll parallax (disabled with reduced motion). */
  parallax?: boolean;
  /** Bottom-up clip-path curtain (default true). */
  reveal?: boolean;
  /** Scale 1.03 on hover of the closest `.group`. */
  hoverZoom?: boolean;
}

/**
 * Image inside a dark frame. If the file is missing the <img> unmounts
 * and the graphite block stays (no broken-image icon).
 */
export function ImageReveal({
  src,
  alt,
  className,
  eager = false,
  parallax = false,
  reveal = true,
  hoverZoom = false,
}: ImageRevealProps) {
  const reduce = useReducedMotion();
  const frameRef = useRef<HTMLDivElement>(null);
  const [failed, setFailed] = useState(false);
  const { scrollYProgress } = useScroll({ target: frameRef, offset: ['start end', 'end start'] });
  const y = useTransform(scrollYProgress, [0, 1], ['-6%', '6%']);
  const withParallax = parallax && !reduce;
  // Observe the unclipped frame: a fully clipped target never reports as intersecting.
  const inView = useInView(frameRef, VIEWPORT);

  const hidden = reduce ? { opacity: 0 } : { clipPath: 'inset(100% 0 0 0)' };
  const shown = reduce ? { opacity: 1 } : { clipPath: 'inset(0% 0 0 0)' };

  return (
    <div ref={frameRef} className={cn('relative overflow-hidden bg-graphite', className)}>
      <motion.div
        className="absolute inset-0"
        initial={reveal ? hidden : false}
        animate={reveal ? (inView ? shown : hidden) : undefined}
        transition={{ duration: 0.7, ease: EASE }}
      >
        <motion.div
          className={cn(
            'h-full w-full',
            hoverZoom && 'transition-transform duration-700 ease-arch group-hover:scale-[1.03]',
          )}
          style={withParallax ? { y, scale: 1.14 } : undefined}
        >
          {src && !failed && (
            <img
              src={src}
              alt={alt}
              loading={eager ? undefined : 'lazy'}
              fetchPriority={eager ? 'high' : undefined}
              decoding="async"
              onError={() => setFailed(true)}
              className="h-full w-full object-cover"
            />
          )}
        </motion.div>
      </motion.div>
    </div>
  );
}

interface LineTitleProps {
  lines: string[];
  as?: 'h1' | 'h2' | 'p' | 'div';
  className?: string;
  delay?: number;
}

/** Title revealed line by line (80ms stagger). */
export function LineTitle({ lines, as: Tag = 'h2', className, delay = 0 }: LineTitleProps) {
  const reduce = useReducedMotion();
  return (
    <Tag className={className}>
      {lines.map((line, i) => (
        <motion.span
          key={line}
          className="block"
          initial={{ opacity: 0, y: reduce ? 0 : 28 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={VIEWPORT}
          transition={{ duration: 0.7, ease: EASE, delay: delay + i * 0.08 }}
        >
          {line}
        </motion.span>
      ))}
    </Tag>
  );
}

/** Short fade + slight rise, once. */
export function FadeIn({
  children,
  className,
  delay = 0,
}: {
  children: ReactNode;
  className?: string;
  delay?: number;
}) {
  const reduce = useReducedMotion();
  return (
    <motion.div
      className={className}
      initial={{ opacity: 0, y: reduce ? 0 : 16 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={VIEWPORT}
      transition={{ duration: 0.6, ease: EASE, delay }}
    >
      {children}
    </motion.div>
  );
}
