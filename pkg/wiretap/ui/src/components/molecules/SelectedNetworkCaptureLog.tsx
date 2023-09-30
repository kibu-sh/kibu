import { cx } from 'class-variance-authority'

import { BoxWithBorder } from '../atoms/DecorativeBox.tsx'
import { useTableControlsContext } from '../contexts/TableControlsContext.tsx'

export interface SelectedCaptureLogProps {
  className?: string
}

function selectBadgeCodeColor(statusCode?: number) {
  if (statusCode === undefined || statusCode === null) {
    return 'bg-gray-500'
  } else if (statusCode >= 200 && statusCode < 300) {
    return 'bg-green-500'
  } else if (statusCode >= 300 && statusCode < 400) {
    return 'bg-yellow-500'
  } else {
    return 'bg-red-500'
  }
}

export function SelectedSnapshotRequestURL({
  className,
}: SelectedCaptureLogProps) {
  const { selectedSnapshot } = useTableControlsContext()
  const url = selectedSnapshot?.Request.URL
  return (
    <div className={className}>
      <BoxWithBorder className={'h-full'}>
        <div className={'row flex h-full p-3'}>
          <div
            className={cx([
              'rounded text-center text-white',
              'mr-2 w-20 px-2 py-1 text-xs leading-none',
              'bg-gray-700',
            ])}
          >
            {selectedSnapshot?.Request.Method}
          </div>

          <div
            className={cx([
              'mr-2 w-10 rounded px-2 py-1 text-center text-xs leading-none text-white',
              selectBadgeCodeColor(selectedSnapshot?.Response?.StatusCode),
            ])}
          >
            {selectedSnapshot?.Response?.StatusCode}
          </div>

          <div className={'mr-2 px-2 py-0.5 text-xs text-gray-400'}>
            <span>{url?.Scheme}://</span>
            <span className={'text-blue-500'}>{url?.Host}</span>
            <span className={'text-green-600'}>{url?.Path}</span>
            <span>{url?.RawQuery}</span>
          </div>
        </div>
      </BoxWithBorder>
    </div>
  )
}
