import type { Snapshot } from './Snapshot.ts'

export type listenerFunc = (event: MessageEvent<Snapshot>) => any

export interface SnapshotEventSource {
  addEventListener(eventType: string, listener: listenerFunc): void
  removeEventListener(type: string, listener: listenerFunc): void
  close(): void
}
