import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import {
  Box, AppBar, Toolbar, Typography, IconButton,
  Drawer, List, ListItemButton, ListItemIcon, ListItemText,
  Tooltip, Chip
} from '@mui/material'
import SearchIcon from '@mui/icons-material/Search'
import DashboardIcon from '@mui/icons-material/Dashboard'
import LogoutIcon from '@mui/icons-material/Logout'
import ExtensionIcon from '@mui/icons-material/Extension'
import VpnKeyIcon from '@mui/icons-material/VpnKey'
import { useAppStore } from '@/store/useAppStore'

const DRAWER_WIDTH = 220

const navItems = [
  { label: 'Explorer', path: '/', icon: <SearchIcon /> },
  { label: 'Admin', path: '/admin', icon: <DashboardIcon /> },
]

export function App() {
  const navigate = useNavigate()
  const location = useLocation()
  const { clearAuth, apiKey } = useAppStore()

  const handleLogout = () => {
    clearAuth()
    navigate('/auth')
  }

  const keyDisplay = apiKey ? apiKey.substring(0, 16) + '…' : 'No key'

  return (
    <Box sx={{ display: 'flex', minHeight: '100vh' }}>
      {/* Sidebar */}
      <Drawer
        variant="permanent"
        sx={{
          width: DRAWER_WIDTH,
          flexShrink: 0,
          '& .MuiDrawer-paper': { width: DRAWER_WIDTH, boxSizing: 'border-box' },
        }}
      >
        {/* Brand */}
        <Box sx={{ p: 2, display: 'flex', alignItems: 'center', gap: 1.5 }}>
          <Box
            sx={{
              width: 32, height: 32, borderRadius: 1.5,
              bgcolor: 'primary.main',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
            }}
          >
            <ExtensionIcon sx={{ fontSize: 18, color: 'white' }} />
          </Box>
          <Box>
            <Typography variant="subtitle2" fontWeight={700} lineHeight={1.2}>
              Skills MCP
            </Typography>
            <Typography variant="caption" color="text.secondary" lineHeight={1}>
              v1.0.0
            </Typography>
          </Box>
        </Box>

        <Box sx={{ px: 1, flex: 1 }}>
          <List dense>
            {navItems.map((item) => (
              <ListItemButton
                key={item.path}
                selected={location.pathname === item.path}
                onClick={() => navigate(item.path)}
                sx={{
                  borderRadius: 1.5,
                  mb: 0.5,
                  '&.Mui-selected': {
                    bgcolor: 'rgba(99, 102, 241, 0.16)',
                    color: 'primary.light',
                    '& .MuiListItemIcon-root': { color: 'primary.light' },
                  },
                }}
              >
                <ListItemIcon sx={{ minWidth: 36 }}>{item.icon}</ListItemIcon>
                <ListItemText
                  primary={item.label}
                  primaryTypographyProps={{ fontSize: '0.875rem', fontWeight: 500 }}
                />
              </ListItemButton>
            ))}
          </List>
        </Box>

        {/* Key display + logout at bottom */}
        <Box sx={{ p: 2, borderTop: '1px solid', borderColor: 'divider' }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
            <VpnKeyIcon sx={{ fontSize: 14, color: 'text.secondary' }} />
            <Chip
              label={keyDisplay}
              size="small"
              sx={{ fontSize: '0.65rem', height: 20, flex: 1, justifyContent: 'flex-start' }}
            />
          </Box>
          <Tooltip title="Sign out">
            <IconButton
              onClick={handleLogout}
              size="small"
              sx={{ width: '100%', borderRadius: 1.5, justifyContent: 'flex-start', px: 1 }}
            >
              <LogoutIcon fontSize="small" sx={{ mr: 1 }} />
              <Typography variant="caption" color="text.secondary">Sign out</Typography>
            </IconButton>
          </Tooltip>
        </Box>
      </Drawer>

      {/* Main content */}
      <Box
        component="main"
        sx={{
          flex: 1,
          bgcolor: 'background.default',
          minHeight: '100vh',
          overflow: 'auto',
        }}
      >
        <AppBar position="static" elevation={0}>
          <Toolbar variant="dense" sx={{ justifyContent: 'flex-end' }}>
            <Typography variant="caption" color="text.secondary">
              Skills MCP Server — Agentic AI Skills Discovery
            </Typography>
          </Toolbar>
        </AppBar>
        <Outlet />
      </Box>
    </Box>
  )
}
