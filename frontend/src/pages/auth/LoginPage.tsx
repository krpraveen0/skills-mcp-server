import { useState } from 'react'
import { useNavigate, Link as RouterLink } from 'react-router-dom'
import {
  Box, Card, CardContent, TextField, Button,
  Typography, Alert, CircularProgress, InputAdornment,
  IconButton, Divider, Link
} from '@mui/material'
import VisibilityIcon from '@mui/icons-material/Visibility'
import VisibilityOffIcon from '@mui/icons-material/VisibilityOff'
import KeyIcon from '@mui/icons-material/Key'
import ExtensionIcon from '@mui/icons-material/Extension'
import { useAppStore } from '@/store/useAppStore'
import { apiClient } from '@/services/api'

export function LoginPage() {
  const navigate = useNavigate()
  const setApiKey = useAppStore((s) => s.setApiKey)
  const [key, setKey] = useState('')
  const [showKey, setShowKey] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleLogin = async () => {
    if (key.trim().length < 8) {
      setError('Please enter a valid API key')
      return
    }
    setLoading(true)
    setError('')
    try {
      // Validate key via /auth/me — returns is_admin flag too
      const { data } = await apiClient.get('/api/v1/auth/me', {
        headers: { Authorization: `Bearer ${key}` },
      })
      setApiKey(key, data.is_admin === true)
      navigate('/')
    } catch {
      setError('Invalid API key or server unreachable. Please check your key and try again.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Box
      sx={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        bgcolor: 'background.default',
        p: 2,
      }}
    >
      <Card sx={{ maxWidth: 440, width: '100%' }}>
        <CardContent sx={{ p: 4 }}>
          {/* Logo */}
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 3 }}>
            <Box
              sx={{
                width: 44, height: 44,
                borderRadius: 2,
                bgcolor: 'primary.main',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
              }}
            >
              <ExtensionIcon sx={{ color: 'white' }} />
            </Box>
            <Box>
              <Typography variant="h6" fontWeight={700} lineHeight={1}>
                Skills MCP Server
              </Typography>
              <Typography variant="caption" color="text.secondary">
                Discover the best SKILL.md files on GitHub
              </Typography>
            </Box>
          </Box>

          <Typography variant="h5" fontWeight={700} sx={{ mb: 0.5 }}>
            Sign in with API Key
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
            Don't have a key?{' '}
            <Link component={RouterLink} to="/register" fontWeight={600}>
              Create a free account
            </Link>
          </Typography>

          {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

          <TextField
            fullWidth
            label="API Key"
            type={showKey ? 'text' : 'password'}
            placeholder="sk_live_... or your ADMIN_API_KEY"
            value={key}
            onChange={(e) => setKey(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleLogin()}
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <KeyIcon color="action" fontSize="small" />
                </InputAdornment>
              ),
              endAdornment: (
                <InputAdornment position="end">
                  <IconButton size="small" onClick={() => setShowKey(!showKey)}>
                    {showKey ? <VisibilityOffIcon fontSize="small" /> : <VisibilityIcon fontSize="small" />}
                  </IconButton>
                </InputAdornment>
              ),
            }}
            sx={{ mb: 2 }}
          />

          <Button
            fullWidth
            variant="contained"
            size="large"
            onClick={handleLogin}
            disabled={loading || !key}
            startIcon={loading ? <CircularProgress size={18} /> : undefined}
          >
            {loading ? 'Verifying…' : 'Connect'}
          </Button>

          <Divider sx={{ my: 2 }} />

          <Typography variant="body2" color="text.secondary" align="center">
            Want to browse first?{' '}
            <Link component={RouterLink} to="/" fontWeight={600}>
              Explore skills
            </Link>{' '}
            without signing in.
          </Typography>
        </CardContent>
      </Card>
    </Box>
  )
}
