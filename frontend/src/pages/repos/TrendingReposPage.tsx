import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import {
  Box, Container, Typography, Grid, Card, CardContent,
  CardActionArea, Chip, CircularProgress, Alert,
  ToggleButtonGroup, ToggleButton, Divider, Stack,
  Tooltip, LinearProgress
} from '@mui/material'
import StarIcon from '@mui/icons-material/Star'
import ForkRightIcon from '@mui/icons-material/ForkRight'
import AutoStoriesIcon from '@mui/icons-material/AutoStories'
import TrendingUpIcon from '@mui/icons-material/TrendingUp'
import WhatshotIcon from '@mui/icons-material/Whatshot'
import CalendarTodayIcon from '@mui/icons-material/CalendarToday'
import OpenInNewIcon from '@mui/icons-material/OpenInNew'
import { reposService, type TrendingRepo } from '@/services/skills.service'

type Period = 'today' | 'week' | 'month' | 'all'
const PERIODS: { value: Period; label: string; icon: React.ReactNode }[] = [
  { value: 'today', label: 'Today', icon: <WhatshotIcon sx={{ fontSize: 14 }} /> },
  { value: 'week', label: 'This Week', icon: <TrendingUpIcon sx={{ fontSize: 14 }} /> },
  { value: 'month', label: 'This Month', icon: <CalendarTodayIcon sx={{ fontSize: 14 }} /> },
  { value: 'all', label: 'All Time', icon: <StarIcon sx={{ fontSize: 14 }} /> },
]

function formatNum(n: number) {
  return n >= 1000 ? `${(n / 1000).toFixed(1)}k` : String(n)
}

function scoreColor(s: number) {
  if (s >= 75) return '#10b981'
  if (s >= 50) return '#6366f1'
  if (s >= 25) return '#f59e0b'
  return '#ef4444'
}

function RepoCard({ repo, rank }: { repo: TrendingRepo; rank: number }) {
  const navigate = useNavigate()

  return (
    <Card
      sx={{
        height: '100%',
        transition: 'transform 0.15s, box-shadow 0.15s',
        '&:hover': { transform: 'translateY(-2px)', boxShadow: 6 },
      }}
    >
      <CardActionArea
        onClick={() => navigate(`/repos/${repo.owner}/${repo.name}`)}
        sx={{ height: '100%', alignItems: 'flex-start' }}
      >
        <CardContent sx={{ p: 2.5, height: '100%', display: 'flex', flexDirection: 'column' }}>
          {/* Rank + name */}
          <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 1.5, mb: 1.5 }}>
            <Box
              sx={{
                minWidth: 28, height: 28,
                borderRadius: '50%',
                bgcolor: rank <= 3 ? 'warning.main' : 'action.selected',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                flexShrink: 0,
              }}
            >
              <Typography
                variant="caption"
                fontWeight={700}
                sx={{ color: rank <= 3 ? 'white' : 'text.secondary' }}
              >
                {rank}
              </Typography>
            </Box>
            <Box sx={{ flex: 1, minWidth: 0 }}>
              <Typography variant="subtitle2" fontWeight={700} noWrap>
                {repo.owner}/{repo.name}
              </Typography>
              {repo.description && (
                <Typography
                  variant="caption"
                  color="text.secondary"
                  sx={{
                    display: '-webkit-box',
                    WebkitLineClamp: 2,
                    WebkitBoxOrient: 'vertical',
                    overflow: 'hidden',
                    lineHeight: 1.4,
                  }}
                >
                  {repo.description}
                </Typography>
              )}
            </Box>
          </Box>

          {/* Stats row */}
          <Stack direction="row" spacing={2} sx={{ mb: 1.5 }}>
            <Tooltip title="GitHub Stars">
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                <StarIcon sx={{ fontSize: 14, color: '#f59e0b' }} />
                <Typography variant="caption" fontWeight={600}>{formatNum(repo.stars)}</Typography>
              </Box>
            </Tooltip>
            <Tooltip title="Forks">
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                <ForkRightIcon sx={{ fontSize: 14, color: 'text.secondary' }} />
                <Typography variant="caption" color="text.secondary">{formatNum(repo.forks)}</Typography>
              </Box>
            </Tooltip>
            <Tooltip title="SKILL.md files indexed">
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                <AutoStoriesIcon sx={{ fontSize: 14, color: 'text.secondary' }} />
                <Typography variant="caption" color="text.secondary">{repo.skill_count} skill{repo.skill_count !== 1 ? 's' : ''}</Typography>
              </Box>
            </Tooltip>
          </Stack>

          {/* Quality score bar */}
          <Box sx={{ mb: 1.5 }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.25 }}>
              <Typography variant="caption" color="text.secondary">Quality</Typography>
              <Typography variant="caption" fontWeight={700} sx={{ color: scoreColor(repo.top_score) }}>
                {repo.top_score.toFixed(0)}/100
              </Typography>
            </Box>
            <LinearProgress
              variant="determinate"
              value={Math.min(repo.top_score, 100)}
              sx={{
                height: 4, borderRadius: 2, bgcolor: 'divider',
                '& .MuiLinearProgress-bar': { bgcolor: scoreColor(repo.top_score), borderRadius: 2 },
              }}
            />
          </Box>

          {/* Tags */}
          {repo.tags.length > 0 && (
            <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5, mt: 'auto' }}>
              {repo.tags.slice(0, 4).map((tag) => (
                <Chip
                  key={tag}
                  label={tag}
                  size="small"
                  variant="outlined"
                  sx={{ height: 18, fontSize: '0.65rem', borderColor: 'divider' }}
                />
              ))}
            </Box>
          )}
        </CardContent>
      </CardActionArea>
    </Card>
  )
}

