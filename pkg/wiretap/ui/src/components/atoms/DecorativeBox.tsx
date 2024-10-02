import { cx } from 'class-variance-authority'
import type { HTMLAttributes } from 'react'

const baseBoxStyle = cx(['border border-1 border-gray-600 rounded bg-gray-f2'])

interface BoxProps extends HTMLAttributes<HTMLDivElement> {}

export function DecorativeBox({ style, children, ...props }: BoxProps) {
  return (
    <div
      className={cx([baseBoxStyle])}
      style={{
        ...style,
        backgroundImage:
          "url(\"data:image/svg+xml,%3Csvg width='6' height='6' viewBox='0 0 6 6' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='%239C92AC' fill-opacity='0.2' fill-rule='evenodd'%3E%3Cpath d='M5 0h1L0 6V5zM6 5v1H5z'/%3E%3C/g%3E%3C/svg%3E\")",
      }}
      {...props}
    >
      {children}
    </div>
  )
}
export function BoxWithBorder({
  style,
  className,
  children,
  ...props
}: BoxProps) {
  return (
    <div className={cx([baseBoxStyle, className])} style={style} {...props}>
      {children}
    </div>
  )
}
