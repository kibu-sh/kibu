import { PauseIcon, PlayIcon, TrashIcon } from '@radix-ui/react-icons'
import { useCallback } from 'react'

import { BoxWithBorder } from '../atoms/DecorativeBox.tsx'
import { useSnapshotStreamContext } from '../contexts/SnapshotStreamContext.tsx'
import { useTableControlsContext } from '../contexts/TableControlsContext.tsx'

export interface ControlProps {
  className?: string
}

export function Control(props: ControlProps) {
  const { paused, pause, play } = useSnapshotStreamContext()
  const { gridApi } = useTableControlsContext()

  const clear = useCallback(() => {
    gridApi?.setRowData([])
  }, [gridApi])

  return (
    <div className={props.className}>
      <BoxWithBorder className={'h-full'}>
        <div className={'flex h-full items-center'}>
          <div
            className={'ml-5 cursor-pointer'}
            onClick={() => {
              paused ? play() : pause()
            }}
          >
            {paused ? <PlayIcon /> : <PauseIcon />}
          </div>

          <div className={'ml-5 cursor-pointer'} onClick={clear}>
            <TrashIcon />
          </div>
        </div>
      </BoxWithBorder>
    </div>
  )
}
