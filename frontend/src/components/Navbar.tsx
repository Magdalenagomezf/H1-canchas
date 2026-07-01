import { useState } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { Menu, X } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { useAuth } from '../hooks/useAuth';
import { cn } from '@/lib/utils';

export default function Navbar() {
  const { isAuthenticated, user, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [mobileOpen, setMobileOpen] = useState(false);

  const handleLogout = () => {
    logout();
    navigate('/');
    setMobileOpen(false);
  };

  const isActive = (to: string) =>
    to === '/' ? location.pathname === '/' : location.pathname.startsWith(to);

  const links = [
    { label: 'Inicio', to: '/' },
    { label: 'Canchas', to: '/canchas' },
    ...(isAuthenticated ? [{ label: 'Mis reservas', to: '/mis-reservas' }] : []),
    ...(user?.role === 'receptionist' || user?.role === 'admin'
      ? [{ label: 'Panel', to: '/panel' }]
      : []),
  ];

  return (
    <>
      <nav className="fixed top-0 inset-x-0 z-50 h-[58px] bg-dark-surface border-b border-white/[0.07] flex items-center justify-between px-6 shadow-[0_2px_16px_rgba(0,0,0,0.2)]">
        <Link to="/" className="font-serif text-xl text-white tracking-tight">
          H1 <span className="text-lime italic">Canchas</span>
        </Link>

        {/* Desktop nav */}
        <div className="hidden md:flex items-center gap-1">
          {links.map((link) => (
            <Link
              key={link.to}
              to={link.to}
              className={cn(
                'px-4 py-1.5 rounded-md text-sm font-medium transition-all duration-normal ease-smooth',
                isActive(link.to)
                  ? 'text-white bg-white/10'
                  : 'text-white/60 hover:text-white hover:bg-white/[0.07]',
              )}
            >
              {link.label}
            </Link>
          ))}
        </div>

        {/* Desktop auth */}
        <div className="hidden md:flex items-center gap-3">
          {isAuthenticated ? (
            <button
              onClick={handleLogout}
              className="text-sm font-semibold text-white/60 hover:text-white transition-all duration-normal active:scale-[0.98]"
            >
              Salir
            </button>
          ) : (
            <Link
              to="/login"
              className="bg-primary text-white font-bold rounded-lg px-4 py-2 text-sm transition-all duration-normal ease-smooth hover:bg-primary-dark hover:scale-[1.02] hover:shadow-glow active:scale-[0.98]"
            >
              Iniciar sesión
            </Link>
          )}
        </div>

        {/* Mobile hamburger */}
        <button
          className="md:hidden text-white/70 hover:text-white transition-colors"
          onClick={() => setMobileOpen(true)}
        >
          <Menu size={22} />
        </button>
      </nav>

      {/* Spacer */}
      <div className="h-[58px]" />

      {/* Mobile overlay + drawer */}
      <AnimatePresence>
        {mobileOpen && (
          <>
            <motion.div
              key="overlay"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.2 }}
              className="fixed inset-0 z-40 bg-black/50"
              onClick={() => setMobileOpen(false)}
            />
            <motion.div
              key="drawer"
              initial={{ x: '100%' }}
              animate={{ x: 0 }}
              exit={{ x: '100%' }}
              transition={{ type: 'tween', duration: 0.25, ease: [0.4, 0, 0.2, 1] }}
              className="fixed right-0 top-0 bottom-0 z-50 w-[280px] bg-dark-surface border-l border-white/[0.07] flex flex-col p-6"
            >
              <div className="flex items-center justify-between mb-8">
                <span className="font-serif text-xl text-white">
                  H1 <span className="text-lime italic">Canchas</span>
                </span>
                <button
                  onClick={() => setMobileOpen(false)}
                  className="text-white/60 hover:text-white transition-colors"
                >
                  <X size={20} />
                </button>
              </div>

              <div className="flex flex-col gap-1">
                {links.map((link) => (
                  <Link
                    key={link.to}
                    to={link.to}
                    onClick={() => setMobileOpen(false)}
                    className={cn(
                      'px-4 py-3 rounded-lg text-sm font-medium transition-all',
                      isActive(link.to)
                        ? 'text-white bg-white/10'
                        : 'text-white/60 hover:text-white hover:bg-white/[0.07]',
                    )}
                  >
                    {link.label}
                  </Link>
                ))}
              </div>

              <div className="mt-auto pt-6 border-t border-white/[0.07]">
                {isAuthenticated ? (
                  <button
                    onClick={handleLogout}
                    className="w-full text-left px-4 py-3 text-sm font-semibold text-white/60 hover:text-white transition-all rounded-lg hover:bg-white/[0.07] active:scale-[0.98]"
                  >
                    Salir
                  </button>
                ) : (
                  <Link
                    to="/login"
                    onClick={() => setMobileOpen(false)}
                    className="block w-full text-center bg-primary text-white font-bold rounded-lg px-4 py-3 text-sm transition-all hover:bg-primary-dark active:scale-[0.98]"
                  >
                    Iniciar sesión
                  </Link>
                )}
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </>
  );
}
