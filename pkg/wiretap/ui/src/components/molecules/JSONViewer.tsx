import { JsonViewer } from '@textea/json-viewer'

import { darkJSON } from './JSONViewerTheme.ts'

export function JSONViewer({ value }: { value: unknown }) {
  return (
    <div className={'h-[250px] overflow-auto'}>
      <JsonViewer rootName={false} theme={darkJSON} value={value} />
    </div>
  )
}
