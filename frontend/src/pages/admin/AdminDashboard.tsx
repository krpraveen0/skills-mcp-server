import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Box, Grid, Card, CardContent, Typography, Button,
  Chip, CircularProgress, Alert, Divider, Container,
  Table, TableBody, TableCell, TableHead, TableRow
} from '@mui/material'
import RefreshIcon from '@mui/icons-material/Refresh'
import AutorenewIcon from '@mui/icons-material/Autorenew'
import ExtensionIcon from '@mui/icons-material/Extension'
import VpnKeyIcon from '@mui/icons-material/VpnKey'
import ScheduleIcon from '@mui/icons-material/Schedule'
import CheckCircleIcon from '@mui/icons-material/CheckCircle'
import ErrorIcon from '@mui/icons-material/Error'
import { adminService } from '@/services/admin.service'

export function AdminDashboard() {
  const qc = useQueryClient()

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
