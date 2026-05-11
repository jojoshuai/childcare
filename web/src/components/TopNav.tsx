// web/src/components/TopNav.tsx
import { useAuth } from '../context/AuthContext'
import { Avatar, Popover, Button } from 'antd'
import { UserOutlined, LogoutOutlined } from '@ant-design/icons'

export default function TopNav() {
  const { user, logout } = useAuth()

  const displayName = user?.nickname || '用户'

  const content = (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      <div style={{ fontSize: 13, color: '#64748b', padding: '4px 0' }}>
        {displayName}
      </div>
      <Button size="small" type="text" danger icon={<LogoutOutlined />} onClick={logout} block>
        退出登录
      </Button>
    </div>
  )

  return (
    <div style={{
      height: 48,
      background: '#fff',
      borderBottom: '1px solid #e2e8f0',
      display: 'flex',
      alignItems: 'center',
      padding: '0 24px',
      flexShrink: 0,
    }}>
      <div style={{ fontSize: 15, fontWeight: 700, color: '#15803d', display: 'flex', alignItems: 'center', gap: 6 }}>
        <span style={{ fontSize: 18 }}>🌱</span>
        儿童成长
      </div>
      <div style={{ marginLeft: 'auto', display: 'flex', alignItems: 'center', gap: 12 }}>
        <Popover content={content} placement="bottomRight" trigger="click">
          <div style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8, padding: '4px 8px', borderRadius: 8 }}>
            <Avatar size={28} icon={<UserOutlined />} style={{ background: '#bbf7d0', color: '#15803d' }}>
              {displayName[0].toUpperCase()}
            </Avatar>
          </div>
        </Popover>
      </div>
    </div>
  )
}
