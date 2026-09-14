// Base URL of the Go backend. Set VITE_API_URL in your environment (or a
// .env file) when the backend is deployed somewhere other than
// http://localhost:8080.
export const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

// WebSocket URL is derived from API_URL (same host, ws/wss scheme, /ws path).
export function wsURL() {
  const url = new URL(API_URL)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.pathname = '/ws'
  return url.toString()
}

async function request(path, options = {}) {
  const res = await fetch(API_URL + path, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  const data = await res.json().catch(() => null)
  if (!res.ok) {
    throw new Error((data && data.error) || `Request failed (${res.status})`)
  }
  return data
}

export function getWindows() {
  return request('/windows')
}

export function addMedia(windowId, { type, url, durationSeconds }) {
  return request(`/windows/${encodeURIComponent(windowId)}/media`, {
    method: 'POST',
    body: JSON.stringify({ type, url, durationSeconds }),
  })
}

export function triggerSync(mediaItem, durationSeconds) {
  return request('/sync', {
    method: 'POST',
    body: JSON.stringify({ mediaItem, durationSeconds }),
  })
}

// connectSocket opens the WebSocket and calls onEvent(event) for every
// message the backend broadcasts. It auto-reconnects (with a short delay)
// if the connection drops, since this app is meant to stay open on a
// display for long periods.
export function connectSocket(onEvent) {
  let socket
  let closedByCaller = false

  function connect() {
    socket = new WebSocket(wsURL())
    socket.onmessage = (msg) => {
      try {
        onEvent(JSON.parse(msg.data))
      } catch (e) {
        console.error('bad ws message', e)
      }
    }
    socket.onclose = () => {
      if (!closedByCaller) setTimeout(connect, 2000)
    }
    socket.onerror = () => socket.close()
  }

  connect()

  return () => {
    closedByCaller = true
    if (socket) socket.close()
  }
}
