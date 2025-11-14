import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { apiClient } from '../api/client'
import './Home.css'

function Home() {
  const navigate = useNavigate()
  const [content, setContent] = useState('')
  const [tone, setTone] = useState('default')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const maxLength = 20000
  const remaining = maxLength - content.length

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!content.trim() || content.length > maxLength) return

    setIsSubmitting(true)
    setError(null)
    
    try {
      const response = await apiClient.createPaste({ content, tone })
      navigate(`/paste/${response.paste_id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create paste')
      setIsSubmitting(false)
    }
  }

  return (
    <div className="home">
      <div className="home-header">
        <h2>Create a new paste</h2>
        <p>Share text with instant translation to any language</p>
      </div>

      <form onSubmit={handleSubmit} className="paste-form">
        {error && <div className="error-message">{error}</div>}
        
        <div className="form-group">
          <label htmlFor="content">Your text</label>
          <textarea
            id="content"
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="Enter your text here..."
            className="content-input"
            maxLength={maxLength}
            rows={15}
          />
          <div className="char-counter">
            <span className={remaining < 100 ? 'warning' : ''}>
              {remaining.toLocaleString()} characters remaining
            </span>
          </div>
        </div>

        <div className="form-group">
          <label htmlFor="tone">Translation tone</label>
          <select
            id="tone"
            value={tone}
            onChange={(e) => setTone(e.target.value)}
            className="tone-select"
          >
            <option value="default">Default</option>
            <option value="professional">Professional</option>
            <option value="friendly">Friendly</option>
            <option value="brusque">Brusque</option>
          </select>
          <p className="help-text">Choose how the AI should translate your text</p>
        </div>

        <button
          type="submit"
          disabled={!content.trim() || content.length > maxLength || isSubmitting}
          className="submit-btn"
        >
          {isSubmitting ? 'Creating...' : 'Create Paste'}
        </button>
      </form>
    </div>
  )
}

export default Home
