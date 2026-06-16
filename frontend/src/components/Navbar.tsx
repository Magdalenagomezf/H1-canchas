import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import Box from '@mui/material/Box';
import Drawer from '@mui/material/Drawer';
import List from '@mui/material/List';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemText from '@mui/material/ListItemText';
import MenuIcon from '@mui/icons-material/Menu';
import CloseIcon from '@mui/icons-material/Close';
import { motion } from 'framer-motion';
import { useAuth } from '../hooks/useAuth';

const MotionButton = motion(Button);

export default function Navbar() {
  const { isAuthenticated, user, logout } = useAuth();
  const navigate = useNavigate();
  const [drawerOpen, setDrawerOpen] = useState(false);

  const handleLogout = () => {
    logout();
    navigate('/');
  };

  return (
    <AppBar
      position="sticky"
      elevation={0}
      sx={{
        backgroundColor: 'rgba(239,239,236,0.92)',
        backdropFilter: 'blur(12px)',
        borderBottom: '1px solid rgba(0,0,0,0.08)',
      }}
    >
      <Toolbar sx={{ maxWidth: 1200, width: '100%', mx: 'auto', px: { xs: 2, md: 4 } }}>
        {/* Logo */}
        <Typography
          component={Link}
          to="/"
          variant="h5"
          sx={{
            flexGrow: 1,
            textDecoration: 'none',
            color: 'primary.dark',
            fontFamily: '"DM Serif Display", serif',
            fontSize: { xs: '1.3rem', md: '1.5rem' },
          }}
        >
          H1 Canchas
        </Typography>

        {/* Desktop nav */}
        <Box sx={{ display: { xs: 'none', md: 'flex' }, gap: 1, alignItems: 'center' }}>
          <Button component={Link} to="/" sx={{ color: 'text.primary' }}>
            Inicio
          </Button>
          <Button component={Link} to="/canchas" sx={{ color: 'text.primary' }}>
            Canchas
          </Button>

          {isAuthenticated ? (
            <>
              <Button component={Link} to="/mis-reservas" sx={{ color: 'text.primary' }}>
                Mis reservas
              </Button>
              {(user?.role === 'receptionist' || user?.role === 'admin') && (
                <Button component={Link} to="/panel" sx={{ color: 'text.primary' }}>
                  Panel
                </Button>
              )}
              <MotionButton
                variant="outlined"
                color="primary"
                onClick={handleLogout}
                whileHover={{ scale: 1.04 }}
                whileTap={{ scale: 0.97 }}
                sx={{ ml: 1 }}
              >
                Salir
              </MotionButton>
            </>
          ) : (
            <MotionButton
              variant="contained"
              color="primary"
              component={Link}
              to="/login"
              whileHover={{ scale: 1.04 }}
              whileTap={{ scale: 0.97 }}
              sx={{ ml: 1 }}
            >
              Iniciar sesión
            </MotionButton>
          )}
        </Box>

        {/* Mobile hamburger */}
        <IconButton
          sx={{ display: { xs: 'flex', md: 'none' }, color: 'text.primary' }}
          onClick={() => setDrawerOpen(true)}
        >
          <MenuIcon />
        </IconButton>
      </Toolbar>

      {/* Mobile Drawer */}
      <Drawer anchor="right" open={drawerOpen} onClose={() => setDrawerOpen(false)}>
        <Box sx={{ width: 260, p: 2 }}>
          <Box sx={{ display: 'flex', justifyContent: 'flex-end' }}>
            <IconButton onClick={() => setDrawerOpen(false)}>
              <CloseIcon />
            </IconButton>
          </Box>
          <List>
            {[
              { label: 'Inicio', to: '/' },
              { label: 'Canchas', to: '/canchas' },
              ...(isAuthenticated ? [{ label: 'Mis reservas', to: '/mis-reservas' }] : []),
              ...((user?.role === 'receptionist' || user?.role === 'admin')
                ? [{ label: 'Panel', to: '/panel' }]
                : []),
            ].map((item) => (
              <ListItemButton
                key={item.to}
                component={Link}
                to={item.to}
                onClick={() => setDrawerOpen(false)}
              >
                <ListItemText primary={item.label} />
              </ListItemButton>
            ))}
            <ListItemButton
              onClick={() => {
                setDrawerOpen(false);
                isAuthenticated ? handleLogout() : navigate('/login');
              }}
            >
              <ListItemText
                primary={isAuthenticated ? 'Salir' : 'Iniciar sesión'}
                primaryTypographyProps={{ color: 'primary.main', fontWeight: 600 }}
              />
            </ListItemButton>
          </List>
        </Box>
      </Drawer>
    </AppBar>
  );
}
