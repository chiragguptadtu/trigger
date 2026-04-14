import { useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Input, Spin, Tooltip, Typography } from 'antd'
import { fetchCommands } from '../api/commands'
import type { Command } from '../api/commands'

const { Text } = Typography

const ADMIN_LINKS = [
  { key: 'users', label: 'Users', path: '/admin/users' },
  { key: 'groups', label: 'Groups', path: '/admin/groups' },
  { key: 'config', label: 'Config', path: '/admin/config' },
  { key: 'permissions', label: 'Permissions', path: '/admin/permissions' },
]

function AdminSidebar() {
  const location = useLocation()
  const navigate = useNavigate()

  return (
    <aside style={styles.sidebar}>
      <div style={styles.sidebarHeader}>
        <Text strong style={styles.sidebarHeaderText}>Admin</Text>
      </div>
      <div style={styles.list}>
        {ADMIN_LINKS.map((link) => {
          const active = location.pathname.startsWith(link.path)
          return (
            <div
              key={link.key}
              style={{ ...styles.item, ...(active ? styles.itemActive : {}) }}
              onClick={() => navigate(link.path)}
            >
              <Text style={{ ...styles.itemName, ...(active ? styles.itemNameActive : {}) }}>
                {link.label}
              </Text>
            </div>
          )
        })}
      </div>
    </aside>
  )
}

function CommandsSidebar() {
  const location = useLocation()
  const navigate = useNavigate()
  const [search, setSearch] = useState('')

  const slug = location.pathname.startsWith('/commands/')
    ? location.pathname.split('/')[2]
    : undefined

  const { data: commands = [], isLoading } = useQuery({
    queryKey: ['commands'],
    queryFn: fetchCommands,
  })

  const filtered = commands.filter(
    (c) =>
      c.name.toLowerCase().includes(search.toLowerCase()) ||
      c.description.toLowerCase().includes(search.toLowerCase()),
  )

  return (
    <aside style={styles.sidebar}>
      <div style={styles.searchWrap}>
        <Input
          placeholder="Search commands…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          allowClear
        />
      </div>
      <div style={styles.list}>
        {isLoading && (
          <div style={styles.centerSpin}>
            <Spin size="small" />
          </div>
        )}
        {!isLoading && filtered.length === 0 && (
          <Text type="secondary" style={styles.emptyList}>
            No commands found.
          </Text>
        )}
        {filtered.map((cmd: Command) => {
          const active = cmd.slug === slug
          return (
            <div
              key={cmd.id}
              style={{ ...styles.item, ...(active ? styles.itemActive : {}) }}
              onClick={() => navigate(`/commands/${cmd.slug}`)}
            >
              <Text strong style={{ ...styles.itemName, ...(active ? styles.itemNameActive : {}) }}>
                {cmd.name}
              </Text>
              {cmd.description && (
                <Tooltip title={cmd.description} placement="right" mouseEnterDelay={0.5} color="#fff" overlayInnerStyle={{ color: 'rgba(0,0,0,0.65)', fontSize: 12 }}>
                  <Text type="secondary" style={styles.itemDesc}>
                    {cmd.description}
                  </Text>
                </Tooltip>
              )}
            </div>
          )
        })}
      </div>
    </aside>
  )
}

export default function Sidebar() {
  const location = useLocation()
  if (location.pathname.startsWith('/admin')) return <AdminSidebar />
  return <CommandsSidebar />
}

const styles: Record<string, React.CSSProperties> = {
  sidebar: {
    width: 260,
    flexShrink: 0,
    background: '#fff',
    borderRight: '1px solid #f0f0f0',
    display: 'flex',
    flexDirection: 'column',
    overflow: 'hidden',
  },
  sidebarHeader: {
    padding: '14px 16px 10px',
    borderBottom: '1px solid #f5f5f5',
  },
  sidebarHeaderText: {
    fontSize: 12,
    color: 'rgba(0,0,0,0.45)',
    textTransform: 'uppercase',
    letterSpacing: '0.5px',
  },
  searchWrap: {
    padding: '12px 12px 8px',
    borderBottom: '1px solid #f5f5f5',
  },
  list: {
    flex: 1,
    overflowY: 'auto',
    padding: '8px 0',
  },
  centerSpin: {
    display: 'flex',
    justifyContent: 'center',
    padding: '24px 0',
  },
  emptyList: {
    display: 'block',
    textAlign: 'center',
    padding: '24px 16px',
    fontSize: 13,
  },
  item: {
    padding: '9px 16px',
    cursor: 'pointer',
    display: 'flex',
    flexDirection: 'column',
    gap: 2,
    borderRadius: 6,
    margin: '1px 8px',
  },
  itemActive: {
    background: '#eff8ff',
  },
  itemName: {
    fontSize: 13,
    color: 'rgba(0,0,0,0.75)',
  },
  itemNameActive: {
    color: '#1677ff',
  },
  itemDesc: {
    fontSize: 12,
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
    maxWidth: 220,
  },
}
