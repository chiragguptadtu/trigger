import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import type { MenuProps } from 'antd'
import { Dropdown, Popover, Typography } from 'antd'
import { useQuery } from '@tanstack/react-query'
import Brand from './Brand'
import { decodeClaims } from '../utils/jwt'
import { PRIMARY } from '../utils/theme'
import { fetchImportErrors } from '../api/admin'
import type { ImportError } from '../api/admin'

const { Text } = Typography

function ImportErrorIndicator() {
  const [open, setOpen] = useState(false)

  const hasToken = !!localStorage.getItem('token')

  const { data: errors = [] } = useQuery({
    queryKey: ['import-errors'],
    queryFn: fetchImportErrors,
    refetchInterval: 30_000,
    enabled: hasToken,
  })

  if (errors.length === 0) return null

  const label =
    errors.length === 1 ? errors[0].filename : `${errors.length} import errors`

  const content = (
    <div style={{ width: '35vw' }}>
      {errors.map((e: ImportError, i: number) => (
        <div
          key={i}
          style={{
            paddingBottom: i < errors.length - 1 ? 12 : 0,
            marginBottom: i < errors.length - 1 ? 12 : 0,
            borderBottom: i < errors.length - 1 ? '1px solid #f0f0f0' : 'none',
          }}
        >
          <Text strong style={{ fontSize: 13 }}>{e.filename}</Text>
          <br />
          <Text type="danger" style={{ fontSize: 12, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
            {e.error}
          </Text>
          <br />
          <Text type="secondary" style={{ fontSize: 11 }}>{e.failed_at}</Text>
        </div>
      ))}
    </div>
  )

  return (
    <Popover
      open={open}
      onOpenChange={setOpen}
      trigger="click"
      placement="bottom"
      title={
        <Text type="danger" strong>
          {errors.length} import {errors.length === 1 ? 'error' : 'errors'}
        </Text>
      }
      content={content}
      styles={{ container: { maxWidth: '38vw' } }}
    >
      <span style={styles.errorLink}>{label}</span>
    </Popover>
  )
}

export default function Navbar() {
  const navigate = useNavigate()
  const claims = decodeClaims()

  function logout() {
    localStorage.removeItem('token')
    navigate('/login')
  }

  const adminItems: MenuProps['items'] = claims?.admin
    ? [
      { key: 'admin-header', type: 'group', label: 'Admin' },
      { key: 'users', label: 'Users', onClick: () => navigate('/admin/users') },
      { key: 'groups', label: 'Groups', onClick: () => navigate('/admin/groups') },
      { key: 'config', label: 'Config', onClick: () => navigate('/admin/config') },
      { key: 'permissions', label: 'Permissions', onClick: () => navigate('/admin/permissions') },
      { type: 'divider' },
    ]
    : []

  const menuItems: MenuProps['items'] = [
    ...adminItems,
    { key: 'logout', label: 'Logout', danger: true, onClick: logout },
  ]

  return (
    <header style={styles.navbar}>
      <div style={styles.left}>
        <Brand />
      </div>

      <div style={styles.center}>
        <ImportErrorIndicator />
      </div>

      <div style={styles.right}>
        {claims && (
          <Dropdown menu={{ items: menuItems }} trigger={['hover', 'click']} placement="bottomRight">
            <div style={styles.avatar}>
              <svg viewBox="0 0 24 24" fill="white" width="15" height="15">
                <path d="M12 12c2.48 0 4.5-2.02 4.5-4.5S14.48 3 12 3 7.5 5.02 7.5 7.5 9.52 12 12 12zm0 2.25c-3.01 0-9 1.51-9 4.5V21h18v-2.25c0-2.99-5.99-4.5-9-4.5z" />
              </svg>
            </div>
          </Dropdown>
        )}
      </div>
    </header>
  )
}

const styles: Record<string, React.CSSProperties> = {
  navbar: {
    height: 56,
    background: '#fff',
    borderBottom: '1px solid #f0f0f0',
    display: 'flex',
    alignItems: 'center',
    padding: '0 24px',
    position: 'sticky',
    top: 0,
    zIndex: 100,
  },
  left: {
    flex: 1,
    display: 'flex',
    alignItems: 'center',
  },
  center: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
  },
  right: {
    flex: 1,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'flex-end',
  },
  avatar: {
    width: 32,
    height: 32,
    borderRadius: '50%',
    background: PRIMARY,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    cursor: 'pointer',
    flexShrink: 0,
  },
  errorLink: {
    fontSize: 12,
    color: '#ff4d4f',
    cursor: 'pointer',
    maxWidth: '22vw',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
    display: 'block',
    textDecoration: 'underline',
    textUnderlineOffset: 2,
  },
}
