import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { Form, Input, Button, Alert, Typography } from 'antd'
import Navbar from '../components/Navbar'
import { login } from '../api/auth'
import { getApiError } from '../utils/error'

const { Text } = Typography

export default function LoginPage() {
  const navigate = useNavigate()
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  async function handleSubmit(values: { email: string; password: string }) {
    setLoading(true)
    setError(null)
    try {
      const token = await login(values.email, values.password)
      localStorage.setItem('token', token)
      navigate('/')
    } catch (err: unknown) {
      setError(getApiError(err, 'Login failed. Please try again.'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={styles.page}>
      <Navbar />

      <div style={styles.body}>
        <div style={styles.card}>
          <Form layout="vertical" onFinish={handleSubmit} requiredMark={false}>
            <Form.Item
              name="email"
              label="Email"
              rules={[
                { required: true, message: 'Email is required' },
                { type: 'email', message: 'Enter a valid email' },
              ]}
            >
              <Input
                size="large"
                placeholder="you@company.com"
                autoComplete="email"
                autoFocus
              />
            </Form.Item>

            <Form.Item
              name="password"
              label="Password"
              rules={[{ required: true, message: 'Password is required' }]}
              style={{ marginBottom: error ? 16 : 24 }}
            >
              <Input.Password
                size="large"
                placeholder="••••••••"
                autoComplete="current-password"
              />
            </Form.Item>

            {error && (
              <Alert
                message={error}
                type="error"
                showIcon
                style={{ marginBottom: 24 }}
              />
            )}

            <Form.Item style={{ marginBottom: 16 }}>
              <Button
                type="primary"
                htmlType="submit"
                size="large"
                block
                loading={loading}
              >
                Sign in
              </Button>
            </Form.Item>
          </Form>

          <div style={styles.signupRow}>
            <Text type="secondary" style={{ fontSize: 13 }}>
              Don't have an account?{' '}
              <Link to="/signup">Sign up</Link>
            </Text>
          </div>
        </div>
      </div>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  page: {
    minHeight: '100vh',
    display: 'flex',
    flexDirection: 'column',
    background: '#f5f5f5',
  },
  body: {
    flex: 1,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
  },
  card: {
    width: 'min(380px, 90vw)',
    background: '#fff',
    borderRadius: 12,
    padding: '32px 36px 28px',
    boxShadow: '0 2px 8px rgba(0,0,0,0.07), 0 8px 24px rgba(0,0,0,0.05)',
  },
  signupRow: {
    textAlign: 'center',
    marginTop: 4,
  },
}
