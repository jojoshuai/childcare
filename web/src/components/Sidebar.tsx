// web/src/components/Sidebar.tsx
import { useNavigate, useLocation } from 'react-router-dom'
import { Button } from 'antd'
import {
  HomeOutlined,
  TeamOutlined,
  LogoutOutlined,
} from '@ant-design/icons'
import { useAuth } from '../context/AuthContext'

const navItems = [
  { path: '/dashboard', label: '首页', icon: <HomeOutlined /> },
  { path: '/family', label: '家庭', icon: <TeamOutlined /> },
]

export default function Sidebar() {
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuth()

  return (
    <div
      style={{
        width: 200,
        background: '#16a34a',
        minHeight: '100vh',
        display: 'flex',
        flexDirection: 'column',
        padding: '24px 0',
        flexShrink: 0,
      }}
    >
      <div
        style={{
          padding: '0 16px 32px',
          color: '#fff',
          fontSize: 18,
          fontWeight: 'bold',
        }}
      >
        🌱 儿童成长
      </div>

      <nav style={{ flex: 1 }}>
        {navItems.map(item => {
          const active = location.pathname.startsWith(item.path)
          return (
            <div
              key={item.path}
              onClick={() => navigate(item.path)}
              style={{
                padding: '12px 16px',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                color: active ? '#fff' : 'rgba(255,255,255,0.75)',
                background: active ? 'rgba(255,255,255,0.15)' : 'transparent',
                borderRadius: '0 20px 20px 0',
                marginRight: 8,
                fontSize: 14,
              }}
            >
              {item.icon} {item.label}
            </div>
          )
        })}
      </nav>

      <div style={{ padding: '0 16px' }}>
        <div
          style={{
            color: 'rgba(255,255,255,0.8)',
            fontSize: 12,
            marginBottom: 8,
          }}
        >
          {user?.nickname}
        </div>
        <Button
          size="small"
          type="text"
          icon={<LogoutOutlined />}
          style={{ color: 'rgba(255,255,255,0.7)', padding: 0 }}
          onClick={logout}
        >
          退出
        </Button>
      </div>
    </div>
  )
}
