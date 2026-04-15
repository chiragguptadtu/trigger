import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Table, Button, Modal, Form, Input, Tag, Popconfirm, Select, Typography, Tooltip, message } from 'antd'
import { TeamOutlined, DeleteOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { fetchGroups, createGroup, deleteGroup, fetchGroupMembers, fetchUsers, addGroupMember, removeGroupMember } from '../../api/admin'
import type { Group, User } from '../../api/admin'
import { colTitle, tableProps } from '../../utils/table'
import { PAGE_PADDING } from '../../utils/theme'

const { Text } = Typography

export default function AdminGroupsPage() {
  const qc = useQueryClient()
  const [messageApi, contextHolder] = message.useMessage()
  const [createOpen, setCreateOpen] = useState(false)
  const [membersGroup, setMembersGroup] = useState<Group | null>(null)
  const [createForm] = Form.useForm()
  const [saving, setSaving] = useState(false)
  const [addingMember, setAddingMember] = useState(false)
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null)

  const { data: groups = [], isLoading } = useQuery({ queryKey: ['admin', 'groups'], queryFn: fetchGroups })
  const { data: allUsers = [] } = useQuery({ queryKey: ['admin', 'users'], queryFn: fetchUsers })
  const { data: members = [] } = useQuery({
    queryKey: ['admin', 'group-members', membersGroup?.id],
    queryFn: () => fetchGroupMembers(membersGroup!.id),
    enabled: !!membersGroup,
  })

  async function handleCreate(values: { name: string }) {
    setSaving(true)
    try {
      await createGroup(values.name)
      qc.invalidateQueries({ queryKey: ['admin', 'groups'] })
      setCreateOpen(false)
      createForm.resetFields()
      messageApi.success('Group created.')
    } catch {
      messageApi.error('Failed to create group.')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id: string) {
    try {
      await deleteGroup(id)
      qc.invalidateQueries({ queryKey: ['admin', 'groups'] })
      messageApi.success('Group deleted.')
    } catch {
      messageApi.error('Failed to delete group.')
    }
  }

  async function handleAddMember() {
    if (!membersGroup || !selectedUserId) return
    setAddingMember(true)
    try {
      await addGroupMember(membersGroup.id, selectedUserId)
      qc.invalidateQueries({ queryKey: ['admin', 'group-members', membersGroup.id] })
      setSelectedUserId(null)
      messageApi.success('Member added.')
    } catch {
      messageApi.error('Failed to add member.')
    } finally {
      setAddingMember(false)
    }
  }

  async function handleRemoveMember(userId: string) {
    if (!membersGroup) return
    try {
      await removeGroupMember(membersGroup.id, userId)
      qc.invalidateQueries({ queryKey: ['admin', 'group-members', membersGroup.id] })
      messageApi.success('Member removed.')
    } catch {
      messageApi.error('Failed to remove member.')
    }
  }

  const memberIds = new Set(members.map((m: User) => m.id))
  const addableUsers = allUsers.filter((u: User) => u.is_active && !memberIds.has(u.id))

  const columns: ColumnsType<Group> = [
    {
      title: colTitle('Name'),
      dataIndex: 'name',
      render: (name: string) => <Text style={{ fontSize: 13 }}>{name}</Text>,
    },
    {
      title: colTitle('Status'),
      dataIndex: 'is_active',
      width: 100,
      render: (isActive: boolean) =>
        isActive ? <Tag color="success">Active</Tag> : <Tag>Inactive</Tag>,
    },
    {
      title: colTitle('Actions'),
      width: 90,
      render: (_: unknown, group: Group) => (
        <div style={{ display: 'flex', gap: 12 }}>
          <Tooltip title="Members" color="#fff" styles={{ container: { color: 'rgba(0,0,0,0.65)', fontSize: 12 } }}>
            <Button type="text" size="small" icon={<TeamOutlined />} onClick={() => setMembersGroup(group)} />
          </Tooltip>
          <Popconfirm title="Delete this group?" onConfirm={() => handleDelete(group.id)} okText="Yes" cancelText="No">
            <Tooltip title="Delete" color="#fff" styles={{ container: { color: 'rgba(0,0,0,0.65)', fontSize: 12 } }}>
              <Button type="text" size="small" danger icon={<DeleteOutlined />} />
            </Tooltip>
          </Popconfirm>
        </div>
      ),
    },
  ]

  return (
    <div style={styles.page}>
      {contextHolder}
      <div style={styles.header}>
        <Text strong style={styles.title}>Groups</Text>
        <Button type="primary" onClick={() => setCreateOpen(true)}>Create Group</Button>
      </div>

      <Table<Group>
        {...tableProps}
        dataSource={groups}
        columns={columns}
        rowKey="id"
        loading={isLoading}
      />

      <Modal
        title="Create Group"
        open={createOpen}
        onCancel={() => { setCreateOpen(false); createForm.resetFields() }}
        onOk={() => createForm.submit()}
        confirmLoading={saving}
        okText="Create"
      >
        <Form form={createForm} layout="vertical" onFinish={handleCreate} requiredMark={false} style={{ marginTop: 16 }}>
          <Form.Item name="name" label="Name" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={`Members — ${membersGroup?.name ?? ''}`}
        open={!!membersGroup}
        onCancel={() => { setMembersGroup(null); setSelectedUserId(null) }}
        footer={null}
        width="min(500px, 90vw)"
      >
        <div style={{ marginBottom: 16, display: 'flex', gap: 8, marginTop: 16 }}>
          <Select
            style={{ flex: 1 }}
            placeholder="Add a user…"
            value={selectedUserId}
            onChange={setSelectedUserId}
            options={addableUsers.map((u: User) => ({ label: `${u.name || u.email} (${u.email})`, value: u.id }))}
            showSearch
            filterOption={(input, option) => (option?.label as string ?? '').toLowerCase().includes(input.toLowerCase())}
          />
          <Button type="primary" loading={addingMember} disabled={!selectedUserId} onClick={handleAddMember}>
            Add
          </Button>
        </div>

        <Table<User>
          {...tableProps}
          dataSource={members}
          rowKey="id"
          pagination={false}
          columns={[
            {
              title: colTitle('User'),
              render: (_: unknown, u: User) => (
                <div>
                  <Text style={{ fontSize: 13 }}>{u.name || u.email}</Text>
                  {u.name && <Text type="secondary" style={{ fontSize: 12, display: 'block' }}>{u.email}</Text>}
                </div>
              ),
            },
            {
              width: 48,
              render: (_: unknown, u: User) => (
                <Popconfirm title="Remove from group?" onConfirm={() => handleRemoveMember(u.id)} okText="Yes" cancelText="No">
                  <Tooltip title="Remove" color="#fff" styles={{ container: { color: 'rgba(0,0,0,0.65)', fontSize: 12 } }}>
                    <Button type="text" size="small" danger icon={<DeleteOutlined />} />
                  </Tooltip>
                </Popconfirm>
              ),
            },
          ]}
        />
      </Modal>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  page: { padding: PAGE_PADDING, background: '#fff', flex: 1, overflow: 'auto' },
  header: { display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 20 },
  title: { fontSize: 16 },
}
