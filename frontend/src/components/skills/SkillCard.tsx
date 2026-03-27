import {
  Card, CardContent, CardActions, Box, Typography,
  Chip, IconButton, Tooltip, LinearProgress, Stack
} from '@mui/material'
import StarIcon from '@mui/icons-material/Star'
import ForkRightIcon from '@mui/icons-material/ForkRight'
import OpenInNewIcon from '@mui/icons-material/OpenInNew'
import ContentCopyIcon from '@mui/icons-material/ContentCopy'
import TrendingUpIcon from '@mui/icons-material/TrendingUp'
import type { Skill } from '@/services/skills.service'
import { ScoreBadge } from './ScoreBadge'
import { useState } from 'react'

interface SkillCardProps {
  skill: Skill
  onClick?: () => void
  rank?: number
}

export function SkillCard({ skill, onClick, rank }: SkillCardProps) {
  const [copied, setCopied] = useState(false)

  const handleCopy = (e: React.MouseEvent) => {
    e.stopPropagation()
    navigator.clipboard.writeText(skill.github_url)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const timeAgo = skill.last_updated_at
    ? formatTimeAgo(new Date(skill.last_updated_at))
    : 'Unknown'

  return (
    <Card
      sx={{
        cursor: onClick ? 'pointer' : 'default',
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
      }}
      onClick={onClick}
    >
      <CardContent sx={{ flex: 1 }}>
        <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 1, mb: 1 }}>
          {rank && (
            <Box
              sx={{
                minWidth: 28, height: 28,
                borderRadius: '50%',
                bgcolor: rank <= 3 ? 'primary.main' : 'background.default',
                border: '1px solid',
                borderColor: rank <= 3 ? 'primary.main' : 'divider',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                fontSize: '0.75rem', fontWeight: 700, color: rank <= 3 ? 'white' : 'text.secondary',
                flexShrink: 0,
              }}
            >
              {rank}
            </Box>
          )}
          <Box sx={{ flex: 1, minWidth: 0 }}>
            <Typography
              variant="subtitle1"
              fontWeight={600}
              noWrap
              sx={{ mb: 0.25 }}
            >
              {skill.title}
            </Typography>
            <Typography variant="caption" color="text.secondary" noWrap display="block">
              {skill.repo_owner}/{skill.repo_name} · {skill.file_path}
            </Typography>
          </Box>
          <ScoreBadge score={skill.score} />
        </Box>

        {/* Description */}
        <Typography
          variant="body2"
          color="text.secondary"
          sx={{
            display: '-webkit-box',
            WebkitLineClamp: 2,
            WebkitBoxOrient: 'vertical',
            overflow: 'hidden',
            mb: 1.5,
            minHeight: '2.4em',
          }}
        >
          {skill.description || 'No description available.'}
        </Typography>

        {/* Tags */}
        {(skill.tags ?? []).length > 0 && (
          <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5, mb: 1.5 }}>
            {(skill.tags ?? []).slice(0, 5).map((tag) => (
              <Chip
                key={tag}
                label={tag}
                size="small"
                variant="outlined"
                sx={{ borderColor: 'divider', color: 'text.secondary', height: 20 }}
              />
            ))}
            {(skill.tags ?? []).length > 5 && (
              <Chip
                label={`+${(skill.tags ?? []).length - 5}`}
                size="small"
                sx={{ height: 20, bgcolor: 'action.selected' }}
              />
            )}
          </Box>
        )}

        {/* Score breakdown mini bar */}
        <Box sx={{ mb: 1 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.25 }}>
            <Typography variant="caption" color="text.secondary">
              <TrendingUpIcon sx={{ fontSize: 12, mr: 0.25, verticalAlign: 'middle' }} />
              Quality Score
            </Typography>
            <Typography variant="caption" fontWeight={600} color="primary.light">
              {skill.score.toFixed(1)}/100
            </Typography>
          </Box>
          <LinearProgress
            variant="determinate"
            value={skill.score}
            sx={{
              height: 4,
              borderRadius: 2,
              bgcolor: 'divider',
              '& .MuiLinearProgress-bar': {
                bgcolor: getScoreColor(skill.score),
                borderRadius: 2,
              },
            }}
          />
        </Box>
      </CardContent>

      <CardActions sx={{ px: 2, pb: 1.5, pt: 0, justifyContent: 'space-between' }}>
        <Stack direction="row" spacing={1.5}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
            <StarIcon sx={{ fontSize: 14, color: '#f59e0b' }} />
            <Typography variant="caption" color="text.secondary">
              {formatNum(skill.stars)}
            </Typography>
          </Box>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
            <ForkRightIcon sx={{ fontSize: 14, color: 'text.secondary' }} />
            <Typography variant="caption" color="text.secondary">
              {formatNum(skill.forks)}
            </Typography>
          </Box>
          <Typography variant="caption" color="text.secondary">
            Updated {timeAgo}
          </Typography>
        </Stack>

        <Box>
          <Tooltip title={copied ? 'Copied!' : 'Copy URL'}>
            <IconButton size="small" onClick={handleCopy} sx={{ mr: 0.5 }}>
              <ContentCopyIcon sx={{ fontSize: 14 }} />
            </IconButton>
          </Tooltip>
          <Tooltip title="Open on GitHub">
            <IconButton
              size="small"
              component="a"
              href={skill.github_url}
              target="_blank"
              rel="noopener noreferrer"
              onClick={(e) => e.stopPropagation()}
            >
              <OpenInNewIcon sx={{ fontSize: 14 }} />
            </IconButton>
          </Tooltip>
        </Box>
      </CardActions>
    </Card>
  )
}

function getScoreColor(score: number): string {
  if (score >= 75) return '#10b981'  // green
  if (score >= 50) return '#6366f1'  // indigo
  if (score >= 25) return '#f59e0b'  // amber
  return '#ef4444'                   // red
}

function formatNum(n: number): string {
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k`
  return n.toString()
}

function formatTimeAgo(date: Date): string {
  const seconds = Math.floor((Date.now() - date.getTime()) / 1000)
  if (seconds < 60) return 'just now'
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days}d ago`
  const months = Math.floor(days / 30)
  if (months < 12) return `${months}mo ago`
  return `${Math.floor(months / 12)}y ago`
}
