import { PRIMARY } from './theme'

export function colTitle(label: string) {
  return <span style={{ color: PRIMARY, fontSize: 12, fontWeight: 600 }}>{label}</span>
}

export const tableProps = {
  size: 'small' as const,
  bordered: true,
  rowClassName: (_: unknown, index: number) => index % 2 !== 0 ? 'row-alt' : '',
  pagination: { pageSize: 20, hideOnSinglePage: true, showSizeChanger: false },
} as const
