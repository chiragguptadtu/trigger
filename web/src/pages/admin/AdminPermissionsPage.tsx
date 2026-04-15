import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Table, Button, Modal, Form, Select, Tag, Popconfirm, Typography, Tooltip, message } from 'antd'
import { StopOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { fetchPermissions, grantPermission, revokePermission, fetchUsers, fetchGroups } from '../../api/admin'
import type { Permission, User, Group } from '../../api/admin'
import { fetchCommands } from '../../api/commands'
import type { Command } from '../../api/commands'
import { colTitle, tableProps } from '../../utils/table'
import { getApiError } from '../../utils/error'
import { PAGE_PADDING } from '../../utils/theme'

const { Text } = Typography

interface ResolvedPermission extends Permission {
  grantee_name: string
}

export default function AdminPermissionsPage() {
  const qc = useQueryClient()
  const [messageApi, contextHolder] = message.useMessage()
  const [selectedSlug, setSelectedSlug] = useState<string | null>(null)
  const [grantOpen, setGrantOpen] = useState(false)
  const [granteeType, setGranteeType] = useState<'user' | 'group'>('user')
  const [granteeId, setGranteeId] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)

  const { data: commands = [] } = useQuery({ queryKey: ['commands'], queryFn: fetchCommands })
  const { data: users = [] } = useQuery({ queryKey: ['admin', 'users'], queryFn: fetchUsers })
  const { data: groups = [] } = useQuery({ queryKey: ['admin', 'groups'], queryFn: fetchGroups })
  const { data: permissions = [], isLoading } = useQuery({
    queryKey: ['admin', 'permissions', selectedSlug],
    queryFn: () => fetchPermissions(selectedSlug!),
    enabled: !!selectedSlug,
  })

  const userMap = Object.fromEntries(users.map((u: User) => [u.id, u.name || u.email]))
  const groupMap = Object.fromEntries(groups.map((g: Group) => [g.id, g.name]))

  const resolved: ResolvedPermission[] = permissions.map((p: Permission) => ({
    ...p,
    grantee_name: p.grantee_type === 'user' ? (userMap[p.grantee_id] ?? p.grantee_id) : (groupMap[p.grantee_id] ?? p.grantee_id),
  }))

  const granteeOptions = granteeType === 'user'
    ? users.filter((u: User) => u.is_active).map((u: User) => ({ label: `${u.name || u.email} (${u.email})`, value: u.id }))
    : groups.filter((g: Group) => g.is_active).map((g: Group) => ({ label: g.name, value: g.id }))

  async function handleGrant() {
    if (!selectedSlug || !granteeId) return
    setSaving(true)
    try {
      await grantPermission(selectedSlug, { grantee_type: granteeType, grantee_id: granteeId })
      qc.invalidateQueries({ queryKey: ['admin', 'permissions', selectedSlug] })
      setGrantOpen(false)
      setGranteeId(null)
      messageApi.success('Access granted.')
    } catch (e: unknown) {
      messageApi.error(getApiError(e, 'Failed to grant access.'))
    } finally {
      setSaving(false)
    }
  }

  async function handleRevoke(p: Permission) {
    if (!selectedSlug) return
    try {
      await revokePermission(selectedSlug, p.grantee_type, p.grantee_id)
      qc.invalidateQueries({ queryKey: ['admin', 'permissions', selectedSlug] })
      messageApi.success('Access revoked.')
    } catch {
      messageApi.error('Failed to revoke access.')
    }
  }

  const columns: ColumnsType<ResolvedPermission> = [
    {
      title: colTitle('Type'),
      dataIndex: 'grantee_type',
      width: 90,
      render: (t: string) => <Tag color={t === 'user' ? 'blue' : 'purple'}>{t}</Tag>,
    },
    {
      title: colTitle('Name'),
      dataIndex: 'grantee_name',
      render: (name: string) => <Text style={{ fontSize: 13 }}>{name}</Text>,
    },
    {
      title: colTitle('Actions'),
      width: 60,
      render: (_: unknown, p: ResolvedPermission) => (
        <Popconfirm title="Revoke access?" onConfirm={() => handleRevoke(p)} okText="Yes" cancelText="No">
          <Tooltip title="Revoke" color="#fff" styles={{ container: { color: 'rgba(0,0,0,0.65)', fontSize: 12 } }}>
            <Button type="text" size="small" danger icon={<StopOutlined />} />
          </Tooltip>
        </Popconfirm>
      ),
    },
  ]

  return (
    <div style={styles.page}>
      {contextHolder}
      <div style={styles.header}>
        <Text strong style={styles.title}>Permissions</Text>
        <Button type="primary" disabled={!selectedSlug} onClick={() => { setGrantOpen(true); setGranteeId(null) }}>
          Grant Access
        </Button>
      </div>

      <Form layout="vertical" requiredMark={false} style={{ marginBottom: 20 }}>
        <Form.Item label={<span>Command <span style={{ color: 'rgba(0,0,0,0.35)', fontWeight: 400 }}>(select a command to view and manage its access grants)</span></span>} style={{ marginBottom: 0 }}>
          <Select
            style={{ width: '22vw' }}
            placeholder="Select a command…"
            value={selectedSlug}
            onChange={setSelectedSlug}
            options={commands.map((c: Command) => ({ label: c.name, value: c.slug }))}
            showSearch
            filterOption={(input, option) => (option?.label as string ?? '').toLowerCase().includes(input.toLowerCase())}
          />
        </Form.Item>
      </Form>

      {selectedSlug && (
        <Table<ResolvedPermission>
          {...tableProps}
          dataSource={resolved}
          columns={columns}
          rowKey="id"
          loading={isLoading}
          locale={{ emptyText: 'No access grants for this command.' }}
        />
      )}

      <Modal
        title="Grant Access"
        open={grantOpen}
        onCancel={() => setGrantOpen(false)}
        onOk={handleGrant}
        confirmLoading={saving}
        okText="Grant"
        okButtonProps={{ disabled: !granteeId }}
      >
        <Form layout="vertical" requiredMark={false} style={{ marginTop: 16 }}>
          <Form.Item label="Type">
            <Select
              value={granteeType}
              onChange={(v) => { setGranteeType(v); setGranteeId(null) }}
              options={[{ label: 'User', value: 'user' }, { label: 'Group', value: 'group' }]}
            />
          </Form.Item>
          <Form.Item label={granteeType === 'user' ? 'User' : 'Group'} style={{ marginBottom: 0 }}>
            <Select
              placeholder={`Select a ${granteeType}…`}
              value={granteeId}
              onChange={setGranteeId}
              options={granteeOptions}
              showSearch
              filterOption={(input, option) => (option?.label as string ?? '').toLowerCase().includes(input.toLowerCase())}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  page: { padding: PAGE_PADDING, background: '#fff', flex: 1, overflow: 'auto' },
  header: { display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 20 },
  title: { fontSize: 16 },
}
