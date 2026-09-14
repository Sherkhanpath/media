import React, { useEffect, useRef, useState, useCallback } from 'react'
import { getWindows, connectSocket, triggerSync } from './api.js'
import WindowCard from './components/WindowCard.jsx'
import ManagePanel from './components/ManagePanel.jsx'

export default function App() {
  const [windows, setWindows] = useState(null) // null = loading
  const [error, setError] = useState(null)
  const [activeSync, setActiveSync] = useState(null) // { mediaItem, durationSeconds } | null
  const [managingId, setManagingId] = useState(null) // window id currently open in the manage panel
  const syncTimeoutRef = useRef(null)

  const refreshWindows = useCallback(() => {
    getWindows()
      .then(setWindows)
      .catch((e) => setError(e.message))
  }, [])

  useEffect(() => {
    refreshWindows()
  }, [refreshWindows])

  // Listen for real-time events from the backend: a sync trigger, or a
  // playlist change made by anyone (including from another browser tab).
  useEffect(() => {
    const disconnect = connectSocket((event) => {
      if (event.type === 'sync') {
        const { mediaItem, durationSeconds } = event.payload
        setActiveSync({ mediaItem, durationSeconds })
        if (syncTimeoutRef.current) clearTimeout(syncTimeoutRef.current)
        syncTimeoutRef.current = setTimeout(() => setActiveSync(null), durationSeconds * 1000)
      } else if (event.type === 'playlist_updated') {
        const updated = event.payload.window
        setWindows((prev) =>
          prev ? prev.map((w) => (w.id === updated.id ? updated : w)) : prev
        )
      }
    })
    return disconnect
  }, [])

  async function handleSyncClick(mediaItem) {
    try {
      await triggerSync(mediaItem, 8)
    } catch (e) {
      setError(e.message)
    }
  }

  if (error) {
    return <div className="state-message error">Couldn't load: {error}</div>
  }
  if (!windows) {
    return <div className="state-message">Loading windows…</div>
  }

  const managingWindow = windows.find((w) => w.id === managingId) || null

  return (
    <div className="app">
      <header className="app-header">
        <div className="brand">
          <span className="brand-dot" />
          Media Sequencer
        </div>
        {activeSync && (
          <div className="sync-banner">
            SYNCING · {activeSync.mediaItem.type.toUpperCase()}
          </div>
        )}
      </header>

      <main className="window-grid">
        {windows.map((w) => (
          <WindowCard
            key={w.id}
            window={w}
            activeSync={activeSync}
            onManage={() => setManagingId(w.id)}
          />
        ))}
      </main>

      {managingWindow && (
        <ManagePanel
          window={managingWindow}
          onClose={() => setManagingId(null)}
          onSync={handleSyncClick}
          onMediaAdded={(updated) =>
            setWindows((prev) => prev.map((w) => (w.id === updated.id ? updated : w)))
          }
        />
      )}
    </div>
  )
}
