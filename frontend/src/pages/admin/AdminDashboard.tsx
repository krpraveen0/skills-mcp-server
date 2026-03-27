import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Box, Grid, Card, CardContent, Typography, Button,
  Chip, CircularProgress, Alert, Divider, Container,
  Table, TableBody, TableCell, TableHead, TableRow,
  TextField, Switch, FormControlLabel, Paper, IconButton,
  Collapse
} from '@mui/material'
import RefreshIcon from '@mui/icons-material/Refresh'
import AutorenewIcon from '@mui/icons-material/Autorenew'
import ExtensionIcon from '@mui/icons-material/Extension'
import VpnKeyIcon from '@mui/icons-material/VpnKey'
import ScheduleIcon from '@mui/icons-material/Schedule'
import CheckCircleIcon from '@mui/icons-material/CheckCircle'
import ErrorIcon from '@mui/icons-material/Error'
import DeleteIcon from '@mui/icons-material/Delete'
import AddIcon from '@mui/icons-material/Add'
import ContentCopyIcon from '@mui/icons-material/ContentCopy'
import { adminService, type APIKey } from '@/services/admin.service'
import { useState } from 'react'

export function AdminDashboard() {
  const qc = useQueryClient()
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [newKeyName, setNewKeyName] = useState('')
  const [newKeyEmail, setNewKeyEmail] = useState('')
  const [newKeyRateLimit, setNewKeyRateLimit] = useState('1000')
  const [newKeyIsAdmin, setNewKeyIsAdmin] = useState(false)
  const [createdKey, setCreatedKey] = useState<APIKey | null>(null)
  const [copiedKeyId, setCopiedKeyId] = useState<string | null>(null)

  const { data: stats, isLoading, refetch } = useQuery({
    queryKey: ['admin', 'stats'],
    queryFn: adminService.getStats,
    refetchInterval: 30000,
  })

  const { data: crawlJobs } = useQuery({
    queryKey: ['admin', 'crawl-jobs'],
    queryFn: () => adminService.listCrawlJobs(10),
    refetchInterval: 10000,
  })

  const triggerMutation = useMutation({
    mutationFn: adminService.triggerCrawl,
    onSuccess: () => {
      setTimeout(() => qc.invalidateQueries({ queryKey: ['admin'] }), 2000)
    },
  })

  const { data: keysData, refetch: refetchKeys } = useQuery({
    queryKey: ['admin', 'keys'],
    queryFn: adminService.listKeys,
  })

  const createKeyMutation = useMutation({
    mutationFn: () =>
      adminService.createKey(newKeyName, newKeyEmail, parseInt(newKeyRateLimit) || 1000, newKeyIsAdmin),
    onSuccess: (data) => {
      setCreatedKey(data as unknown as APIKey)
      setNewKeyName('')
      setNewKeyEmail('')
      setNewKeyRateLimit('1000')
      setNewKeyIsAdmin(false)
      refetchKeys()
      qc.invalidateQueries({ queryKey: ['admin', 'stats'] })
    },
  })

  const revokeKeyMutation = useMutation({
    mutationFn: adminService.revokeKey,
    onSuccess: () => refetchKeys(),
  })

  const handleCopyKey = (key: string, id: string) => {
    navigator.clipboard.writeText(key)
    setCopiedKeyId(id)
    setTimeout(() => setCopiedKeyId(null), 2000)
  }

  const statCards = [
    {
      label: 'Total Skills',
      value: stats?.total_skills ?? 0,
      icon: <ExtensionIcon />,
      color: 'primary.main',
    },
    {
      label: 'API Keys',
      value: stats?.total_api_keys ?? 0,
      icon: <VpnKeyIcon />,
      color: 'secondary.main',
    },
    {
      label: 'Added Today',
      value: stats?.skills_added_today ?? 0,
      icon: <AutorenewIcon />,
      color: 'success.main',
    },
    {
      label: 'Last Crawl',
      value: stats?.last_crawl_status ?? '—',
      icon: <ScheduleIcon />,
      color: getCrawlStatusColor(stats?.last_crawl_status),
    },
  ]

  if (isLoading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}>
        <CircularProgress />
      </Box>
    )
  }

  return (
    <Container maxWidth="xl" sx={{ py: 3 }}>
      {/* Header */}
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 3 }}>
        <Typography variant="h4" fontWeight={700}>Admin Dashboard</Typography>
        <Box sx={{ display: 'flex', gap: 1 }}>
          <Button
            variant="outlined"
            startIcon={<RefreshIcon />}
            onClick={() => refetch()}
            size="small"
          >
            Refresh
          </Button>
          <Button
            variant="contained"
            startIcon={<AutorenewIcon />}
            onClick={() => triggerMutation.mutate()}
            disabled={triggerMutation.isPending}
            size="small"
          >
            {triggerMutation.isPending ? 'Triggering…' : 'Trigger Crawl'}
          </Button>
        </Box>
      </Box>

      {triggerMutation.isSuccess && (
        <Alert severity="info" sx={{ mb: 2 }}>
          Crawl job triggered — it's running in the background.
        </Alert>
      )}

      {/* Stats cards */}
      <Grid container spacing={2} sx={{ mb: 3 }}>
        {statCards.map((card) => (
          <Grid item xs={12} sm={6} md={3} key={card.label}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 1 }}>
                  <Typography variant="body2" color="text.secondary">{card.label}</Typography>
                  <Box sx={{ color: card.color }}>{card.icon}</Box>
                </Box>
                <Typography variant="h4" fontWeight={700} sx={{ color: card.color }}>
                  {typeof card.value === 'number' ? card.value.toLocaleString() : card.value}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      {/* Crawl Jobs */}
      <Card>
        <CardContent>
          <Typography variant="h6" fontWeight={600} sx={{ mb: 2 }}>
            Recent Crawl Jobs
          </Typography>
          <Divider sx={{ mb: 2 }} />
          {crawlJobs?.jobs && crawlJobs.jobs.length > 0 ? (
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Status</TableCell>
                  <TableCell>Started</TableCell>
                  <TableCell align="right">Found</TableCell>
                  <TableCell align="right">New</TableCell>
                  <TableCell align="right">Updated</TableCell>
                  <TableCell align="right">API Calls</TableCell>
                  <TableCell>Duration</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {crawlJobs.jobs.map((job) => (
                  <TableRow key={job.id} hover>
                    <TableCell>
                      <Chip
                        size="small"
                        label={job.status}
                        color={
                          job.status === 'completed' ? 'success' :
                          job.status === 'failed' ? 'error' :
                          job.status === 'running' ? 'info' : 'default'
                        }
                        icon={
                          job.status === 'completed' ? <CheckCircleIcon /> :
                          job.status === 'failed' ? <ErrorIcon /> : undefined
                        }
                      />
                    </TableCell>
                    <TableCell sx={{ fontFamily: 'monospace', fontSize: '0.75rem' }}>
                      {new Date(job.started_at).toLocaleString()}
                    </TableCell>
                    <TableCell align="right">{job.skills_found}</TableCell>
                    <TableCell align="right" sx={{ color: 'success.main' }}>{job.skills_new}</TableCell>
                    <TableCell align="right">{job.skills_updated}</TableCell>
                    <TableCell align="right">{job.github_queries}</TableCell>
                    <TableCell sx={{ fontSize: '0.75rem', color: 'text.secondary' }}>
                      {job.completed_at
                        ? formatDuration(
                            new Date(job.completed_at).getTime() - new Date(job.started_at).getTime()
                          )
                        : '…'}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : (
            <Typography variant="body2" color="text.secondary">
              No crawl jobs yet. Trigger your first crawl above.
            </Typography>
          )}
        </CardContent>
      </Card>

      {/* API Keys Management */}
      <Card sx={{ mt: 3 }}>
        <CardContent>
          <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
            <Typography variant="h6" fontWeight={600}>API Keys</Typography>
            <Button
              size="small"
              variant="contained"
              startIcon={<AddIcon />}
              onClick={() => { setShowCreateForm((v) => !v); setCreatedKey(null) }}
            >
              {showCreateForm ? 'Cancel' : 'Create Key'}
            </Button>
          </Box>
          <Divider sx={{ mb: 2 }} />

          {/* Create key form */}
          <Collapse in={showCreateForm}>
            <Paper variant="outlined" sx={{ p: 2, mb: 2, borderRadius: 1.5 }}>
              <Typography variant="subtitle2" fontWeight={600} sx={{ mb: 1.5 }}>New API Key</Typography>
              {createdKey?.raw_key && (
                <Alert severity="success" sx={{ mb: 2 }}>
                  <Typography variant="body2" sx={{ mb: 0.5 }}>Key created! Copy it now — it won't be shown again:</Typography>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <Typography sx={{ fontFamily: 'monospace', fontSize: '0.8rem', flex: 1 }}>
                      {createdKey.raw_key}
                    </Typography>
                    <IconButton size="small" onClick={() => handleCopyKey(createdKey.raw_key!, createdKey.id)}>
                      {copiedKeyId === createdKey.id ? <CheckCircleIcon fontSize="small" color="success" /> : <ContentCopyIcon fontSize="small" />}
                    </IconButton>
                  </Box>
                </Alert>
              )}
              <Grid container spacing={2}>
                <Grid item xs={12} sm={4}>
                  <TextField
                    fullWidth size="small" label="Name" required
                    value={newKeyName} onChange={(e) => setNewKeyName(e.target.value)}
                  />
                </Grid>
                <Grid item xs={12} sm={4}>
                  <TextField
                    fullWidth size="small" label="Email"
                    value={newKeyEmail} onChange={(e) => setNewKeyEmail(e.target.value)}
                  />
                </Grid>
                <Grid item xs={12} sm={2}>
                  <TextField
                    fullWidth size="small" label="Rate limit/day" type="number"
                    value={newKeyRateLimit} onChange={(e) => setNewKeyRateLimit(e.target.value)}
                  />
                </Grid>
                <Grid item xs={12} sm={2} sx={{ display: 'flex', alignItems: 'center' }}>
                  <FormControlLabel
                    control={
                      <Switch
                        size="small"
                        checked={newKeyIsAdmin}
                        onChange={(e) => setNewKeyIsAdmin(e.target.checked)}
                      />
                    }
                    label={<Typography variant="body2">Admin</Typography>}
                  />
                </Grid>
                <Grid item xs={12}>
                  <Button
                    variant="contained" size="small"
                    onClick={() => createKeyMutation.mutate()}
                    disabled={createKeyMutation.isPending || !newKeyName}
                    startIcon={createKeyMutation.isPending ? <CircularProgress size={14} /> : <AddIcon />}
                  >
                    {createKeyMutation.isPending ? 'Creating…' : 'Create'}
                  </Button>
                </Grid>
              </Grid>
            </Paper>
          </Collapse>

          {/* Keys table */}
          {keysData?.keys && keysData.keys.length > 0 ? (
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Name</TableCell>
                  <TableCell>Email</TableCell>
                  <TableCell>Prefix</TableCell>
                  <TableCell align="center">Admin</TableCell>
                  <TableCell align="right">Rate limit</TableCell>
                  <TableCell align="right">Calls today</TableCell>
                  <TableCell>Status</TableCell>
                  <TableCell>Created</TableCell>
                  <TableCell align="right">Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {keysData.keys.map((key) => (
                  <TableRow key={key.id} hover>
                    <TableCell sx={{ fontWeight: 500 }}>{key.name}</TableCell>
                    <TableCell sx={{ color: 'text.secondary', fontSize: '0.8rem' }}>{key.user_email || '—'}</TableCell>
                    <TableCell sx={{ fontFamily: 'monospace', fontSize: '0.75rem' }}>{key.key_prefix}</TableCell>
                    <TableCell align="center">
                      {key.is_admin && <Chip label="Admin" size="small" color="warning" />}
                    </TableCell>
                    <TableCell align="right">{key.rate_limit.toLocaleString()}</TableCell>
                    <TableCell align="right">{key.calls_today}</TableCell>
                    <TableCell>
                      <Chip
                        size="small"
                        label={key.is_active ? 'Active' : 'Revoked'}
                        color={key.is_active ? 'success' : 'default'}
                      />
                    </TableCell>
                    <TableCell sx={{ fontSize: '0.75rem', color: 'text.secondary' }}>
                      {new Date(key.created_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell align="right">
                      {key.is_active && (
                        <IconButton
                          size="small"
                          color="error"
                          onClick={() => revokeKeyMutation.mutate(key.id)}
                          disabled={revokeKeyMutation.isPending}
                        >
                          <DeleteIcon fontSize="small" />
                        </IconButton>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : (
            <Typography variant="body2" color="text.secondary">
              No API keys yet. Create one above.
            </Typography>
          )}
        </CardContent>
      </Card>
    </Container>
  )
}

function getCrawlStatusColor(status?: string): string {
  if (status === 'completed') return 'success.main'
  if (status === 'failed') return 'error.main'
  if (status === 'running') return 'info.main'
  return 'text.secondary'
}

function formatDuration(ms: number): string {
  const s = Math.floor(ms / 1000)
  if (s < 60) return `${s}s`
  const m = Math.floor(s / 60)
  return `${m}m ${s % 60}s`
}
