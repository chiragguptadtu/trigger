import { useNavigate } from 'react-router-dom'
import { Typography } from 'antd'
import { PRIMARY } from '../utils/theme'

const { Text } = Typography

// Gun icon — a clean pistol silhouette pointing right
function GunIcon() {
  return (
    <svg viewBox="0 0 22 14" fill="white" xmlns="http://www.w3.org/2000/svg" width="17" height="11">
      {/* Barrel */}
      <rect x="11" y="3" width="11" height="3.5" rx="1" />
      {/* Body / slide */}
      <rect x="1" y="1" width="12" height="9" rx="1.5" />
      {/* Handle / grip */}
      <rect x="2.5" y="9" width="5.5" height="5" rx="1" />
    </svg>
  )
}

export default function Brand() {
  const navigate = useNavigate()
  return (
    <div style={styles.brand} onClick={() => navigate('/')} role="button">
      <div style={styles.mark}>
        <GunIcon />
      </div>
      <Text strong style={styles.name}>Trigger</Text>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  brand: {
    display: 'flex',
    alignItems: 'center',
    gap: 10,
    cursor: 'pointer',
  },
  mark: {
    width: 32,
    height: 32,
    borderRadius: 8,
    background: PRIMARY,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
  },
  name: {
    fontSize: 16,
    color: 'rgba(0,0,0,0.88)',
    letterSpacing: '-0.2px',
  },
}
