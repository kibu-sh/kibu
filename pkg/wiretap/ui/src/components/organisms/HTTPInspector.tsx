import { Tabs } from '@radix-ui/themes'
import type { TabsContentProps } from '@radix-ui/themes/dist/cjs/components/tabs'
import { cx } from 'class-variance-authority'

import { isSerializableRequest } from '../../dtos/Snapshot.ts'
import { BoxWithBorder } from '../atoms/DecorativeBox.tsx'
import { useTableControlsContext } from '../contexts/TableControlsContext.tsx'
import { SyntaxHighlighter } from '../molecules/SyntaxHighlighter.tsx'

export type HTTPInspectorProps = {
  mode: 'Request' | 'Response'
  className?: string
}

export function HTTPInspector({ mode, className }: HTTPInspectorProps) {
  const isRequestMode = mode === 'Request'
  const { selectedSnapshot } = useTableControlsContext()
  const message = selectedSnapshot?.[mode]

  return (
    <BoxWithBorder className={className}>
      <Tabs.Root className={'h-full'} defaultValue="body">
        <Tabs.List>
          <Tabs.Trigger value="body">Body</Tabs.Trigger>
          <Tabs.Trigger value="header">Header</Tabs.Trigger>
          {isRequestMode && <Tabs.Trigger value="query">Query</Tabs.Trigger>}
          <Tabs.Trigger value="raw">Raw</Tabs.Trigger>
        </Tabs.List>

        <TabContentWithOverflow value="body">
          <SyntaxHighlighter
            code={message?.Body || ''}
            language={selectLanguage(message?.ContentType)}
          />
        </TabContentWithOverflow>

        <TabContentWithOverflow value="header">
          <SyntaxHighlighter
            code={jsonStringifyOrNull(message?.Header)}
            language={'json'}
          />
        </TabContentWithOverflow>

        {isSerializableRequest(message) && (
          <TabContentWithOverflow value="query">
            <SyntaxHighlighter
              code={jsonStringifyOrNull(
                message?.URL
                  ? searchParamsToObject(
                      new URLSearchParams(message.URL.RawQuery),
                    )
                  : null,
              )}
              language={'json'}
            />
          </TabContentWithOverflow>
        )}

        <TabContentWithOverflow value="raw">
          <SyntaxHighlighter code={message?.Raw ?? ''} language={'plaintext'} />
        </TabContentWithOverflow>
      </Tabs.Root>
    </BoxWithBorder>
  )
}

function TabContentWithOverflow({ className, ...props }: TabsContentProps) {
  return (
    <Tabs.Content
      className={cx([className, 'h-[86%] overflow-y-scroll'])}
      {...props}
    />
  )
}

function searchParamsToObject(searchParams: URLSearchParams) {
  const obj: Record<string, string[]> = {}
  for (const [key, value] of searchParams) {
    if (obj[key]) {
      obj[key].push(value)
    } else {
      obj[key] = [value]
    }
  }
  return obj
}

function jsonStringifyOrNull(obj: any) {
  return JSON.stringify(obj || null, null, 2)
}

function selectLanguage(contentType: string | undefined): string {
  if (contentType?.includes('json')) {
    return 'json'
  }
  if (contentType?.includes('xml')) {
    return 'xml'
  }
  if (contentType?.includes('html')) {
    return 'html'
  }
  return 'plaintext'
}
