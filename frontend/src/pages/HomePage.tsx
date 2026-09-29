import { Complex } from '@/components/landing/Complex';
import { FinalCta } from '@/components/landing/FinalCta';
import { Hero } from '@/components/landing/Hero';
import { HowToBook } from '@/components/landing/HowToBook';
import { LandingFooter } from '@/components/landing/LandingFooter';
import { LandingNav } from '@/components/landing/LandingNav';
import { Place } from '@/components/landing/Place';
import { Spaces } from '@/components/landing/Spaces';

// No animate-fade-up here: a transform on an ancestor breaks the fixed nav.
export default function HomePage() {
  return (
    <div className="landing">
      <header>
        <LandingNav />
      </header>
      <main>
        <Hero />
        <Place />
        <Spaces />
        <Complex />
        <HowToBook />
        <FinalCta />
      </main>
      <LandingFooter />
    </div>
  );
}
