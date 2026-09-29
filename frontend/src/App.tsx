import { Routes, Route, Navigate, Outlet, useLocation, matchPath } from 'react-router-dom';
import Navbar from './components/Navbar';
import { useAuth } from './hooks/useAuth';
import HomePage from './pages/HomePage';
import LoginPage from './pages/LoginPage';
import SpacesPage from './pages/SpacesPage';
import SpaceDetailPage from './pages/SpaceDetailPage';
import MyBookingsPage from './pages/MyBookingsPage';
import StaffPanelPage from './pages/StaffPanelPage';
import PaymentResultPage from './pages/PaymentResultPage';

function PrivateRoute() {
  const { isAuthenticated } = useAuth();
  const location = useLocation();
  return isAuthenticated ? <Outlet /> : <Navigate to="/login" state={{ from: location }} replace />;
}

function PublicOnlyRoute() {
  const { isAuthenticated } = useAuth();
  const location = useLocation();
  const state = location.state as { from?: { pathname: string; search?: string }; pendingBooking?: { date: string; slotId: number } } | null;
  if (!isAuthenticated) return <Outlet />;
  const target = state?.from ? state.from.pathname + (state.from.search ?? '') : '/';
  return <Navigate to={target} state={state?.pendingBooking ? { pendingBooking: state.pendingBooking } : undefined} replace />;
}

function StaffRoute() {
  const { isAuthenticated, user } = useAuth();
  if (!isAuthenticated) return <Navigate to="/login" replace />;
  if (user?.role === 'customer') return <Navigate to="/" replace />;
  return <Outlet />;
}

// Routes that render their own landing-styled nav instead of the legacy Navbar.
const LANDING_PATHS = ['/', '/login', '/canchas', '/canchas/:id', '/mis-reservas', '/pago/resultado/:bookingId', '/panel'];

const isLandingPath = (pathname: string) =>
  LANDING_PATHS.some((path) => matchPath({ path, end: true }, pathname));

export default function App() {
  const { pathname } = useLocation();
  return (
    <div className="min-h-screen bg-bg">
      {!isLandingPath(pathname) && <Navbar />}
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/canchas" element={<SpacesPage />} />
        <Route path="/canchas/:id" element={<SpaceDetailPage />} />

        <Route element={<PublicOnlyRoute />}>
          <Route path="/login" element={<LoginPage />} />
        </Route>

        <Route element={<PrivateRoute />}>
          <Route path="/mis-reservas" element={<MyBookingsPage />} />
          <Route path="/pago/resultado/:bookingId" element={<PaymentResultPage />} />
        </Route>

        <Route element={<StaffRoute />}>
          <Route path="/panel" element={<StaffPanelPage />} />
        </Route>
      </Routes>
    </div>
  );
}
