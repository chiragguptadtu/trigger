import { Link } from 'react-router-dom'
import { Typography } from 'antd'
import Navbar from '../components/Navbar'

const { Title, Text } = Typography

export default function SignupPage() {
  return (
    <div style={styles.page}>
      <Navbar />
      <div style={styles.body}>
        <div style={styles.card}>
          <Title level={4} style={styles.heading}>Sign up</Title>
          <Text style={styles.message}>
            Account creation is managed by your administrator.
          </Text>
          <Text style={styles.message}>
            Ask your admin to create an account for you, then come back to{' '}
            <Link to="/login">sign in</Link>.
          </Text>
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
    display: 'flex',
    flexDirection: 'column',
    gap: 12,
  },
  heading: {
    margin: 0,
    fontWeight: 600,
  },
  message: {
    fontSize: 14,
    color: 'rgba(0,0,0,0.65)',
  },
}
