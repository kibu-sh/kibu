import 'ag-grid-community/styles/ag-grid.css'
import './NetworkCaptureTableStyle.css'
// import 'ag-grid-community/styles/ag-theme-balham.css'
import type { GridApi, GridReadyEvent } from 'ag-grid-community'
import { AgGridReact } from 'ag-grid-react'
import { cx } from 'class-variance-authority'
import { useCallback, useEffect, useRef, useState } from 'react'

import { selectEntireRow } from '../../ag-grid/ag-grid-utils.tsx'
import type { Snapshot } from '../../dtos/Snapshot.ts'
import { useSnapshotStreamContext } from '../contexts/SnapshotStreamContext.tsx'
import { useTableControlsContext } from '../contexts/TableControlsContext.tsx'

import {
  createByteCellRenderer,
  renderDurationCell,
  renderStatusIndicatorCell,
  renderURLCell,
} from './NetworkCaptureLogTableCells.tsx'

export interface NetworkCaptureLogTableProps {
  className?: string
}

export function NetworkCaptureLogTable(props: NetworkCaptureLogTableProps) {
  const [data] = useState<Snapshot[]>([])
  const { stream } = useSnapshotStreamContext()
  const { onSelectionChange, setGridApi } = useTableControlsContext()
  const gridApiRef = useRef<GridApi<Snapshot> | null>(null)

  const onGridReady = useCallback(
    (event: GridReadyEvent) => {
      gridApiRef.current = event.api
      setGridApi(event.api)
    },
    [setGridApi],
  )

  useEffect(() => {
    const subscription = stream.subscribe({
      next: (snapshot) => {
        console.debug('snapshot', snapshot)
        gridApiRef.current?.applyTransaction({ add: [snapshot] })
      },
    })
    return () => {
      subscription.unsubscribe()
    }
  }, [stream, gridApiRef.current])

  return (
    <div className={cx(['ag-theme-kibu-dark', props.className])}>
      <AgGridReact
        columnDefs={[
          {
            cellRenderer: renderStatusIndicatorCell,
            width: 50,
          },
          {
            headerName: 'URL',
            field: 'Request.URL',
            flex: 1,
            filter: true,
            cellRenderer: renderURLCell,
          },
          { headerName: 'Method', field: 'Request.Method', width: 100 },
          { headerName: 'Status', field: 'Response.Status', width: 100 },
          { headerName: 'Code', field: 'Response.StatusCode', width: 100 },
          {
            headerName: 'Duration',
            field: 'Duration',
            width: 100,
            cellRenderer: renderDurationCell,
          },
          {
            headerName: 'Req/size',
            field: 'Request.ContentLength',
            width: 100,
            cellRenderer: createByteCellRenderer('Request'),
          },
          {
            headerName: 'Res/size',
            field: 'Response.ContentLength',
            width: 100,
            cellRenderer: createByteCellRenderer('Response'),
          },
          { headerName: 'Secure', field: 'Secure', width: 75 },
        ]}
        enableCellTextSelection={true}
        enableRangeSelection={false}
        navigateToNextCell={selectEntireRow}
        rowData={data}
        rowSelection={'single'}
        onGridReady={onGridReady}
        onRowSelected={onSelectionChange}
      />
    </div>
  )
}
