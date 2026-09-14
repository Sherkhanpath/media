import React, { useState } from 'react'
import { addMedia } from '../api.js'

export default function ManagePanel({ window: win, onClose, onSync, onMediaAdded }) {
  const [type, setType] = useState('image')
  const [url, setUrl] = useState('')
  const [duration, setDuration] = useState(8)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState(null)

  async function handleAdd(e) {
    e.preventDefault()
    setError(null)
    if (type !== 'blank' && !url.trim()) {
      setError('URL is required for image/video items.')
      return
    }
    setSubmitting(true)
    try {
      const updated = await addMedia(win.id, {
        type,
        url: type === 'blank' ? '' : url.trim(),
        durationSeconds: Number(duration) || 8,
      })
      onMediaAdded(updated)
      setUrl('')
    } catch (err) {
      setError(err.message)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="panel-overlay" onClick={onClose}>
      <div className="panel" onClick={(e) => e.stopPropagation()}>
        <div className="panel-head">
          <h2>{win.name}</h2>
          <button className="close-btn" onClick={onClose}>&times;</button>
        </div>

        <h3 className="panel-subhead">Playlist</h3>
        <ul className="playlist-list">
          {win.playlist.map((item) => (
            <li key={item.id} className="playlist-row">
              <span className={`type-pill ${item.type}`}>{item.type}</span>
              <span className="playlist-url" title={item.url}>
                {item.type === 'blank' ? '(blank)' : item.url}
              </span>
              <span className="playlist-duration">{item.durationSeconds}s</span>
              <button className="sync-btn" onClick={() => onSync(item)}>Sync</button>
            </li>
          ))}
          {win.playlist.length === 0 && <li className="playlist-empty">No media yet.</li>}
        </ul>

        <h3 className="panel-subhead">Add media</h3>
        <form className="add-form" onSubmit={handleAdd}>
          <div className="add-form-row">
            <select value={type} onChange={(e) => setType(e.target.value)}>
              <option value="image">Image</option>
              <option value="video">Video</option>
              <option value="blank">Blank</option>
            </select>
            {type !== 'blank' && (
              <input
                type="url"
                placeholder="https://…"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
              />
            )}
            <input
              type="number"
              min="1"
              value={duration}
              onChange={(e) => setDuration(e.target.value)}
              title="Duration (seconds)"
            />
          </div>
          <button className="primary-btn" type="submit" disabled={submitting}>
            {submitting ? 'Adding…' : 'Add to playlist'}
          </button>
          {error && <div className="form-error">{error}</div>}
        </form>
      </div>
    </div>
  )
}
