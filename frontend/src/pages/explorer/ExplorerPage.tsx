import { useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import {
  Box, Grid, Typography, TextField, InputAdornment,
  Button, Chip, CircularProgress, Pagination, Alert,
  Tabs, Tab, Divider, Container, Tooltip, Switch,
  FormControlLabel
} from '@mui/material'
import SearchIcon from '@mui/icons-material/Search'
import TrendingUpIcon from '@mui/icons-material/TrendingUp'
import AddIcon from '@mui/icons-material/Add'
import StarIcon from '@mui/icons-material/Star'
import { skillsService } from '@/services/skills.service'
import { SkillCard } from '@/components/skills/SkillCard'
import { SubmitSkillDialog } from './SubmitSkillDialog'

const POPULAR_TAGS = ['devops', 'testing', 'documentation', 'api', 'database', 'security', 'frontend', 'python', 'golang']
const MIN_QUALITY_STARS = 100

export function ExplorerPage() {
  const navigate = useNavigate()
  const [query, setQuery] = useState('')
  const [searchInput, setSearchInput] = useState('')
  const [selectedTags, setSelectedTags] = useState<string[]>([])
  const [page, setPage] = useState(1)
  const [tab, setTab] = useState(0) // 0 = Search, 1 = Trending
  const [submitOpen, setSubmitOpen] = useState(false)
  const [qualityFilter, setQualityFilter] = useState(true) // default ON: 100+ stars

  const limit = 12
  const minStars = qualityFilter ? MIN_QUALITY_STARS : 0

  // Search query
  const { data: searchData, isLoading: searchLoading, error: searchError } = useQuery({
    queryKey: ['skills', 'search', query, selectedTags, page, minStars],
    queryFn: () => skillsService.search(query, selectedTags, limit, (page - 1) * limit, minStars),
    enabled: tab === 0,
    staleTime: 60 * 1000,
  })

  // Trending query
  const { data: trendingData, isLoading: trendingLoading } = useQuery({
    queryKey: ['skills', 'trending', minStars],
    queryFn: () => skillsService.getTrending(20, minStars),
    enabled: tab === 1,
    staleTime: 5 * 60 * 1000,
  })

  const handleSearch = useCallback(() => {
    setQuery(searchInput)
    setPage(1)
  }, [searchInput])

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') handleSearch()
  }

  const toggleTag = (tag: string) => {
    setSelectedTags((prev) =>
      prev.includes(tag) ? prev.filter((t) => t !== tag) : [...prev, tag]
    )
    setPage(1)
  }

  const skills = tab === 0 ? searchData?.skills : trendingData?.skills
  const total = searchData?.total ?? 0
  const isLoading = tab === 0 ? searchLoading : trendingLoading

  return (
    <Container maxWidth="xl" sx={{ py: 3 }}>
      {/* Header */}
      <Box sx={{ mb: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 1 }}>
          <Box>
            <Typography variant="h4" fontWeight={700} gutterBottom>
              Skills Explorer
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Discover ranked SKILL.md files from across GitHub · Powered by MCP
            </Typography>
          </Box>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => setSubmitOpen(true)}
            sx={{ height: 40 }}
          >
            Submit Skill
          </Button>
        </Box>
      </Box>

      {/* Search bar */}
      <Box sx={{ display: 'flex', gap: 1, mb: 2 }}>
        <TextField
          fullWidth
          placeholder="Search skills… e.g. 'docker deployment', 'code review', 'database migration'"
          value={searchInput}
          onChange={(e) => setSearchInput(e.target.value)}
          onKeyDown={handleKeyDown}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <SearchIcon color="action" />
              </InputAdornment>
            ),
          }}
          sx={{ bgcolor: 'background.paper', borderRadius: 2 }}
        />
        <Button variant="contained" onClick={handleSearch} sx={{ px: 3, flexShrink: 0 }}>
          Search
        </Button>
      </Box>

      {/* Tag filters + Quality filter */}
      <Box sx={{ display: 'flex', flexWrap: 'wrap', alignItems: 'center', gap: 1, mb: 2 }}>
        {POPULAR_TAGS.map((tag) => (
          <Chip
            key={tag}
            label={tag}
            clickable
            size="small"
            variant={selectedTags.includes(tag) ? 'filled' : 'outlined'}
            color={selectedTags.includes(tag) ? 'primary' : 'default'}
            onClick={() => toggleTag(tag)}
          />
        ))}

        <Box sx={{ ml: 'auto', display: 'flex', alignItems: 'center', gap: 0.5 }}>
          <Tooltip
            title={qualityFilter
              ? `Showing skills from repos with ≥${MIN_QUALITY_STARS} ⭐. Toggle off to see all.`
              : 'Quality filter OFF — showing all indexed skills'}
            placement="top"
          >
            <FormControlLabel
              control={
                <Switch
                  size="small"
                  checked={qualityFilter}
                  onChange={(e) => { setQualityFilter(e.target.checked); setPage(1) }}
                  color="warning"
                />
              }
              label={
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                  <StarIcon sx={{ fontSize: 14, color: qualityFilter ? '#f59e0b' : 'text.disabled' }} />
                  <Typography variant="caption" color={qualityFilter ? 'warning.main' : 'text.secondary'} fontWeight={600}>
                    {qualityFilter ? `${MIN_QUALITY_STARS}+ stars` : 'All skills'}
                  </Typography>
                </Box>
              }
              sx={{ m: 0 }}
            />
          </Tooltip>
        </Box>
      </Box>

      <Divider sx={{ mb: 2 }} />

      {/* Tabs */}
      <Tabs value={tab} onChange={(_, v) => setTab(v)} sx={{ mb: 2 }}>
        <Tab
          icon={<SearchIcon sx={{ fontSize: 16 }} />}
          iconPosition="start"
          label="Search Results"
        />
        <Tab
          icon={<TrendingUpIcon sx={{ fontSize: 16 }} />}
          iconPosition="start"
          label="Trending"
        />
      </Tabs>

      {/* Results */}
      {searchError && (
        <Alert severity="error" sx={{ mb: 2 }}>
          Failed to load skills. Make sure your API key is valid.
        </Alert>
      )}

      {isLoading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}>
          <CircularProgress />
        </Box>
      ) : (
        <>
          {tab === 0 && (
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
              <Typography variant="caption" color="text.secondary">
                {total > 0
                  ? `${total} skill${total === 1 ? '' : 's'} found${qualityFilter ? ` · ≥${MIN_QUALITY_STARS} ⭐ quality filter active` : ''}`
                  : qualityFilter
                    ? `No high-quality skills found — try turning off the quality filter`
                    : 'No skills found — try different keywords'}
              </Typography>
            </Box>
          )}

          <Grid container spacing={2}>
            {skills?.map((skill, i) => (
              <Grid item xs={12} sm={6} md={4} lg={3} key={skill.id}>
                <SkillCard
                  skill={skill}
                  rank={tab === 1 ? i + 1 : undefined}
                  onClick={() => navigate(`/skills/${skill.id}`)}
                />
              </Grid>
            ))}
          </Grid>

          {tab === 0 && total > limit && (
            <Box sx={{ display: 'flex', justifyContent: 'center', mt: 4 }}>
              <Pagination
                count={Math.ceil(total / limit)}
                page={page}
                onChange={(_, p) => setPage(p)}
                color="primary"
              />
            </Box>
          )}
        </>
      )}

      <SubmitSkillDialog open={submitOpen} onClose={() => setSubmitOpen(false)} />
    </Container>
  )
}
