import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Spin, Typography } from 'antd'
import { fetchCommand } from '../api/commands'
import CommandDetail from '../components/CommandDetail'

const { Text } = Typography

export default function CommandsPage() {
  const { slug } = useParams<{ slug?: string }>()

  const { data: selectedCommand, isLoading } = useQuery({
    queryKey: ['command', slug],
    queryFn: () => fetchCommand(slug!),
    enabled: !!slug,
  })

  if (!slug) {
    return (
      <div style={styles.placeholder}>
        <svg viewBox="0 0 48 48" fill="none" width="40" height="40" style={{ marginBottom: 16 }}>
          <rect x="8" y="14" width="32" height="22" rx="3" stroke="#d9d9d9" strokeWidth="2" />
          <path d="M16 22h16M16 28h10" stroke="#d9d9d9" strokeWidth="2" strokeLinecap="round" />
        </svg>
        <Text type="secondary" style={{ fontSize: 14, textAlign: 'center' }}>
          Your command form &amp; history will be visible here once selected.
        </Text>
      </div>
    )
  }

  if (isLoading || !selectedCommand) {
    return (
      <div style={styles.placeholder}>
        <Spin />
      </div>
    )
  }

  return <CommandDetail command={selectedCommand} />
}

const styles: Record<string, React.CSSProperties> = {
  placeholder: {
    flex: 1,
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    justifyContent: 'center',
    background: '#fff',
  },
}
