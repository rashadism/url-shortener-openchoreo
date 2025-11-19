import { useState, useEffect } from 'react'

// Use relative URLs - nginx will proxy to backend services
// Empty strings because endpoints already include /api and /analytics paths
const API_URL = ''
const ANALYTICS_URL = ''
const API_KEY = import.meta.env.VITE_API_KEY || 'test-api-key-12345'

function App() {
  const [activeTab, setActiveTab] = useState('create')
  const [longUrl, setLongUrl] = useState('')
  const [customCode, setCustomCode] = useState('')
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState(null)
  const [urls, setUrls] = useState([])
  const [stats, setStats] = useState(null)
  const [topUrls, setTopUrls] = useState([])

  useEffect(() => {
    loadUrls()
    loadAnalytics()
  }, [])

  const loadUrls = async () => {
    try {
      const response = await fetch(`${API_URL}/api/urls?api_key=${API_KEY}`)
      if (response.ok) {
        const data = await response.json()
        setUrls(data || [])
      }
    } catch (error) {
      console.error('Failed to load URLs:', error)
    }
  }

  const loadAnalytics = async () => {
    try {
      const [summaryRes, topUrlsRes] = await Promise.all([
        fetch(`${ANALYTICS_URL}/api/analytics/summary?api_key=${API_KEY}`),
        fetch(`${ANALYTICS_URL}/api/analytics/top-urls?api_key=${API_KEY}&limit=10`)
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

  const createShortUrl = async (e) => {
    e.preventDefault()
    setLoading(true)
    setMessage(null)

    try {
      const response = await fetch(`${API_URL}/api/urls`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          long_url: longUrl,
          custom_code: customCode || undefined,
          api_key: API_KEY,
        }),
      })

      const data = await response.json()

      if (response.ok) {
        // Construct full URL from relative path for display
        const fullUrl = `${window.location.origin}${data.short_url}`
        setMessage({ type: 'success', text: `Short URL created: ${fullUrl}` })
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
      <header className="header">
        <h1>⚡ URL Shortener</h1>
        <p className="subtitle">Fast, simple, and powerful link shortening</p>
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
                      type="url"
                      placeholder="https://example.com/very-long-url"
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
