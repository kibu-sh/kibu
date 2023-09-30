import { SnapshotStreamProvider } from './components/contexts/SnapshotStreamContext.tsx'
import { SelectedSnapshotProvider } from './components/contexts/TableControlsContext.tsx'
import { Control } from './components/molecules/Control.tsx'
import { SelectedSnapshotRequestURL } from './components/molecules/SelectedNetworkCaptureLog.tsx'
import { HTTPInspector } from './components/organisms/HTTPInspector.tsx'
import { NetworkCaptureLogTable } from './components/organisms/NetworkCaptureLogTable.tsx'

function App() {
  return (
    <SnapshotStreamProvider>
      <SelectedSnapshotProvider>
        <div className={'flex h-full flex-col gap-4 p-4'}>
          <Control className={'h-10'} />
          <NetworkCaptureLogTable className={'h-48'} />
          <SelectedSnapshotRequestURL />

          <div className={'flex h-20 flex-grow flex-row gap-4'}>
            <HTTPInspector className={'w-1/2'} mode={'Request'} />
            <HTTPInspector className={'w-1/2'} mode={'Response'} />
          </div>
        </div>
      </SelectedSnapshotProvider>
    </SnapshotStreamProvider>
  )
}

export default App
