import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { apiClient, GetPasteResponse } from '../api/client'
import './View.css'

function View() {
  const { id } = useParams<{ id: string }>()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [paste, setPaste] = useState<GetPasteResponse | null>(null)
  const [viewMode, setViewMode] = useState<'translation' | 'original' | 'side-by-side'>('translation')
  const [selectedLanguage, setSelectedLanguage] = useState('')
  const [translating, setTranslating] = useState(false)

  useEffect(() => {
    const fetchPaste = async () => {
      if (!id) return
      
      try {
        const data = await apiClient.getPaste(id)
        setPaste(data)
        
        // Set initial language from browser or default to original
        const browserLang = navigator.language.split('-')[0]
        const availableLang = data.available_translations.includes(browserLang) 
          ? browserLang 
          : data.original_language
        setSelectedLanguage(availableLang)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load paste')
      } finally {
        setLoading(false)
      }
    }
    
    fetchPaste()
  }, [id])

  const handleLanguageChange = async (lang: string) => {
    if (!id || !paste) return
    
    setSelectedLanguage(lang)
    
    // If we don't have this translation yet, fetch it
    if (!paste.translations[lang]) {
      setTranslating(true)
      try {
        const response = await apiClient.translate(id, lang)
        setPaste({
          ...paste,
          translations: {
            ...paste.translations,
            [lang]: response.translation,
          },
          available_translations: [...paste.available_translations, lang],
        })
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Translation failed')
      } finally {
        setTranslating(false)
      }
    }
  }

  if (loading) {
    return <div className="view-container"><p>Loading...</p></div>
  }

  if (error) {
    return <div className="view-container"><p className="error-message">{error}</p></div>
  }

  if (!paste) {
    return <div className="view-container"><p>Paste not found</p></div>
  }

  const currentTranslation = paste.translations[selectedLanguage] || paste.original

  return (
    <div className="view-container">
      <div className="view-header">
        {selectedLanguage !== paste.original_language && (
          <div className="translation-notice">
            ⚠️ This is a machine-generated translation that may be inaccurate
          </div>
        )}
        
        <div className="view-controls">
          <div className="language-selector">
            <label htmlFor="language">Language:</label>
            <select
              id="language"
              value={selectedLanguage}
              onChange={(e) => handleLanguageChange(e.target.value)}
              disabled={translating}
            >
              <option value="en">English</option>
              <option value="es">Spanish</option>
              <option value="fr">French</option>
              <option value="de">German</option>
              <option value="ja">Japanese</option>
              <option value="zh">Chinese</option>
              <option value="pt">Portuguese</option>
              <option value="ru">Russian</option>
              <option value="ko">Korean</option>
              <option value="it">Italian</option>
            </select>
            {translating && <span className="translating">Translating...</span>}
          </div>

          <div className="view-mode-tabs">
            <button
              className={viewMode === 'translation' ? 'active' : ''}
              onClick={() => setViewMode('translation')}
            >
              {selectedLanguage === paste.original_language ? 'Original' : 'Translation'}
            </button>
            <button
              className={viewMode === 'original' ? 'active' : ''}
              onClick={() => setViewMode('original')}
              disabled={selectedLanguage === paste.original_language}
            >
              Original
            </button>
            <button
              className={viewMode === 'side-by-side' ? 'active' : ''}
              onClick={() => setViewMode('side-by-side')}
              disabled={selectedLanguage === paste.original_language}
            >
              Side-by-side
            </button>
          </div>
        </div>
      </div>

      <div className="paste-content">
        {viewMode === 'translation' && (
          <div className="text-box">
            {currentTranslation}
          </div>
        )}
        
        {viewMode === 'original' && (
          <div className="text-box">
            {paste.original}
          </div>
        )}
        
        {viewMode === 'side-by-side' && (
          <div className="side-by-side">
            <div className="text-box">
              <h3>Original ({paste.original_language})</h3>
              {paste.original}
            </div>
            <div className="text-box">
              <h3>Translation ({selectedLanguage})</h3>
              {currentTranslation}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default View
