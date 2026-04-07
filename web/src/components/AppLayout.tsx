// web/src/components/AppLayout.tsx
import { Outlet } from 'react-router-dom'
import Sidebar from './Sidebar'

export default function AppLayout() {
  return (
    <div style={{ display: 'flex', minHeight: '100vh' }}>
      <Sidebar />
      <main
        style={{
          flex: 1,
          background: '#f0fdf4',
          padding: 24,
          overflowY: 'auto',
        }}
      >
        <Outlet />
      </main>
    </div>
  )
}
