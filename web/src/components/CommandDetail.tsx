import { useState, useEffect, useRef } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Card, Form, Input, Select, Button, Typography, Table, Tag, Tooltip } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import type { Command } from '../api/commands'
import type { Execution } from '../api/executions'
import { triggerExecution, fetchExecution, fetchExecutions } from '../api/executions'
import { colTitle, tableProps } from '../utils/table'

import { AVATAR_COLORS } from '../utils/theme'
import { getApiError } from '../utils/error'

const { Title, Text } = Typography

const STATUS_COLOR: Record<string, string> = {
  pending: 'default',
  running: 'processing',
  success: 'success',
  failure: 'error',
}

const STATUS_LABEL: Record<string, string> = {
  pending: 'Pending',
  running: 'Running',
  success: 'Success',
  failure: 'Failed',
}

function statusTag(status: string, errorMessage: string) {
  const tag = (
    <Tag color={STATUS_COLOR[status] ?? 'default'} style={status === 'failure' ? { cursor: 'help' } : {}}>
      {STATUS_LABEL[status] ?? status}
    </Tag>
  )
  if (status === 'failure' && errorMessage) {
    return <Tooltip title={errorMessage}>{tag}</Tooltip>
  }
  return tag
}

function userAvatar(name: string, email: string) {
  const display = name || email || '?'
  const initials = display
    .split(' ')
    .map((p) => p[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)
  const color = AVATAR_COLORS[display.charCodeAt(0) % AVATAR_COLORS.length]
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
      <div style={{
        width: 24, height: 24, borderRadius: '50%', background: color,
        display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0,
      }}>
        <span style={{ fontSize: 10, color: '#fff', fontWeight: 600 }}>{initials}</span>
      </div>
      <Text style={{ fontSize: 13 }}>{display}</Text>
    </div>
  )
}


const columns: ColumnsType<Execution> = [
  {
    title: colTitle('Parameters'),
    dataIndex: 'inputs',
    render: (inputs: Record<string, unknown> | undefined) => {
      if (!inputs || Object.keys(inputs).length === 0) {
        return <Text type="secondary" style={{ fontSize: 13 }}>—</Text>
      }
      return (
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
          {Object.entries(inputs).map(([k, v]) => (
            <Tag key={k} style={{ margin: 0, fontSize: 12 }}>
              <span style={{ color: 'rgba(0,0,0,0.4)' }}>{k}: </span>
              <span style={{ color: 'rgba(0,0,0,0.88)', fontWeight: 500 }}>
                {Array.isArray(v) ? v.join(', ') : String(v)}
              </span>
            </Tag>
          ))}
        </div>
      )
    },
  },
  {
    title: colTitle('Status'),
    dataIndex: 'status',
    width: 110,
    render: (s: string, row: Execution) => statusTag(s, row.error_message),
  },
  {
    title: colTitle('Triggered by'),
    dataIndex: 'triggered_by_name',
    width: 200,
    render: (name: string, row: Execution) => userAvatar(name, row.triggered_by_email),
  },
  {
    title: colTitle('Triggered at'),
    dataIndex: 'created_at',
    width: 170,
    render: (v: string) => (
      <Text type="secondary" style={{ fontSize: 13 }}>
        {new Date(v).toLocaleString()}
      </Text>
    ),
  },
]

interface Props {
  command: Command
}

