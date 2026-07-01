import { Routes, Route, Navigate, Outlet, useLocation } from 'react-router-dom';
import Navbar from './components/Navbar';
import { useAuth } from './hooks/useAuth';
import HomePage from './pages/HomePage';
import LoginPage from './pages/LoginPage';
import SpacesPage from './pages/SpacesPage';
import SpaceDetailPage from './pages/SpaceDetailPage';
import MyBookingsPage from './pages/MyBookingsPage';
import StaffPanelPage from './pages/StaffPanelPage';

function PrivateRoute() {
  const { isAuthenticated } = useAuth();
  const location = useLocation();
  return isAuthenticated ? <Outlet /> : <Navigate to="/login" state={{ from: location }} replace />;
}

function PublicOnlyRoute() {
  const { isAuthenticated } = useAuth();
  return isAuthenticated ? <Navigate to="/" replace /> : <Outlet />;
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
        </Route>

        <Route element={<StaffRoute />}>
          <Route path="/panel" element={<StaffPanelPage />} />
        </Route>
      </Routes>
    </div>
  );
}
