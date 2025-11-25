import { useState, useEffect } from 'react'

// Use relative URLs - nginx will proxy to backend services
const API_URL = ''
const ANALYTICS_URL = ''

function App() {
  const [username, setUsername] = useState(localStorage.getItem('username') || '')
  const [showUsernameModal, setShowUsernameModal] = useState(!localStorage.getItem('username'))
  const [activeTab, setActiveTab] = useState('create')
  const [longUrl, setLongUrl] = useState('')
  const [customCode, setCustomCode] = useState('')
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState(null)
  const [urls, setUrls] = useState([])
  const [stats, setStats] = useState(null)
  const [topUrls, setTopUrls] = useState([])

  useEffect(() => {
    if (username) {
      loadUrls()
      loadAnalytics()
    }
  }, [username])

  const handleUsernameSubmit = (e) => {
    e.preventDefault()
    if (username.trim()) {
      localStorage.setItem('username', username.trim())
      setShowUsernameModal(false)
      loadUrls()
      loadAnalytics()
    }
  }

  const handleChangeUsername = () => {
    setShowUsernameModal(true)
  }

  const loadUrls = async () => {
    if (!username) return
    try {
      const response = await fetch(`${API_URL}/api/urls?username=${encodeURIComponent(username)}`)
      if (response.ok) {
        const data = await response.json()
        setUrls(data || [])
      }
    } catch (error) {
      console.error('Failed to load URLs:', error)
    }
  }

  const loadAnalytics = async () => {
    if (!username) return
    try {
      const [summaryRes, topUrlsRes] = await Promise.all([
        fetch(`${ANALYTICS_URL}/api/analytics/summary?username=${encodeURIComponent(username)}`),
        fetch(`${ANALYTICS_URL}/api/analytics/top-urls?username=${encodeURIComponent(username)}&limit=10`)
      ])

      if (summaryRes.ok) {
        const summaryData = await summaryRes.json()
        setStats(summaryData)
      }

      if (topUrlsRes.ok) {
        const topUrlsData = await topUrlsRes.json()
        setTopUrls(topUrlsData)
      }
    } catch (error) {
      console.error('Failed to load analytics:', error)
    }
  }

  const normalizeUrl = (url) => {
    // Trim whitespace
    url = url.trim()

    // If URL doesn't start with http:// or https://, add https://
    if (!url.match(/^https?:\/\//i)) {
      url = 'https://' + url
    }

    return url
  }

  const copyToClipboard = async (text) => {
    try {
      await navigator.clipboard.writeText(text)
      setMessage({
        type: 'success',
        text: 'Copied to clipboard!',
        url: text
      })
      setTimeout(() => {
        const fullUrl = `${window.location.origin}${text.replace(window.location.origin, '')}`
        setMessage({
          type: 'success',
          text: 'Short URL created:',
          url: fullUrl
        })
      }, 1000)
    } catch (error) {
      console.error('Failed to copy:', error)
    }
  }

  const createShortUrl = async (e) => {
    e.preventDefault()
    setLoading(true)
    setMessage(null)

    try {
      // Normalize the URL before sending
      const normalizedUrl = normalizeUrl(longUrl)

      const response = await fetch(`${API_URL}/api/urls`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          long_url: normalizedUrl,
          custom_code: customCode || undefined,
          username: username,
        }),
      })

      const data = await response.json()

      if (response.ok) {
        // Construct full URL from relative path for display
        const fullUrl = `${window.location.origin}${data.short_url}`
        setMessage({
          type: 'success',
          text: 'Short URL created:',
          url: fullUrl
        })
        setLongUrl('')
        setCustomCode('')
        loadUrls()
        setTimeout(loadAnalytics, 1000)
      } else {
        setMessage({ type: 'error', text: data.error || 'Failed to create URL' })
      }
    } catch (error) {
      setMessage({ type: 'error', text: 'Network error: ' + error.message })
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="app">
      {showUsernameModal && (
        <div className="modal-overlay">
          <div className="modal">
            <h2>Welcome!</h2>
            <p>Enter your username to get started</p>
            <form onSubmit={handleUsernameSubmit}>
              <input
                type="text"
                placeholder="Enter username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                autoFocus
                required
              />
              <button type="submit" className="primary-btn">
                Continue
              </button>
            </form>
          </div>
        </div>
      )}

      <header className="header">
        <h1>⚡ URL Shortener</h1>
        <p className="subtitle">Fast, simple, and powerful link shortening</p>
        {username && !showUsernameModal && (
          <button onClick={handleChangeUsername} className="username-btn">
            👤 {username}
          </button>
        )}
      </header>

      <div className="container">
        <div className="tabs">
          <button
            className={`tab ${activeTab === 'create' ? 'active' : ''}`}
            onClick={() => setActiveTab('create')}
          >
            ✨ Create URL
          </button>
          <button
            className={`tab ${activeTab === 'urls' ? 'active' : ''}`}
            onClick={() => setActiveTab('urls')}
          >
            🔗 My URLs ({urls.length})
          </button>
          <button
            className={`tab ${activeTab === 'analytics' ? 'active' : ''}`}
            onClick={() => setActiveTab('analytics')}
          >
            📊 Analytics
          </button>
          <button
            className={`tab ${activeTab === 'top' ? 'active' : ''}`}
            onClick={() => setActiveTab('top')}
          >
            🏆 Top URLs
          </button>
        </div>

        <div className="tab-content">
          {activeTab === 'create' && (
            <div className="create-tab">
              <div className="form-card">
                <h2>Create Short URL</h2>
                <form onSubmit={createShortUrl}>
                  <div className="input-group">
                    <label>Long URL</label>
                    <input
                      type="text"
                      placeholder="google.com or https://example.com/very-long-url"
                      value={longUrl}
                      onChange={(e) => setLongUrl(e.target.value)}
                      required
                    />
                  </div>
                  <div className="input-group">
                    <label>Custom Short Code (optional)</label>
                    <input
                      type="text"
                      placeholder="my-custom-code"
                      value={customCode}
                      onChange={(e) => setCustomCode(e.target.value)}
                    />
                  </div>
                  <button type="submit" className="primary-btn" disabled={loading}>
                    {loading ? '⏳ Creating...' : '🚀 Shorten URL'}
                  </button>
                </form>

                {message && (
                  <div className={`message ${message.type}`}>
                    {message.text}
                    {message.url && (
                      <div className="url-result">
                        <a
                          href={message.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="result-link"
                        >
                          {message.url}
                        </a>
                        <button
                          onClick={() => copyToClipboard(message.url)}
                          className="copy-btn"
                          title="Copy to clipboard"
                        >
                          📋 Copy
                        </button>
                      </div>
                    )}
                  </div>
                )}
              </div>
            </div>
          )}

          {activeTab === 'urls' && (
            <div className="urls-tab">
              <h2>Your URLs</h2>
              {urls.length === 0 ? (
                <div className="empty-state">
                  <p>No URLs created yet. Create your first short URL!</p>
                  <button onClick={() => setActiveTab('create')} className="secondary-btn">
                    Create URL
                  </button>
                </div>
              ) : (
                <div className="url-list">
                  {urls.map((url) => (
                    <div key={url.id} className="url-card">
                      <div className="url-header">
                        <span className="short-code">/{url.short_code}</span>
                        <span className="created-date">
                          {new Date(url.created_at).toLocaleDateString()}
                        </span>
                      </div>
                      {url.title && <div className="url-title">{url.title}</div>}
                      <div className="url-long">{url.long_url}</div>
                      <div className="url-actions">
                        <a
                          href={`/${url.short_code}`}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="link-btn"
                        >
                          🔗 Visit
                        </a>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {activeTab === 'analytics' && (
            <div className="analytics-tab">
              <h2>Analytics Overview</h2>
              {stats ? (
                <div className="stats-grid">
                  <div className="stat-card blue">
                    <div className="stat-icon">🔗</div>
                    <div className="stat-content">
                      <div className="stat-label">Total URLs</div>
                      <div className="stat-value">{stats.total_urls}</div>
                    </div>
                  </div>
                  <div className="stat-card green">
                    <div className="stat-icon">👆</div>
                    <div className="stat-content">
                      <div className="stat-label">Total Clicks</div>
                      <div className="stat-value">{stats.total_clicks}</div>
                    </div>
                  </div>
                  <div className="stat-card orange">
                    <div className="stat-icon">📅</div>
                    <div className="stat-content">
                      <div className="stat-label">Clicks Today</div>
                      <div className="stat-value">{stats.clicks_today}</div>
                    </div>
                  </div>
                  <div className="stat-card purple">
                    <div className="stat-icon">📈</div>
                    <div className="stat-content">
                      <div className="stat-label">Clicks This Week</div>
                      <div className="stat-value">{stats.clicks_this_week}</div>
                    </div>
                  </div>
                </div>
              ) : (
                <div className="loading">Loading analytics...</div>
              )}
            </div>
          )}

          {activeTab === 'top' && (
            <div className="top-tab">
              <h2>Top Performing URLs</h2>
              {topUrls.length === 0 ? (
                <div className="empty-state">
                  <p>No data available yet. Start creating URLs and getting clicks!</p>
                </div>
              ) : (
                <div className="top-urls-list">
                  {topUrls.map((url, idx) => (
                    <div key={url.url_id} className="top-url-card">
                      <div className="rank">#{idx + 1}</div>
                      <div className="top-url-content">
                        <div className="url-header">
                          <span className="short-code">/{url.short_code}</span>
                          <span className="clicks-badge">{url.total_clicks} clicks</span>
                        </div>
                        {url.title && <div className="url-title">{url.title}</div>}
                        <div className="url-long">{url.long_url}</div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default App
