import { Box, Tooltip, Typography } from '@mui/material'

interface ScoreBadgeProps {
  score: number
  size?: 'small' | 'medium'
}

export function ScoreBadge({ score, size = 'medium' }: ScoreBadgeProps) {
  const tier = getScoreTier(score)
  const isSmall = size === 'small'

  return (
    <Tooltip
      title={`Quality Score: ${score.toFixed(1)}/100 (${tier.label})`}
      placement="top"
    >
      <Box
        sx={{
          px: isSmall ? 0.75 : 1,
          py: isSmall ? 0.25 : 0.5,
          borderRadius: 1.5,
          bgcolor: tier.bgColor,
          border: '1px solid',
          borderColor: tier.borderColor,
          flexShrink: 0,
          display: 'flex',
          alignItems: 'center',
          gap: 0.5,
          cursor: 'default',
        }}
      >
        <Box
          component="span"
          sx={{ fontSize: isSmall ? '0.65rem' : '0.7rem' }}
        >
          {tier.emoji}
        </Box>
        <Typography
          variant="caption"
          fontWeight={700}
          sx={{ color: tier.textColor, fontSize: isSmall ? '0.65rem' : '0.75rem' }}
        >
          {score.toFixed(1)}
        </Typography>
      </Box>
    </Tooltip>
  )
}

function getScoreTier(score: number) {
  if (score >= 80) return {
    label: 'Elite',
    emoji: '🏆',
    bgColor: 'rgba(16, 185, 129, 0.12)',
    borderColor: 'rgba(16, 185, 129, 0.4)',
    textColor: '#10b981',
  }
  if (score >= 60) return {
    label: 'Great',
    emoji: '⭐',
    bgColor: 'rgba(99, 102, 241, 0.12)',
    borderColor: 'rgba(99, 102, 241, 0.4)',
    textColor: '#818cf8',
  }
  if (score >= 40) return {
    label: 'Good',
    emoji: '✅',
    bgColor: 'rgba(6, 182, 212, 0.12)',
    borderColor: 'rgba(6, 182, 212, 0.4)',
    textColor: '#06b6d4',
  }
  if (score >= 20) return {
    label: 'Fair',
    emoji: '📄',
    bgColor: 'rgba(245, 158, 11, 0.12)',
    borderColor: 'rgba(245, 158, 11, 0.4)',
    textColor: '#f59e0b',
  }
  return {
    label: 'New',
    emoji: '🆕',
    bgColor: 'rgba(148, 163, 184, 0.08)',
    borderColor: 'rgba(148, 163, 184, 0.2)',
    textColor: '#94a3b8',
  }
}
