import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import {
  Box, Container, Typography, Grid, Card, CardContent,
  CardActionArea, Chip, CircularProgress, Alert, Button,
  Divider, Stack, Paper, LinearProgress
} from '@mui/material'
import ArrowBackIcon from '@mui/icons-material/ArrowBack'
import StarIcon from '@mui/icons-material/Star'
import ForkRightIcon from '@mui/icons-material/ForkRight'
import RemoveRedEyeIcon from '@mui/icons-material/RemoveRedEye'
import AutoStoriesIcon from '@mui/icons-material/AutoStories'
import OpenInNewIcon from '@mui/icons-material/OpenInNew'
import { reposService, type Skill } from '@/services/skills.service'
import { ScoreBadge } from '@/components/skills/ScoreBadge'

function formatNum(n: number) {
  return n >= 1000 ? `${(n / 1000).toFixed(1)}k` : String(n)
}

function scoreColor(s: number) {
  if (s >= 75) return '#10b981'
  if (s >= 50) return '#6366f1'
  if (s >= 25) return '#f59e0b'
  return '#ef4444'
}

function SkillListCard({ skill, onClick }: { skill: Skill; onClick: () => void }) {
  return (
    <Card sx={{ transition: 'box-shadow 0.15s', '&:hover': { boxShadow: 4 } }}>
      <CardActionArea onClick={onClick}>
        <CardContent sx={{ p: 2 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 1 }}>
            <Box sx={{ flex: 1, minWidth: 0, pr: 1 }}>
              <Typography variant="subtitle2" fontWeight={700} noWrap>
                {skill.title || skill.file_path}
              </Typography>
              {skill.description && (
                <Typography
                  variant="caption"
                  color="text.secondary"
                  sx={{ display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', overflow: 'hidden' }}
                >
                  {skill.description}
                </Typography>
              )}
            </Box>
            <ScoreBadge score={skill.score} size="small" />
          </Box>

          {/* Score bar */}
          <Box sx={{ mb: 1 }}>
            <LinearProgress
              variant="determinate"
              value={Math.min(skill.score, 100)}
              sx={{
                height: 3, borderRadius: 2, bgcolor: 'divider',
                '& .MuiLinearProgress-bar': { bgcolor: scoreColor(skill.score), borderRadius: 2 },
              }}
            />
          </Box>

          {(skill.tags ?? []).length > 0 && (
            <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
              {(skill.tags ?? []).slice(0, 5).map((tag) => (
                <Chip
                  key={tag}
                  label={tag}
                  size="small"
                  variant="outlined"
                  sx={{ height: 18, fontSize: '0.65rem', borderColor: 'divider', color: 'text.secondary' }}
                />
              ))}
            </Box>
          )}
        </CardContent>
      </CardActionArea>
    </Card>
  )
}

export function RepoDetailPage() {
  const { owner, repo } = useParams<{ owner: string; repo: string }>()
  const navigate = useNavigate()

  const { data, isLoading, error } = useQuery({
    queryKey: ['repos', 'detail', owner, repo],
    queryFn: () => reposService.getRepo(owner!, repo!),
    enabled: !!owner && !!repo,
  })

  if (isLoading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '60vh' }}>
        <CircularProgress />
      </Box>
    )
  }

  if (error || !data) {
    return (
      <Container maxWidth="md" sx={{ py: 4 }}>
        <Alert severity="error" sx={{ mb: 2 }}>
          Repository not found or has no indexed skills.
        </Alert>
        <Button startIcon={<ArrowBackIcon />} onClick={() => navigate('/trending')}>
          Back to Trending
        </Button>
      </Container>
    )
  }

  const avgScore = data.skills.length > 0
    ? data.skills.reduce((sum, s) => sum + s.score, 0) / data.skills.length
    : 0

  const lastUpdated = data.skills[0]?.last_updated_at
    ? new Date(data.skills[0].last_updated_at).toLocaleDateString('en-US', {
        year: 'numeric', month: 'long', day: 'numeric',
      })
    : 'Unknown'

  return (
    <Container maxWidth="xl" sx={{ py: 3 }}>
      <Button
        startIcon={<ArrowBackIcon />}
        onClick={() => navigate('/trending')}
        sx={{ mb: 2, color: 'text.secondary' }}
      >
        Back to Trending
      </Button>

      {/* Repo header */}
      <Paper sx={{ p: 3, mb: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', flexWrap: 'wrap', gap: 2 }}>
          <Box>
            <Typography variant="h4" fontWeight={700} sx={{ mb: 0.5 }}>
              {data.owner}/{data.name}
            </Typography>
            <Button
              size="small"
              variant="outlined"
              startIcon={<OpenInNewIcon />}
              component="a"
              href={data.github_url}
              target="_blank"
              rel="noopener noreferrer"
              sx={{ mt: 1 }}
            >
              View on GitHub
            </Button>
          </Box>

          {/* Stats grid */}
          <Stack direction="row" spacing={3}>
            <Box sx={{ textAlign: 'center' }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, justifyContent: 'center' }}>
                <StarIcon sx={{ color: '#f59e0b', fontSize: 16 }} />
                <Typography variant="h6" fontWeight={700}>{formatNum(data.stars)}</Typography>
              </Box>
              <Typography variant="caption" color="text.secondary">Stars</Typography>
            </Box>
            <Box sx={{ textAlign: 'center' }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, justifyContent: 'center' }}>
                <ForkRightIcon sx={{ color: 'text.secondary', fontSize: 16 }} />
                <Typography variant="h6" fontWeight={700}>{formatNum(data.forks)}</Typography>
              </Box>
              <Typography variant="caption" color="text.secondary">Forks</Typography>
            </Box>
            <Box sx={{ textAlign: 'center' }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, justifyContent: 'center' }}>
                <RemoveRedEyeIcon sx={{ color: 'text.secondary', fontSize: 16 }} />
                <Typography variant="h6" fontWeight={700}>{formatNum(data.watchers)}</Typography>
              </Box>
              <Typography variant="caption" color="text.secondary">Watchers</Typography>
            </Box>
            <Box sx={{ textAlign: 'center' }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, justifyContent: 'center' }}>
                <AutoStoriesIcon sx={{ color: 'text.secondary', fontSize: 16 }} />
                <Typography variant="h6" fontWeight={700}>{data.skill_count}</Typography>
              </Box>
              <Typography variant="caption" color="text.secondary">Skills</Typography>
            </Box>
            <Box sx={{ textAlign: 'center' }}>
              <Typography variant="h6" fontWeight={700} sx={{ color: scoreColor(avgScore) }}>
                {avgScore.toFixed(0)}
              </Typography>
              <Typography variant="caption" color="text.secondary">Avg Score</Typography>
            </Box>
          </Stack>
        </Box>

        <Divider sx={{ my: 2 }} />

        <Typography variant="body2" color="text.secondary">
          Last GitHub push: <strong>{lastUpdated}</strong>
        </Typography>
      </Paper>

      {/* Skills list */}
      <Box sx={{ mb: 2 }}>
        <Typography variant="h6" fontWeight={700}>
          SKILL.md Files · {data.skill_count} indexed
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Click a skill to view full content, quality breakdown, and MCP config
        </Typography>
      </Box>

      <Grid container spacing={2}>
        {data.skills.map((skill) => (
          <Grid item xs={12} sm={6} md={4} key={skill.id}>
            <SkillListCard
              skill={skill}
              onClick={() => navigate(`/skills/${skill.id}`)}
            />
          </Grid>
        ))}
      </Grid>
    </Container>
  )
}
