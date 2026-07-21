import { Routes, Route, Navigate, Outlet, useLocation } from 'react-router-dom';
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

export default function App() {
  return (
    <div className="min-h-screen bg-bg">
      <Navbar />
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
