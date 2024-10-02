import {
  createSnapshotStreamURL,
  currentWindowURL,
  useWebsocketStream,
} from './Socket.ts'

export type HTTPMessage = {
  Body: string
  ContentType: string
  ContentLength: number
  Header: Record<string, string[]>
  Raw: string
}

export type SerializableURL = {
  Scheme: 'https' | 'http'
  Opaque: string
  User: string | null
  Host: string
  Path: string
  RawQuery: ''
}

export type SerializableRequest = HTTPMessage & {
  URL: SerializableURL
  Method: string
}

export type SerializableResponse = HTTPMessage & {
  Status: string
  StatusCode: number
}

export type Snapshot = {
  ID: string
  Duration: string
  Secure: boolean
  Error: string
  Request: SerializableRequest
  Response?: SerializableResponse
}

export function isSerializableRequest(
  v: HTTPMessage | null | undefined,
): v is SerializableRequest {
  if (!v) return false
  return 'URL' in v && 'Method' in v
}

export function isSerializableResponse(
  v: HTTPMessage | null | undefined,
): v is SerializableResponse {
  if (!v) return false
  return 'Status' in v && 'StatusCode' in v
}

export function SerializeURL(s: SerializableURL): URL {
  return new URL(`${s.Scheme}://${s.Host}${s.Path}?${s.RawQuery}`)
}

export function useSnapshotStream() {
  return useWebsocketStream<Snapshot>(
    createSnapshotStreamURL(currentWindowURL()),
  )
}
