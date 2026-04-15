import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Table, Button, Modal, Form, Input, Switch, Tag, Popconfirm, Typography, message, Tooltip } from 'antd'
import { EditOutlined, StopOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { fetchUsers, createUser, updateUser, deactivateUser } from '../../api/admin'
import { getApiError } from '../../utils/error'
import type { User } from '../../api/admin'
import { colTitle, tableProps } from '../../utils/table'
import { PAGE_PADDING } from '../../utils/theme'

const { Text } = Typography

export default function AdminUsersPage() {
  const qc = useQueryClient()
  const [messageApi, contextHolder] = message.useMessage()
  const [createOpen, setCreateOpen] = useState(false)
  const [editUser, setEditUser] = useState<User | null>(null)
  const [createForm] = Form.useForm()
  const [editForm] = Form.useForm()
  const [saving, setSaving] = useState(false)

  const { data: users = [], isLoading } = useQuery({ queryKey: ['admin', 'users'], queryFn: fetchUsers })

  async function handleCreate(values: { email: string; name: string; password: string; is_admin: boolean }) {
    setSaving(true)
    try {
      await createUser(values)
      qc.invalidateQueries({ queryKey: ['admin', 'users'] })
      setCreateOpen(false)
      createForm.resetFields()
      messageApi.success('User created.')
    } catch (e: unknown) {
      messageApi.error(getApiError(e, 'Failed to create user.'))
    } finally {
      setSaving(false)
    }
  }

  async function handleEdit(values: { name: string; password?: string; is_admin: boolean; is_active: boolean }) {
    if (!editUser) return
    setSaving(true)
    try {
      const payload: Record<string, unknown> = { name: values.name, is_admin: values.is_admin, is_active: values.is_active }
      if (values.password) payload.password = values.password
      await updateUser(editUser.id, payload)
      qc.invalidateQueries({ queryKey: ['admin', 'users'] })
      setEditUser(null)
      messageApi.success('User updated.')
    } catch {
      messageApi.error('Failed to update user.')
    } finally {
      setSaving(false)
    }
  }

  async function handleDeactivate(id: string) {
    try {
      await deactivateUser(id)
      qc.invalidateQueries({ queryKey: ['admin', 'users'] })
      messageApi.success('User deactivated.')
    } catch {
      messageApi.error('Failed to deactivate user.')
    }
  }

  const columns: ColumnsType<User> = [
    {
      title: colTitle('Name'),
      dataIndex: 'name',
      render: (name: string) => <Text style={{ fontSize: 13 }}>{name || '—'}</Text>,
    },
    {
      title: colTitle('Email'),
      dataIndex: 'email',
      render: (email: string) => <Text style={{ fontSize: 13 }}>{email}</Text>,
    },
    {
      title: colTitle('Role'),
      dataIndex: 'is_admin',
      width: 100,
      render: (isAdmin: boolean) =>
        isAdmin ? <Tag color="blue">Admin</Tag> : <Tag>Operator</Tag>,
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
      render: (_: unknown, user: User) => (
        <div style={{ display: 'flex', gap: 12 }}>
          <Tooltip title="Edit" color="#fff" styles={{ container: { color: 'rgba(0,0,0,0.65)', fontSize: 12 } }}>
            <Button
              type="text"
              size="small"
              icon={<EditOutlined />}
              onClick={() => {
                setEditUser(user)
                editForm.setFieldsValue({ name: user.name, is_admin: user.is_admin, is_active: user.is_active, password: '' })
              }}
            />
          </Tooltip>
          {user.is_active && (
            <Popconfirm title="Deactivate this user?" onConfirm={() => handleDeactivate(user.id)} okText="Yes" cancelText="No">
              <Tooltip title="Deactivate" color="#fff" styles={{ container: { color: 'rgba(0,0,0,0.65)', fontSize: 12 } }}>
                <Button type="text" size="small" danger icon={<StopOutlined />} />
              </Tooltip>
            </Popconfirm>
          )}
        </div>
      ),
    },
  ]

  return (
    <div style={styles.page}>
      {contextHolder}
      <div style={styles.header}>
        <Text strong style={styles.title}>Users</Text>
        <Button type="primary" onClick={() => setCreateOpen(true)}>Create User</Button>
      </div>

      <Table<User>
        {...tableProps}
        dataSource={users}
        columns={columns}
        rowKey="id"
        loading={isLoading}
      />

      {/* Create modal */}
      <Modal
        title="Create User"
        open={createOpen}
        onCancel={() => { setCreateOpen(false); createForm.resetFields() }}
        onOk={() => createForm.submit()}
        confirmLoading={saving}
        okText="Create"
      >
        <Form form={createForm} layout="vertical" onFinish={handleCreate} requiredMark={false} style={{ marginTop: 16 }}>
          <Form.Item name="name" label="Name">
            <Input />
          </Form.Item>
          <Form.Item name="email" label="Email" rules={[{ required: true }, { type: 'email' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="password" label="Password" rules={[{ required: true, min: 6 }]}>
            <Input.Password />
          </Form.Item>
          <Form.Item name="is_admin" label="Admin" valuePropName="checked" initialValue={false}>
            <Switch />
          </Form.Item>
        </Form>
      </Modal>

      {/* Edit modal */}
      <Modal
        title="Edit User"
        open={!!editUser}
        onCancel={() => setEditUser(null)}
        onOk={() => editForm.submit()}
        confirmLoading={saving}
        okText="Save"
      >
        <Form form={editForm} layout="vertical" onFinish={handleEdit} requiredMark={false} style={{ marginTop: 16 }}>
          <Form.Item name="name" label="Name">
            <Input />
          </Form.Item>
          <Form.Item name="password" label="New Password" extra="Leave blank to keep current password.">
            <Input.Password />
          </Form.Item>
          <Form.Item name="is_admin" label="Admin" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="is_active" label="Active" valuePropName="checked">
            <Switch />
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
