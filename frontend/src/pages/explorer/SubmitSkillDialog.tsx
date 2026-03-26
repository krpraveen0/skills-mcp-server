import { useState } from 'react'
import {
  Dialog, DialogTitle, DialogContent, DialogActions,
  TextField, Button, Typography, Alert, CircularProgress, Box
} from '@mui/material'
import GitHubIcon from '@mui/icons-material/GitHub'
import { skillsService } from '@/services/skills.service'

interface Props {
  open: boolean
  onClose: () => void
}

export function SubmitSkillDialog({ open, onClose }: Props) {
  const [url, setUrl] = useState('')
  const [notes, setNotes] = useState('')
  const [loading, setLoading] = useState(false)
  const [success, setSuccess] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async () => {
    if (!url.startsWith('https://github.com/')) {
      setError('Please enter a valid GitHub URL')
      return
    }
    setLoading(true)
    setError('')
    try {
      await skillsService.submit(url, notes)
      setSuccess(true)
      setUrl('')
      setNotes('')
    } catch {
      setError('Failed to submit. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  const handleClose = () => {
    setSuccess(false)
    setError('')
    onClose()
  }

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
        <GitHubIcon />
        Submit a SKILL.md
      </DialogTitle>
      <DialogContent>
        {success ? (
          <Alert severity="success" sx={{ mt: 1 }}>
            Your skill has been queued for indexing! It will appear in search results after the next crawl.
          </Alert>
        ) : (
          <Box sx={{ mt: 1 }}>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
              Submit a GitHub URL pointing to a SKILL.md file or repository. Our crawler will index and rank it.
            </Typography>
            {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
            <TextField
              fullWidth
              label="GitHub URL"
              placeholder="https://github.com/owner/repo/blob/main/SKILL.md"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              sx={{ mb: 2 }}
            />
            <TextField
              fullWidth
              label="Notes (optional)"
              placeholder="What makes this skill useful?"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              multiline
              rows={2}
            />
          </Box>
        )}
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button onClick={handleClose} color="inherit">
          {success ? 'Close' : 'Cancel'}
        </Button>
        {!success && (
          <Button
            variant="contained"
            onClick={handleSubmit}
            disabled={loading || !url}
            startIcon={loading ? <CircularProgress size={16} /> : undefined}
          >
            Submit
          </Button>
        )}
      </DialogActions>
    </Dialog>
  )
}
