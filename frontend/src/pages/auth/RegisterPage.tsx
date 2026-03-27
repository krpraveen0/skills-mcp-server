import { useState } from 'react'
import { Link as RouterLink } from 'react-router-dom'
import {
  Box, Card, CardContent, TextField, Button,
  Typography, Alert, CircularProgress, Divider,
  Link, Paper, InputAdornment, IconButton
} from '@mui/material'
import ContentCopyIcon from '@mui/icons-material/ContentCopy'
import CheckIcon from '@mui/icons-material/Check'
import ExtensionIcon from '@mui/icons-material/Extension'
import PersonIcon from '@mui/icons-material/Person'
import EmailIcon from '@mui/icons-material/Email'
import { apiClient } from '@/services/api'

export function RegisterPage() {
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [generatedKey, setGeneratedKey] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)

  const handleRegister = async () => {
    if (!name.trim()) {
      setError('Name is required')
      return
    }
    if (!email.trim() || !email.includes('@')) {
      setError('A valid email address is required')
      return
    }
    setLoading(true)
    setError('')
    try {
      const { data } = await apiClient.post('/api/v1/auth/register', { name, email })
      setGeneratedKey(data.raw_key)
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
        'Registration failed. Please try again.'
      setError(msg)
    } finally {
      setLoading(false)
    }
  }

  const handleCopy = async () => {
    if (!generatedKey) return
    await navigator.clipboard.writeText(generatedKey)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
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

          {!generatedKey ? (
            <>
              <Typography variant="h5" fontWeight={700} sx={{ mb: 0.5 }}>
                Create a free account
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
                Already have a key?{' '}
                <Link component={RouterLink} to="/auth" fontWeight={600}>
                  Sign in
                </Link>
              </Typography>

              {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

              <TextField
                fullWidth
                label="Your name"
                placeholder="Jane Smith"
                value={name}
                onChange={(e) => setName(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleRegister()}
                InputProps={{
                  startAdornment: (
                    <InputAdornment position="start">
                      <PersonIcon color="action" fontSize="small" />
                    </InputAdornment>
                  ),
                }}
                sx={{ mb: 2 }}
              />

              <TextField
                fullWidth
                label="Email address"
                type="email"
                placeholder="jane@example.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleRegister()}
                InputProps={{
                  startAdornment: (
                    <InputAdornment position="start">
                      <EmailIcon color="action" fontSize="small" />
                    </InputAdornment>
                  ),
                }}
                sx={{ mb: 2 }}
              />

              <Button
                fullWidth
                variant="contained"
                size="large"
                onClick={handleRegister}
                disabled={loading || !name || !email}
                startIcon={loading ? <CircularProgress size={18} /> : undefined}
              >
                {loading ? 'Creating key…' : 'Get my API key'}
              </Button>

              <Typography variant="caption" color="text.secondary" display="block" sx={{ mt: 2, textAlign: 'center' }}>
                Free tier: 100 requests / day
              </Typography>
            </>
          ) : (
            <>
              <Alert severity="success" sx={{ mb: 3 }}>
                Your API key has been created! Copy it now — it won't be shown again.
              </Alert>

              <Typography variant="subtitle2" fontWeight={600} sx={{ mb: 1 }}>
                Your API Key
              </Typography>

              <Paper
                variant="outlined"
                sx={{
                  p: 2, mb: 3,
                  display: 'flex', alignItems: 'center', gap: 1,
                  bgcolor: 'background.default',
                  fontFamily: 'monospace',
                  fontSize: '0.8rem',
                  wordBreak: 'break-all',
                  borderRadius: 1.5,
                }}
              >
                <Typography
                  component="span"
                  sx={{ flex: 1, fontFamily: 'monospace', fontSize: '0.8rem' }}
                >
                  {generatedKey}
                </Typography>
                <IconButton size="small" onClick={handleCopy} sx={{ flexShrink: 0 }}>
                  {copied ? (
                    <CheckIcon fontSize="small" color="success" />
                  ) : (
                    <ContentCopyIcon fontSize="small" />
                  )}
                </IconButton>
              </Paper>

              <Button
                fullWidth
                variant="contained"
                size="large"
                component={RouterLink}
                to="/auth"
              >
                Sign in with this key
              </Button>

              <Divider sx={{ my: 2 }} />

              <Typography variant="body2" color="text.secondary" align="center">
                Or{' '}
                <Link component={RouterLink} to="/" fontWeight={600}>
                  explore skills
                </Link>{' '}
                without signing in.
              </Typography>
            </>
          )}
        </CardContent>
      </Card>
    </Box>
  )
}
