import type { GridApi, RowSelectedEvent } from 'ag-grid-community'
import type { PropsWithChildren } from 'react'
import { createContext, useCallback, useContext, useState } from 'react'

import type { Snapshot } from '../../dtos/Snapshot.ts'

export const TableControlsContext = createContext<{
  selectedSnapshot: Snapshot | null
  setSelectedSnapshot: (snapshot: Snapshot | null) => void
  onSelectionChange: (e: RowSelectedEvent<Snapshot>) => void
  gridApi: GridApi<Snapshot> | null
  setGridApi: (gridApi: GridApi<Snapshot> | null) => void
}>({
  selectedSnapshot: null,
  setSelectedSnapshot: () => {},
  onSelectionChange: () => {},
  gridApi: null,
  setGridApi: () => {},
})

export const useTableControlsContext = () => useContext(TableControlsContext)

export function SelectedSnapshotProvider({ children }: PropsWithChildren) {
  const [gridApi, setGridApi] = useState<GridApi<Snapshot> | null>(null)
  const [selectedSnapshot, setSelectedSnapshot] = useState<Snapshot | null>(
    null,
  )

  const onSelectionChange = useCallback((e: RowSelectedEvent<Snapshot>) => {
    setSelectedSnapshot(e.data || null)
  }, [])

  return (
    <TableControlsContext.Provider
      value={{
        selectedSnapshot,
        setSelectedSnapshot,
        onSelectionChange,
        gridApi,
        setGridApi,
      }}
    >
      {children}
    </TableControlsContext.Provider>
  )
}
