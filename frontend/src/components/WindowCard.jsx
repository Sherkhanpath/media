import React, { useEffect, useRef, useState } from 'react'

export default function WindowCard({ window: win, activeSync, onManage }) {
  const [currentIndex, setCurrentIndex] = useState(0)
  const advanceTimerRef = useRef(null)
  const advancedRef = useRef(false)

  const playlist = win.playlist || []
  const hasItems = playlist.length > 0
  // Clamp in case the playlist shrank or hasn't loaded an item at this index yet.
  const safeIndex = hasItems ? currentIndex % playlist.length : 0
  const current = hasItems ? playlist[safeIndex] : null

  function advance() {
    if (advancedRef.current) return // guard against double-advance (video onEnded + fallback timer both firing)
    advancedRef.current = true
    setCurrentIndex((i) => (hasItems ? (i + 1) % playlist.length : i))
  }

  // Drive normal playback: advance to the next playlist item after the
  // current one's duration elapses (images/blank), or when the video ends.
  // This effect intentionally does nothing while a sync is active — the
  // window just holds its place until the sync ends, then this effect
  // re-runs and resumes from the same index.
  useEffect(() => {
    if (activeSync || !current) return
    advancedRef.current = false

    if (current.type === 'video') {
      // Fallback in case the video fails to load/play: advance anyway
      // after durationSeconds so a broken link can't freeze the window.
      if (current.durationSeconds > 0) {
        advanceTimerRef.current = setTimeout(advance, current.durationSeconds * 1000)
      }
    } else {
      const seconds = current.durationSeconds > 0 ? current.durationSeconds : 8
      advanceTimerRef.current = setTimeout(advance, seconds * 1000)
    }

    return () => clearTimeout(advanceTimerRef.current)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [safeIndex, activeSync, current?.id])

  const displayed = activeSync ? activeSync.mediaItem : current

  return (
    <div className="window-card">
      <div className="window-card-head">
        <span className="window-name">{win.name}</span>
        <button className="manage-btn" onClick={onManage}>Manage</button>
      </div>

      <div className="window-viewport">
        {activeSync && <div className="sync-flag">SYNC</div>}
        <MediaView
          item={displayed}
          onVideoEnded={activeSync ? undefined : advance}
        />
      </div>

      <div className="window-foot">
        {hasItems ? `${safeIndex + 1} / ${playlist.length}` : 'empty playlist'}
      </div>
    </div>
  )
}

function MediaView({ item, onVideoEnded }) {
  if (!item) {
    return <div className="media-blank">No media configured</div>
  }
  if (item.type === 'blank') {
    return <div className="media-blank" />
  }
  if (item.type === 'image') {
    return <img className="media-el" src={item.url} alt="" />
  }
  if (item.type === 'video') {
    return (
      <video
        className="media-el"
        src={item.url}
        autoPlay
        muted
        playsInline
        onEnded={onVideoEnded}
        // key forces the element to remount (and restart from 0:00) whenever
        // the source changes, including switching into/out of a sync.
        key={item.id + item.url}
      />
    )
  }
  return <div className="media-blank">Unsupported media</div>
}
