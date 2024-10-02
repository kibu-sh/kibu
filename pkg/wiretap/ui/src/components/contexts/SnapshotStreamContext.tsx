import type { ReactNode } from 'react'
import { createContext, useContext } from 'react'
import { Observable } from 'rxjs'

import type { Snapshot } from '../../dtos/Snapshot.ts'
import { useSnapshotStream } from '../../dtos/Snapshot.ts'

interface SnapshotStreamContextValue
  extends ReturnType<typeof useSnapshotStream> {}

export const SnapshotStreamContext = createContext<SnapshotStreamContextValue>({
  stream: new Observable<Snapshot>(),
  paused: false,
  pause: () => {},
  play: () => {},
})

export const useSnapshotStreamContext = () => useContext(SnapshotStreamContext)

interface SnapshotStreamProviderProps {
  children: ReactNode
}

export function SnapshotStreamProvider({
  children,
}: SnapshotStreamProviderProps) {
  return (
    <SnapshotStreamContext.Provider value={useSnapshotStream()}>
      {children}
    </SnapshotStreamContext.Provider>
  )
}
