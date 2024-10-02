import htmlParser from 'prettier/plugins/html'
import prettier from 'prettier/standalone'
import { Suspense } from 'react'
import { Await } from 'react-router-dom'
import ReactSyntaxHighlighter from 'react-syntax-highlighter'
import theme from 'react-syntax-highlighter/dist/esm/styles/hljs/railscasts'

export interface SyntaxHighlighterProps {
  code: string
  language: string
}

async function format(code: string, language: string) {
  if (code.trim() === '') {
    return 'No content'
  }
  if (language === 'json') {
    return JSON.stringify(JSON.parse(code), null, 2)
  }

  if (language === 'plaintext') {
    return code
  }

  return prettier.format(code, {
    parser: language,
    plugins: [htmlParser],
  })
}

export function SyntaxHighlighter(props: SyntaxHighlighterProps) {
  return (
    <Suspense fallback={<div>Decoding...</div>}>
      <Await
        errorElement={
          <p>Failed to parse body for content type: {props.language}</p>
        }
        resolve={format(props.code, props.language)}
      >
        {(code) => (
          <ReactSyntaxHighlighter
            className={'text-xs'}
            customStyle={{ background: 'transparent' }}
            language={props.language}
            showLineNumbers={true}
            style={{ ...theme }}
            wrapLines={true}
            wrapLongLines={true}
          >
            {code}
          </ReactSyntaxHighlighter>
        )}
      </Await>
    </Suspense>
  )
}
