import type { ICellRendererParams } from 'ag-grid-community'
import { cx } from 'class-variance-authority'

import type { Snapshot } from '../../dtos/Snapshot.ts'
import { SerializeURL } from '../../dtos/Snapshot.ts'

export interface StatusBubbleProps {
  statusCode: number
}

export function selectStatusBubbleColor({ statusCode }: StatusBubbleProps) {
  if (statusCode >= 200 && statusCode < 300) {
    return 'bg-green-500'
  } else if (statusCode >= 300 && statusCode < 400) {
    return 'bg-yellow-500'
  } else {
    return 'bg-red-500'
  }
}

export function renderStatusIndicatorCell(
  params: ICellRendererParams<Snapshot>,
) {
  return (
    <div className={'flex h-6 items-center justify-center'}>
      <div
        className={cx([
          'h-2 w-2 rounded-full opacity-70',
          selectStatusBubbleColor({
            statusCode: params.data?.Response?.StatusCode ?? 0,
          }),
        ])}
      />
    </div>
  )
}

export function createByteCellRenderer(field: 'Request' | 'Response') {
  return (params: ICellRendererParams<Snapshot>) => {
    if (!params.data) return null
    const value = params.data[field]?.ContentLength ?? 0
    return hunanReadableByteCount(value)
  }
}

export function hunanReadableByteCount(bytes: number) {
  const thresh = 1024
  if (Math.abs(bytes) < thresh) {
    return `${bytes} B`
  }
  const units = ['kB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
  let u = -1
  do {
    bytes /= thresh
    ++u
  } while (Math.abs(bytes) >= thresh && u < units.length - 1)
  return `${bytes.toFixed(1)} ${units[u]}`
}

export function renderURLCell(params: ICellRendererParams<Snapshot>) {
  if (!params.data) return null
  return SerializeURL(params.data?.Request.URL).toString()
}

export function renderDurationCell(params: ICellRendererParams<Snapshot>) {
  if (!params.data) return null
  return humanizeDuration(+params.data.Duration)
}

export function humanizeDuration(durationNano: number): string {
  // Constants for time conversions
  const NS_PER_MILLISECOND = 1e6
  const NS_PER_SECOND = 1e9
  const NS_PER_MINUTE = NS_PER_SECOND * 60
  const NS_PER_HOUR = NS_PER_MINUTE * 60
  const NS_PER_DAY = NS_PER_HOUR * 24

  if (durationNano >= NS_PER_DAY) {
    return `${(durationNano / NS_PER_DAY).toFixed(1)}d`
  } else if (durationNano >= NS_PER_HOUR) {
    return `${(durationNano / NS_PER_HOUR).toFixed(1)}hr`
  } else if (durationNano >= NS_PER_MINUTE) {
    return `${(durationNano / NS_PER_MINUTE).toFixed(1)}m`
  } else if (durationNano >= NS_PER_SECOND) {
    return `${(durationNano / NS_PER_SECOND).toFixed(1)}s`
  } else if (durationNano >= NS_PER_MILLISECOND) {
    return `${(durationNano / NS_PER_MILLISECOND).toFixed(1)}ms`
  } else {
    return `${durationNano}ns`
  }
}