export default function CommandDetail({ command }: Props) {
  const [form] = Form.useForm()
  const [running, setRunning] = useState(false)
  const [result, setResult] = useState<{ type: 'success' | 'error'; message: string } | null>(null)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const resultTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const queryClient = useQueryClient()

  function showResult(r: { type: 'success' | 'error'; message: string }) {
    if (resultTimerRef.current) clearTimeout(resultTimerRef.current)
    setResult(r)
    resultTimerRef.current = setTimeout(() => setResult(null), 5000)
  }

  const { data: executions = [], isLoading: historyLoading } = useQuery({
    queryKey: ['executions', command.slug],
    queryFn: () => fetchExecutions(command.slug),
  })

  // Reset form and stop any in-flight poll/timer when the selected command changes.
  useEffect(() => {
    form.resetFields()
    setResult(null)
    stopPolling()
    if (resultTimerRef.current) clearTimeout(resultTimerRef.current)
  }, [command.slug, form])

  // Stop polling and clear timer when the component unmounts.
  useEffect(() => {
    return () => {
      stopPolling()
      if (resultTimerRef.current) clearTimeout(resultTimerRef.current)
    }
  }, [])

  function stopPolling() {
    if (pollRef.current) {
      clearInterval(pollRef.current)
      pollRef.current = null
    }
  }

  async function handleSubmit(values: Record<string, unknown>) {
    setRunning(true)
    setResult(null)
    try {
      const exec = await triggerExecution(command.slug, values)
      queryClient.invalidateQueries({ queryKey: ['executions', command.slug] })

      pollRef.current = setInterval(async () => {
        const updated = await fetchExecution(exec.id)
        if (updated.status === 'success') {
          stopPolling()
          setRunning(false)
          showResult({ type: 'success', message: 'Command completed successfully.' })
          queryClient.invalidateQueries({ queryKey: ['executions', command.slug] })
        } else if (updated.status === 'failure') {
          stopPolling()
          setRunning(false)
          showResult({ type: 'error', message: updated.error_message || 'Command failed.' })
          queryClient.invalidateQueries({ queryKey: ['executions', command.slug] })
        }
      }, 1500)
    } catch (err: unknown) {
      setRunning(false)
      showResult({ type: 'error', message: getApiError(err, 'Failed to start command.') })
    }
  }

  return (
    <div style={styles.wrap}>
      {/* Header card */}
      <Card size="small" style={styles.card} styles={{ body: styles.headerBody }}>
        <Title level={5} style={styles.title}>{command.name}</Title>
        {command.description && (
          <Text type="secondary" style={styles.desc}>{command.description}</Text>
        )}
      </Card>

      {/* Form card */}
      <Card size="small" style={styles.card} styles={{ body: styles.formBody }}>
        <Form
          form={form}
          layout="vertical"
          size="small"
          onFinish={handleSubmit}
          requiredMark={false}
        >
          <div style={styles.formScroll}>
            {/* Label row — fixed height, acts as column headers */}
            <div style={styles.formLabelRow}>
              {command.inputs.map((input) => (
                <div key={input.name} style={styles.formCell}>
                  <Text style={styles.formLabel}>
                    {input.label}
                    {input.required && <span style={{ color: '#ff4d4f', marginLeft: 2 }}>*</span>}
                  </Text>
                </div>
              ))}
            </div>

            {/* Input row — can grow downward independently */}
            <div style={styles.formInputRow}>
              {command.inputs.map((input) => (
                <Form.Item
                  key={input.name}
                  name={input.name}
                  style={styles.formCell}
                  rules={input.required ? [{ required: true, message: `${input.label} is required` }] : []}
                >
                  {input.type === 'open' ? (
                    <Input placeholder={input.label} />
                  ) : input.multi ? (
                    <Select
                      mode="multiple"
                      options={(input.options ?? []).map((o) => ({ label: o, value: o }))}
                      placeholder={`Select ${input.label}`}
                    />
                  ) : (
                    <Select
                      options={(input.options ?? []).map((o) => ({ label: o, value: o }))}
                      placeholder={`Select ${input.label}`}
                    />
                  )}
                </Form.Item>
              ))}
            </div>
          </div>

          <div style={styles.formActions}>
            <Button type="primary" htmlType="submit" loading={running} disabled={running}>
              {running ? 'Running…' : 'Run'}
            </Button>
            {result && (
              <Text type={result.type === 'success' ? 'success' : 'danger'} style={{ fontSize: 13 }}>
                {result.message}
              </Text>
            )}
          </div>
        </Form>
      </Card>

      {/* History card */}
      <Card
        size="small"
        style={{ ...styles.card, flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column' }}
        styles={{ body: styles.historyBody }}
      >
        <Table<Execution>
          {...tableProps}
          dataSource={executions}
          columns={columns}
          rowKey="id"
          loading={historyLoading}
          pagination={{ pageSize: 10, hideOnSinglePage: true }}
        />
      </Card>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  wrap: {
    width: '100%',
    height: '100%',
    display: 'flex',
    flexDirection: 'column',
    gap: 12,
    padding: 16,
    overflow: 'hidden',
    boxSizing: 'border-box',
  },
  card: {
    flexShrink: 0,
    borderRadius: 8,
  },
  headerBody: {
    padding: '12px 16px',
  },
  title: {
    margin: 0,
    fontWeight: 600,
    lineHeight: 1.3,
  },
  desc: {
    fontSize: 12,
  },
  formBody: {
    padding: '12px 16px',
  },
  formScroll: {
    overflowX: 'auto',
    paddingBottom: 24,
  },
  formLabelRow: {
    display: 'flex',
    gap: 12,
    marginBottom: 4,
  },
  formInputRow: {
    display: 'flex',
    gap: 12,
    alignItems: 'flex-start',
  },
  formCell: {
    width: '13vw',
    flexShrink: 0,
    marginBottom: 0,
  },
  formLabel: {
    fontSize: 12,
    color: 'rgba(0,0,0,0.45)',
    fontWeight: 500,
    whiteSpace: 'nowrap' as const,
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    display: 'block',
  },
  formActions: {
    display: 'flex',
    alignItems: 'center',
    gap: 12,
    paddingTop: 4,
  },
  historyBody: {
    padding: '0 0',
    flex: 1,
    overflow: 'auto',
  },
}
