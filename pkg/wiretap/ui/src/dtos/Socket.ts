import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Observable } from 'rxjs'

export function createSocket(url: URL) {
  url.protocol = url.protocol.replace('http', 'ws')
  return new WebSocket(url)
}

export function createSnapshotSocket(url: URL) {
  return createSocket(createSnapshotStreamURL(url))
}

export function createBaseAdminURL(url: URL): URL {
  url.pathname = '__admin/api'
  url.search = ''
  return url
}

export function currentWindowURL(): URL {
  return new URL(window.location.href)
}

export function createSnapshotStreamURL(url: URL): URL {
  url = createBaseAdminURL(url)
  url.pathname = joinURL(url.pathname, 'snapshot', 'stream')
  return url
}

export function joinURL(base: string, ...paths: string[]) {
  base = base.replace(/\/$/, '') // Remove trailing slash from base URL
  paths = paths.map((path) => path.replace(/^\//, '').replace(/\/$/, '')) // Remove leading and trailing slashes from each path
  return [base, ...paths].join('/')
}

export function useWebsocketStream<T>(wsURL: URL) {
  const urlRef = useRef(wsURL)
  const socketRef = useRef<WebSocket | null>(null)
  const [paused, setPaused] = useState(false)

  const stream = useMemo(
    () =>
      new Observable<T>((subscriber) => {
        const socket = socketRef.current
        if (!socket) return
        if (paused) return
        console.debug('subscribing with websocket', socket)

        const onMessage = (event: MessageEvent) => {
          const { data, error } = decodeEvent<T>(event)
          if (data) {
            subscriber.next(data)
          }
          if (error) {
            subscriber.error(error)
          }
        }

        const onError = (event: Event) => {
          subscriber.error(event)
        }

        socket.addEventListener('message', onMessage)
        socket.addEventListener('error', onError)
        return () => {
          socket.removeEventListener('message', onMessage)
          socket.removeEventListener('error', onError)
        }
      }),
    [socketRef, paused],
  )

  useEffect(() => {
    const socket = (socketRef.current = createSocket(urlRef.current))
    socket.addEventListener('open', (event: Event) => {
      console.debug('websocket opened', urlRef.current, event)
    })
    socket.addEventListener('close', (event: CloseEvent) => {
      console.debug('websocket closed', urlRef.current, event)
    })
    return () => socketRef.current?.close()
  }, [])

  return {
    stream,
    paused,
    play: useCallback(() => setPaused(false), []),
    pause: useCallback(() => setPaused(true), []),
  }
}

interface decodeResult<T> {
  data: T | null
  error: Error | null
}
function decodeEvent<T>(event: MessageEvent): decodeResult<T> {
  try {
    const data = JSON.parse(event.data) as T
    return { data, error: null }
  } catch (err) {
    return { data: null, error: err as Error }
  }
}