export function TrendingReposPage() {
  const [period, setPeriod] = useState<Period>('week')

  const { data, isLoading, error } = useQuery({
    queryKey: ['repos', 'trending', period],
    queryFn: () => reposService.getTrending(period, 100, 20),
    staleTime: 30 * 60 * 1000, // 30 min client-side stale
  })

  const lastUpdated = data?.repos?.[0]?.indexed_at
    ? new Date(data.repos[0].indexed_at).toLocaleDateString('en-US', {
        month: 'short', day: 'numeric', year: 'numeric',
      })
    : null

  return (
    <Container maxWidth="xl" sx={{ py: 3 }}>
      {/* Header */}
      <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', mb: 3, flexWrap: 'wrap', gap: 2 }}>
        <Box>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.5 }}>
            <WhatshotIcon sx={{ color: '#f59e0b', fontSize: 28 }} />
            <Typography variant="h4" fontWeight={700}>Trending Repositories</Typography>
          </Box>
          <Typography variant="body2" color="text.secondary">
            Top GitHub repos with SKILL.md files, ranked by popularity · Updated daily
          </Typography>
          {lastUpdated && (
            <Typography variant="caption" color="text.disabled" sx={{ mt: 0.5, display: 'block' }}>
              Data as of {lastUpdated}
            </Typography>
          )}
        </Box>

        {/* Period selector */}
        <ToggleButtonGroup
          value={period}
          exclusive
          onChange={(_, v) => { if (v) setPeriod(v) }}
          size="small"
          sx={{ height: 36 }}
        >
          {PERIODS.map((p) => (
            <ToggleButton key={p.value} value={p.value} sx={{ gap: 0.5, px: 1.5 }}>
              {p.icon}
              <Typography variant="caption" fontWeight={600}>{p.label}</Typography>
            </ToggleButton>
          ))}
        </ToggleButtonGroup>
      </Box>

      <Divider sx={{ mb: 3 }} />

      {/* Period description */}
      <Box sx={{ mb: 2 }}>
        <Typography variant="body2" color="text.secondary">
          {period === 'today' && 'Repos crawled or refreshed in the last 24 hours, ranked by star count.'}
          {period === 'week' && 'Repos with recent GitHub activity in the last 7 days, ranked by star count.'}
          {period === 'month' && 'Repos with recent GitHub activity in the last 30 days, ranked by star count.'}
          {period === 'all' && 'All indexed repos with ≥100 stars, ranked by total star count.'}
          {' '}Only repos with ≥100 ⭐ are shown.
        </Typography>
      </Box>

      {/* Content */}
      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          Failed to load trending repos. Please try again.
        </Alert>
      )}

      {isLoading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}>
          <CircularProgress />
        </Box>
      ) : data?.repos && data.repos.length > 0 ? (
        <Grid container spacing={2}>
          {data.repos.map((repo, i) => (
            <Grid item xs={12} sm={6} md={4} lg={3} key={`${repo.owner}/${repo.name}`}>
              <RepoCard repo={repo} rank={i + 1} />
            </Grid>
          ))}
        </Grid>
      ) : (
        <Box sx={{ textAlign: 'center', py: 8 }}>
          <TrendingUpIcon sx={{ fontSize: 48, color: 'text.disabled', mb: 2 }} />
          <Typography variant="h6" color="text.secondary">No trending repos for this period</Typography>
          <Typography variant="body2" color="text.disabled" sx={{ mt: 1 }}>
            {period === 'today'
              ? 'No repos were crawled today yet. Try "This Week" instead.'
              : 'Try a different time period or check back after the next crawl.'}
          </Typography>
        </Box>
      )}

      {/* GitHub link */}
      {data && data.repos.length > 0 && (
        <Box sx={{ mt: 4, display: 'flex', alignItems: 'center', gap: 1 }}>
          <OpenInNewIcon sx={{ fontSize: 14, color: 'text.disabled' }} />
          <Typography variant="caption" color="text.disabled">
            Stats sourced from GitHub via daily crawl. Star counts reflect the last crawl run.
          </Typography>
        </Box>
      )}
    </Container>
  )
}
