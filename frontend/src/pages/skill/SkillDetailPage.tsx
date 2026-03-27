import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import {
  Box, Container, Typography, Chip, CircularProgress,
  Alert, Button, Divider, Paper, Grid, LinearProgress,
  Stack
} from '@mui/material'
import ArrowBackIcon from '@mui/icons-material/ArrowBack'
import StarIcon from '@mui/icons-material/Star'
import ForkRightIcon from '@mui/icons-material/ForkRight'
import OpenInNewIcon from '@mui/icons-material/OpenInNew'
import ContentCopyIcon from '@mui/icons-material/ContentCopy'
import TrendingUpIcon from '@mui/icons-material/TrendingUp'
import DescriptionIcon from '@mui/icons-material/Description'
import { skillsService } from '@/services/skills.service'
import { ScoreBadge } from '@/components/skills/ScoreBadge'
import { useState } from 'react'

export function SkillDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [copied, setCopied] = useState(false)

  const { data: skill, isLoading, error } = useQuery({
    queryKey: ['skill', id],
    queryFn: () => skillsService.getById(id!),
    enabled: !!id,
  })

  const handleCopy = () => {
    if (!skill) return
    navigator.clipboard.writeText(skill.github_url)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  if (isLoading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '60vh' }}>
        <CircularProgress />
      </Box>
    )
  }

  if (error || !skill) {
    return (
      <Container maxWidth="md" sx={{ py: 4 }}>
        <Alert severity="error" sx={{ mb: 2 }}>
          Skill not found or failed to load.
        </Alert>
        <Button startIcon={<ArrowBackIcon />} onClick={() => navigate('/')}>
          Back to Explorer
        </Button>
      </Container>
    )
  }

  const scoreColor = (s: number) => {
    if (s >= 75) return '#10b981'
    if (s >= 50) return '#6366f1'
    if (s >= 25) return '#f59e0b'
    return '#ef4444'
  }

  const formatNum = (n: number) => n >= 1000 ? `${(n / 1000).toFixed(1)}k` : n.toString()

  const lastUpdated = skill.last_updated_at
    ? new Date(skill.last_updated_at).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })
    : 'Unknown'

  return (
    <Container maxWidth="lg" sx={{ py: 3 }}>
      {/* Back button */}
      <Button
        startIcon={<ArrowBackIcon />}
        onClick={() => navigate('/')}
        sx={{ mb: 2, color: 'text.secondary' }}
      >
        Back to Explorer
      </Button>

      {/* Header */}
      <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 2, mb: 3 }}>
        <Box sx={{ flex: 1, minWidth: 0 }}>
          <Typography variant="h4" fontWeight={700} gutterBottom>
            {skill.title}
          </Typography>
          <Typography variant="body1" color="text.secondary" sx={{ mb: 1 }}>
            {skill.repo_owner}/{skill.repo_name} · {skill.file_path}
          </Typography>
          {skill.description && (
            <Typography variant="body1" sx={{ mb: 2 }}>
              {skill.description}
            </Typography>
          )}
          {/* Tags */}
          {(skill.tags ?? []).length > 0 && (
            <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.75, mb: 2 }}>
              {(skill.tags ?? []).map((tag) => (
                <Chip
                  key={tag}
                  label={tag}
                  size="small"
                  variant="outlined"
                  sx={{ borderColor: 'divider', color: 'text.secondary' }}
                />
              ))}
            </Box>
          )}
        </Box>
        <Box sx={{ flexShrink: 0 }}>
          <ScoreBadge score={skill.score} />
        </Box>
      </Box>

      <Grid container spacing={3}>
        {/* Left: content */}
        <Grid item xs={12} md={8}>
          <Paper sx={{ p: 3 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
              <DescriptionIcon sx={{ color: 'text.secondary', fontSize: 20 }} />
              <Typography variant="h6" fontWeight={600}>Skill Content</Typography>
            </Box>
            <Divider sx={{ mb: 2 }} />
            {skill.content ? (
              <Typography
                component="pre"
                sx={{
                  fontFamily: 'monospace',
                  fontSize: '0.8rem',
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-word',
                  color: 'text.primary',
                  bgcolor: 'action.hover',
                  p: 2,
                  borderRadius: 1,
                  maxHeight: 500,
                  overflow: 'auto',
                  lineHeight: 1.6,
                }}
              >
                {skill.content}
              </Typography>
            ) : (
              <Typography color="text.secondary" fontStyle="italic">
                Content not available. View the full file on GitHub.
              </Typography>
            )}
          </Paper>
        </Grid>

        {/* Right: metadata */}
        <Grid item xs={12} md={4}>
          {/* Quality Score */}
          <Paper sx={{ p: 3, mb: 2 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
              <TrendingUpIcon sx={{ color: 'text.secondary', fontSize: 20 }} />
              <Typography variant="h6" fontWeight={600}>Quality Score</Typography>
            </Box>
            <Divider sx={{ mb: 2 }} />
            <Box sx={{ mb: 2 }}>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.5 }}>
                <Typography variant="body2" color="text.secondary">Overall</Typography>
                <Typography variant="body2" fontWeight={700} sx={{ color: scoreColor(skill.score) }}>
                  {skill.score.toFixed(1)}/100
                </Typography>
              </Box>
              <LinearProgress
                variant="determinate"
                value={Math.min(skill.score, 100)}
                sx={{
                  height: 8, borderRadius: 4, bgcolor: 'divider',
                  '& .MuiLinearProgress-bar': { bgcolor: scoreColor(skill.score), borderRadius: 4 },
                }}
              />
            </Box>

            {skill.score_breakdown && (
              <Stack spacing={1.5}>
                {[
                  { label: 'Stars', value: skill.score_breakdown.star_score },
                  { label: 'Adoption', value: skill.score_breakdown.adoption_score },
                  { label: 'Recency', value: skill.score_breakdown.recency_score },
                  { label: 'Composite', value: skill.score_breakdown.composite_score },
                ].map(({ label, value }) => (
                  <Box key={label}>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.25 }}>
                      <Typography variant="caption" color="text.secondary">{label}</Typography>
                      <Typography variant="caption" fontWeight={600}>{value.toFixed(1)}</Typography>
                    </Box>
                    <LinearProgress
                      variant="determinate"
                      value={Math.min(value, 100)}
                      sx={{
                        height: 4, borderRadius: 2, bgcolor: 'divider',
                        '& .MuiLinearProgress-bar': { bgcolor: 'primary.main', borderRadius: 2 },
                      }}
                    />
                  </Box>
                ))}
              </Stack>
            )}
          </Paper>

          {/* Repository stats */}
          <Paper sx={{ p: 3, mb: 2 }}>
            <Typography variant="h6" fontWeight={600} sx={{ mb: 2 }}>Repository</Typography>
            <Divider sx={{ mb: 2 }} />
            <Stack spacing={1.5}>
              <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                  <StarIcon sx={{ fontSize: 16, color: '#f59e0b' }} />
                  <Typography variant="body2" color="text.secondary">Stars</Typography>
                </Box>
                <Typography variant="body2" fontWeight={600}>{formatNum(skill.stars)}</Typography>
              </Box>
              <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                  <ForkRightIcon sx={{ fontSize: 16, color: 'text.secondary' }} />
                  <Typography variant="body2" color="text.secondary">Forks</Typography>
                </Box>
                <Typography variant="body2" fontWeight={600}>{formatNum(skill.forks)}</Typography>
              </Box>
              <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                <Typography variant="body2" color="text.secondary">Last updated</Typography>
                <Typography variant="body2" fontWeight={600}>{lastUpdated}</Typography>
              </Box>
              <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                <Typography variant="body2" color="text.secondary">Indexed</Typography>
                <Typography variant="body2" fontWeight={600}>
                  {new Date(skill.indexed_at).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })}
                </Typography>
              </Box>
            </Stack>
          </Paper>

          {/* Actions */}
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" fontWeight={600} sx={{ mb: 2 }}>Actions</Typography>
            <Divider sx={{ mb: 2 }} />
            <Stack spacing={1.5}>
              <Button
                variant="contained"
                startIcon={<OpenInNewIcon />}
                component="a"
                href={skill.github_url}
                target="_blank"
                rel="noopener noreferrer"
                fullWidth
              >
                View on GitHub
              </Button>
              <Button
                variant="outlined"
                startIcon={<ContentCopyIcon />}
                onClick={handleCopy}
                fullWidth
              >
                {copied ? 'Copied!' : 'Copy GitHub URL'}
              </Button>
            </Stack>
          </Paper>
        </Grid>
      </Grid>
    </Container>
  )
}
