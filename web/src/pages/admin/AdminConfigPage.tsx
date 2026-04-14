import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Table, Button, Modal, Form, Input, Popconfirm, Typography, Tooltip, message } from 'antd'
import { EditOutlined, DeleteOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { fetchConfig, createConfig, updateConfig, deleteConfig } from '../../api/admin'
import type { ConfigEntry } from '../../api/admin'
import { colTitle, tableProps } from '../../utils/table'
import { getApiError } from '../../utils/error'

const { Text } = Typography

export default function AdminConfigPage() {
  const qc = useQueryClient()
  const [messageApi, contextHolder] = message.useMessage()
  const [createOpen, setCreateOpen] = useState(false)
  const [editEntry, setEditEntry] = useState<ConfigEntry | null>(null)
  const [createForm] = Form.useForm()
  const [editForm] = Form.useForm()
  const [saving, setSaving] = useState(false)

  const { data: entries = [], isLoading } = useQuery({ queryKey: ['admin', 'config'], queryFn: fetchConfig })

  async function handleCreate(values: { key: string; value: string; description: string }) {
    setSaving(true)
    try {
      await createConfig(values)
      qc.invalidateQueries({ queryKey: ['admin', 'config'] })
      setCreateOpen(false)
      createForm.resetFields()
      messageApi.success('Config entry created.')
    } catch (e: unknown) {
      messageApi.error(getApiError(e, 'Failed to create config entry.'))
    } finally {
      setSaving(false)
    }
  }

  async function handleEdit(values: { value: string; description: string }) {
    if (!editEntry) return
    setSaving(true)
    try {
      await updateConfig(editEntry.key, values)
      qc.invalidateQueries({ queryKey: ['admin', 'config'] })
      setEditEntry(null)
      messageApi.success('Config entry updated.')
    } catch {
      messageApi.error('Failed to update config entry.')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(key: string) {
    try {
      await deleteConfig(key)
      qc.invalidateQueries({ queryKey: ['admin', 'config'] })
      messageApi.success('Config entry deleted.')
    } catch {
      messageApi.error('Failed to delete config entry.')
    }
  }

  const columns: ColumnsType<ConfigEntry> = [
    {
      title: colTitle('Key'),
      dataIndex: 'key',
      width: 280,
      ellipsis: true,
      render: (key: string) => <Text code style={{ fontSize: 13 }}>{key}</Text>,
    },
    {
      title: colTitle('Description'),
      dataIndex: 'description',
      render: (desc: string) => <Text type="secondary" style={{ fontSize: 13 }}>{desc || '—'}</Text>,
    },
    {
      title: colTitle('Updated'),
      dataIndex: 'updated_at',
      width: 160,
      render: (v: string) => <Text type="secondary" style={{ fontSize: 13 }}>{new Date(v).toLocaleString()}</Text>,
    },
    {
      title: colTitle('Actions'),
      width: 90,
      render: (_: unknown, entry: ConfigEntry) => (
        <div style={{ display: 'flex', gap: 12 }}>
          <Tooltip title="Edit" color="#fff" styles={{ body: { color: 'rgba(0,0,0,0.65)', fontSize: 12 } }}>
            <Button
              type="text"
              size="small"
              icon={<EditOutlined />}
              onClick={() => {
                setEditEntry(entry)
                editForm.setFieldsValue({ description: entry.description, value: '' })
              }}
            />
          </Tooltip>
          <Popconfirm title="Delete this entry?" onConfirm={() => handleDelete(entry.key)} okText="Yes" cancelText="No">
            <Tooltip title="Delete" color="#fff" styles={{ body: { color: 'rgba(0,0,0,0.65)', fontSize: 12 } }}>
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
        <Text strong style={styles.title}>Config</Text>
        <Button type="primary" onClick={() => setCreateOpen(true)}>Add Entry</Button>
      </div>

      <Table<ConfigEntry>
        {...tableProps}
        dataSource={entries}
        columns={columns}
        rowKey="id"
        loading={isLoading}
      />

      <Modal
        title="Add Config Entry"
        open={createOpen}
        onCancel={() => { setCreateOpen(false); createForm.resetFields() }}
        onOk={() => createForm.submit()}
        confirmLoading={saving}
        okText="Create"
      >
        <Form form={createForm} layout="vertical" onFinish={handleCreate} requiredMark={false} style={{ marginTop: 16 }}>
          <Form.Item name="key" label="Key" rules={[{ required: true }]}>
            <Input placeholder="e.g. STRIPE_SECRET_KEY" />
          </Form.Item>
          <Form.Item name="value" label="Value" rules={[{ required: true }]}>
            <Input.Password placeholder="Secret value (encrypted at rest)" />
          </Form.Item>
          <Form.Item name="description" label="Description">
            <Input />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={`Edit — ${editEntry?.key ?? ''}`}
        open={!!editEntry}
        onCancel={() => setEditEntry(null)}
        onOk={() => editForm.submit()}
        confirmLoading={saving}
        okText="Save"
      >
        <Form form={editForm} layout="vertical" onFinish={handleEdit} requiredMark={false} style={{ marginTop: 16 }}>
          <Form.Item name="value" label="New Value" rules={[{ required: true }]} extra="Values are encrypted at rest. You must re-enter the value to update it.">
            <Input.Password />
          </Form.Item>
          <Form.Item name="description" label="Description">
            <Input />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  page: { padding: '28px 32px', background: '#fff', flex: 1, overflow: 'auto' },
  header: { display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 20 },
  title: { fontSize: 16 },
}
