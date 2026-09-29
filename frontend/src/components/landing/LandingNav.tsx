import { useEffect, useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { Menu, X } from 'lucide-react';
import { Sheet, SheetClose, SheetContent, SheetTitle, SheetTrigger } from '@/components/ui/sheet';
import { buttonVariants } from '@/components/ui/button';
import { useAuth } from '@/hooks/useAuth';
import { cn } from '@/lib/utils';
import { NAV_LINKS } from './content';
import { CONTAINER, MONO } from './ui';

interface LandingNavProps {
  /** Always render the solid (night) background, e.g. on pages without a hero. */
  solid?: boolean;
  /** Hide the account link (e.g. on the login page itself). */
  hideAccountLink?: boolean;
}

export function LandingNav({ solid = false, hideAccountLink = false }: LandingNavProps) {
  const { isAuthenticated, user, logout } = useAuth();
  const { pathname } = useLocation();
  const navigate = useNavigate();
  // Section anchors live on "/": from any other route, prefix them so they navigate home first.
  const anchor = (href: string) => (pathname === '/' ? href : `/${href}`);
  const [scrolled, setScrolled] = useState(false);
  const [open, setOpen] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 32);
    onScroll();
    window.addEventListener('scroll', onScroll, { passive: true });
    return () => window.removeEventListener('scroll', onScroll);
  }, []);

  const isStaff = user?.role === 'receptionist' || user?.role === 'admin';
  const accountLinks = hideAccountLink
    ? []
    : isAuthenticated
      ? [
          { label: 'Mis reservas', to: '/mis-reservas' },
          ...(isStaff ? [{ label: 'Panel', to: '/panel' }] : []),
        ]
      : [{ label: 'Iniciar sesión', to: '/login' }];

  const handleLogout = () => {
    logout();
    setOpen(false);
    navigate('/');
  };

  const linkClass =
    'font-arch text-sm font-medium tracking-wide text-paper/80 transition-colors hover:text-paper';

  return (
    <nav
      aria-label="Principal"
      className={cn(
        'on-dark fixed inset-x-0 top-0 z-40 transition-colors duration-500',
        solid || scrolled || open ? 'bg-night' : 'bg-transparent',
      )}
    >
      <div className={cn(CONTAINER, 'flex h-16 items-center justify-between md:h-20')}>
        <Link to="/" className="flex items-baseline gap-3 text-paper">
          <span className="font-arch font-expanded text-2xl font-extrabold leading-none tracking-[-0.02em]">
            H1
          </span>
          <span className={cn(MONO, 'hidden text-concrete sm:inline')}>Espacio deportivo</span>
        </Link>

        <div className="hidden items-center gap-10 lg:flex">
          {NAV_LINKS.map((link) => (
            <a key={link.href} href={anchor(link.href)} className={linkClass}>
              {link.label}
            </a>
          ))}
          {accountLinks.map((link) => (
            <Link key={link.to} to={link.to} className={linkClass}>
              {link.label}
            </Link>
          ))}
          {isAuthenticated && (
            <button type="button" onClick={handleLogout} className={linkClass}>
              Salir
            </button>
          )}
          <Link to="/canchas" className={cn(buttonVariants({ variant: 'court' }), 'h-11 px-6 text-sm font-semibold')}>
            Reservar
          </Link>
        </div>

        <div className="flex items-center gap-3 lg:hidden">
          <Link to="/canchas" className={cn(buttonVariants({ variant: 'court' }), 'h-10 px-5 text-sm font-semibold')}>
            Reservar
          </Link>
          <Sheet open={open} onOpenChange={setOpen}>
            <SheetTrigger
              aria-label="Abrir menú"
              className="inline-flex size-10 items-center justify-center rounded-none text-paper outline-none focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-paper"
            >
              <Menu size={24} />
            </SheetTrigger>
            <SheetContent
              side="right"
              showCloseButton={false}
              className="on-dark landing-sheet h-full w-full max-w-none gap-0 rounded-none border-0 bg-night p-0 font-arch text-paper shadow-none data-[side=right]:w-full data-[side=right]:border-0 data-[side=right]:sm:max-w-none"
            >
              <SheetTitle className="sr-only">Menú</SheetTitle>
              <div className={cn(CONTAINER, 'flex h-16 items-center justify-between')}>
                <span className="font-arch font-expanded text-2xl font-extrabold leading-none tracking-[-0.02em]">
                  H1
                </span>
                <SheetClose
                  aria-label="Cerrar menú"
                  className="inline-flex size-10 items-center justify-center rounded-none text-paper outline-none focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-paper"
                >
                  <X size={24} />
                </SheetClose>
              </div>
              <div className={cn(CONTAINER, 'flex flex-1 flex-col justify-center gap-6 pb-16')}>
                {[
                  ...NAV_LINKS.map((l) => ({ label: l.label, href: anchor(l.href), internal: false })),
                  ...accountLinks.map((l) => ({ label: l.label, href: l.to, internal: true })),
                ].map(
                  (item) => {
                    const cls =
                      'font-arch font-expanded text-[clamp(2rem,9vw,3.5rem)] font-bold uppercase leading-[1] tracking-[-0.02em] text-paper outline-none focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-paper';
                    return item.internal ? (
                      <Link key={item.href} to={item.href} onClick={() => setOpen(false)} className={cls}>
                        {item.label}
                      </Link>
                    ) : (
                      <a key={item.href} href={item.href} onClick={() => setOpen(false)} className={cls}>
                        {item.label}
                      </a>
                    );
                  },
                )}
                {isAuthenticated && (
                  <button
                    type="button"
                    onClick={handleLogout}
                    className="self-start font-jb text-sm uppercase tracking-[0.12em] text-paper/80 outline-none hover:text-paper focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-paper"
                  >
                    Salir
                  </button>
                )}
                <div aria-hidden="true" className="mt-4 h-[2px] w-24 bg-light" />
              </div>
            </SheetContent>
          </Sheet>
        </div>
      </div>
    </nav>
  );
}
