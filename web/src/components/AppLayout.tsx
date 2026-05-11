// web/src/components/AppLayout.tsx
import { Outlet } from 'react-router-dom'
import TopNav from './TopNav'

export default function AppLayout() {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
      <TopNav />
      <main style={{ flex: 1, background: '#f8fafc', overflowY: 'auto' }}>
        <Outlet />
      </main>
    </div>
  )
}
