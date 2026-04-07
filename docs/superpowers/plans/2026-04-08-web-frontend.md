# Web 前端实现计划

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现完整的 React Web 前端，包含登录、孩子管理、生长曲线图、家庭管理四个模块。

**Architecture:** Vite + React 18 + TypeScript，左侧绿色导航栏布局，Axios 统一管理 API 调用（含 token 自动刷新），页面数据用 useState + useEffect 管理，AuthContext 存用户状态。

**Tech Stack:** React 18, TypeScript, Vite, Ant Design 5, Recharts, React Router v6, Axios, dayjs

---

## Chunk 1: 脚手架 + API 层 + AuthContext

### Task 1: 初始化项目

**Files:**
- Create: `web/` (整个目录)

- [ ] **Step 1: 在 childcare/ 下创建 Vite 项目**

```bash
cd /Users/zyb/Own/code/jojoshuai/childcare
npm create vite@latest web -- --template react-ts
cd web
npm install
```

- [ ] **Step 2: 安装依赖**

```bash
npm install antd @ant-design/icons recharts react-router-dom axios dayjs
```

- [ ] **Step 3: 清理默认文件**

删除 `src/App.css`、`src/index.css`（或清空内容），删除 `src/assets/` 目录，将 `src/App.tsx` 内容替换为空占位（后续步骤会完整写入）。

- [ ] **Step 4: 配置 Vite 代理（开发用）**

编辑 `web/vite.config.ts`：

```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
```

- [ ] **Step 5: 创建目录结构**

```bash
mkdir -p src/api src/context src/components src/pages
```

- [ ] **Step 6: 确认能启动**

```bash
npm run dev
```

预期：浏览器打开 http://localhost:5173 显示默认页面。

---

### Task 2: Axios 实例 + 拦截器

**Files:**
- Create: `web/src/api/axios.ts`

- [ ] **Step 1: 创建 axios.ts**

```typescript
// web/src/api/axios.ts
import axios from 'axios'

const api = axios.create({ baseURL: '' })

// 自动附加 token
api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// 401 自动刷新 token
let refreshing = false
let queue: Array<(token: string) => void> = []

api.interceptors.response.use(
  res => res,
  async err => {
    const orig = err.config
    if (err.response?.status !== 401 || orig._retry) {
      return Promise.reject(err)
    }
    orig._retry = true

    if (refreshing) {
      return new Promise(resolve => {
        queue.push(token => {
          orig.headers.Authorization = `Bearer ${token}`
          resolve(api(orig))
        })
      })
    }

    refreshing = true
    try {
      const refreshToken = localStorage.getItem('refresh_token')
      if (!refreshToken) throw new Error('no refresh token')
      const { data } = await axios.post('/api/auth/refresh', {
        refresh_token: refreshToken,
      })
      localStorage.setItem('token', data.token)
      localStorage.setItem('refresh_token', data.refresh_token)
      queue.forEach(cb => cb(data.token))
      queue = []
      orig.headers.Authorization = `Bearer ${data.token}`
      return api(orig)
    } catch {
      localStorage.clear()
      window.location.href = '/login'
      return Promise.reject(err)
    } finally {
      refreshing = false
    }
  },
)

export default api
```

---

### Task 3: API 模块

**Files:**
- Create: `web/src/api/auth.ts`
- Create: `web/src/api/children.ts`
- Create: `web/src/api/measurements.ts`
- Create: `web/src/api/family.ts`
- Create: `web/src/api/who.ts`

- [ ] **Step 1: 创建 auth.ts**

```typescript
// web/src/api/auth.ts
import api from './axios'

export interface AuthUser {
  id: string
  nickname: string
  family_id: string | null
  role: string | null
}

export interface AuthResponse {
  token: string
  refresh_token: string
  user: AuthUser
}

export const login = (data: { username: string; password: string }) =>
  api.post<AuthResponse>('/api/auth/login', data).then(r => r.data)

export const register = (data: {
  username: string
  password: string
  family_name: string
  nickname: string
}) => api.post<AuthResponse>('/api/auth/register', data).then(r => r.data)
```

- [ ] **Step 2: 创建 children.ts**

```typescript
// web/src/api/children.ts
import api from './axios'

export interface Child {
  id: string
  family_id: string
  name: string
  gender: 'male' | 'female'
  birth_date: string
  created_at: string
}

export const listChildren = () =>
  api.get<Child[]>('/api/children').then(r => r.data)

export const createChild = (data: {
  name: string
  gender: string
  birth_date: string
}) => api.post<Child>('/api/children', data).then(r => r.data)

export const deleteChild = (id: string) =>
  api.delete(`/api/children/${id}`)
```

- [ ] **Step 3: 创建 measurements.ts**

```typescript
// web/src/api/measurements.ts
import api from './axios'

export interface Measurement {
  id: string
  child_id: string
  type: 'weight' | 'height' | 'head_circumference'
  value: number
  measured_at: string
  note: string | null
  created_by: string
  created_at: string
}

export const listMeasurements = (childId: string, type?: string) =>
  api
    .get<Measurement[]>(`/api/children/${childId}/measurements`, {
      params: type ? { type } : {},
    })
    .then(r => r.data)

export const createMeasurement = (
  childId: string,
  data: {
    type: string
    value: number
    measured_at: string
    note: string | null
  },
) =>
  api
    .post<Measurement>(`/api/children/${childId}/measurements`, data)
    .then(r => r.data)

export const updateMeasurement = (
  childId: string,
  mid: string,
  data: {
    type: string
    value: number
    measured_at: string
    note: string | null
  },
) =>
  api
    .put<Measurement>(`/api/children/${childId}/measurements/${mid}`, data)
    .then(r => r.data)

export const deleteMeasurement = (childId: string, mid: string) =>
  api.delete(`/api/children/${childId}/measurements/${mid}`)
```

- [ ] **Step 4: 创建 family.ts**

```typescript
// web/src/api/family.ts
import api from './axios'

export interface FamilyMember {
  id: string
  nickname: string
  role: string
}

export interface FamilyInfo {
  id: string
  name: string
  members: FamilyMember[]
}

export const getFamily = () =>
  api.get<FamilyInfo>('/api/family').then(r => r.data)

export const generateInvite = () =>
  api
    .post<{ code: string; expires_at: string }>('/api/family/invite')
    .then(r => r.data)
```

- [ ] **Step 5: 创建 who.ts**

```typescript
// web/src/api/who.ts
import api from './axios'

export interface WHOPoint {
  month: number
  p3: number
  p50: number
  p97: number
}

export const getWHOStandards = (gender: string, type: string) =>
  api
    .get<{ data: WHOPoint[] }>('/api/who-standards', {
      params: { gender, type },
    })
    .then(r => r.data.data)
```

---

### Task 4: AuthContext

**Files:**
- Create: `web/src/context/AuthContext.tsx`

- [ ] **Step 1: 创建 AuthContext.tsx**

```typescript
// web/src/context/AuthContext.tsx
import {
  createContext,
  useContext,
  useState,
  ReactNode,
} from 'react'
import { AuthUser } from '../api/auth'

interface AuthContextType {
  user: AuthUser | null
  token: string | null
  login: (token: string, refreshToken: string, user: AuthUser) => void
  logout: () => void
}

const AuthContext = createContext<AuthContextType | null>(null)

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<AuthUser | null>(() => {
    try {
      const s = localStorage.getItem('user')
      return s ? JSON.parse(s) : null
    } catch {
      return null
    }
  })
  const [token, setToken] = useState<string | null>(() =>
    localStorage.getItem('token'),
  )

  const login = (token: string, refreshToken: string, user: AuthUser) => {
    localStorage.setItem('token', token)
    localStorage.setItem('refresh_token', refreshToken)
    localStorage.setItem('user', JSON.stringify(user))
    setToken(token)
    setUser(user)
  }

  const logout = () => {
    localStorage.clear()
    setToken(null)
    setUser(null)
    window.location.href = '/login'
  }

  return (
    <AuthContext.Provider value={{ user, token, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
```

- [ ] **Step 2: 验证编译**

```bash
cd /Users/zyb/Own/code/jojoshuai/childcare/web && npm run build 2>&1 | tail -5
```

预期：无 TypeScript 错误（可能有 unused import 警告，忽略）。

- [ ] **Step 3: Commit**

```bash
cd /Users/zyb/Own/code/jojoshuai/childcare
git add web/
git commit -m "feat(web): scaffold project + API layer + AuthContext"
```

---

## Chunk 2: 布局 + 路由 + 登录页

### Task 5: App.tsx + AppLayout + Sidebar

**Files:**
- Create: `web/src/components/AppLayout.tsx`
- Create: `web/src/components/Sidebar.tsx`
- Modify: `web/src/App.tsx`
- Modify: `web/src/main.tsx`

- [ ] **Step 1: 更新 main.tsx**

```typescript
// web/src/main.tsx
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
```

- [ ] **Step 2: 创建 Sidebar.tsx**

```typescript
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
```

- [ ] **Step 3: 创建 AppLayout.tsx**

```typescript
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
```

- [ ] **Step 4: 创建 App.tsx（路由）**

```typescript
// web/src/App.tsx
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider, useAuth } from './context/AuthContext'
import AppLayout from './components/AppLayout'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import ChildDetail from './pages/ChildDetail'
import Family from './pages/Family'

const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const { token } = useAuth()
  return token ? <>{children}</> : <Navigate to="/login" replace />
}

const AppRoutes = () => (
  <Routes>
    <Route path="/login" element={<Login />} />
    <Route
      path="/"
      element={
        <ProtectedRoute>
          <AppLayout />
        </ProtectedRoute>
      }
    >
      <Route index element={<Navigate to="/dashboard" replace />} />
      <Route path="dashboard" element={<Dashboard />} />
      <Route path="children/:id" element={<ChildDetail />} />
      <Route path="family" element={<Family />} />
    </Route>
  </Routes>
)

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <AppRoutes />
      </BrowserRouter>
    </AuthProvider>
  )
}
```

- [ ] **Step 5: 先创建四个空白页面占位（避免编译报错，各自创建为独立文件）**

```typescript
// web/src/pages/Login.tsx
export default function Login() { return <div>Login</div> }
```

```typescript
// web/src/pages/Dashboard.tsx
export default function Dashboard() { return <div>Dashboard</div> }
```

```typescript
// web/src/pages/ChildDetail.tsx
export default function ChildDetail() { return <div>ChildDetail</div> }
```

```typescript
// web/src/pages/Family.tsx
export default function Family() { return <div>Family</div> }
```

- [ ] **Step 6: 验证编译**

```bash
npm run build 2>&1 | tail -5
```

预期：Build succeeded，无报错。

---

### Task 6: 登录页

**Files:**
- Modify: `web/src/pages/Login.tsx`

- [ ] **Step 1: 实现 Login.tsx**

```typescript
// web/src/pages/Login.tsx
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, Tabs, Form, Input, Button, message } from 'antd'
import { login, register } from '../api/auth'
import { useAuth } from '../context/AuthContext'

const GREEN = '#16a34a'

export default function Login() {
  const navigate = useNavigate()
  const { login: authLogin } = useAuth()
  const [loading, setLoading] = useState(false)

  const handleLogin = async (values: {
    username: string
    password: string
  }) => {
    setLoading(true)
    try {
      const data = await login(values)
      authLogin(data.token, data.refresh_token, data.user)
      navigate('/dashboard')
    } catch (err: any) {
      message.error(err.response?.data?.message ?? '登录失败')
    } finally {
      setLoading(false)
    }
  }

  const handleRegister = async (values: {
    username: string
    password: string
    family_name: string
    nickname: string
  }) => {
    setLoading(true)
    try {
      const data = await register(values)
      authLogin(data.token, data.refresh_token, data.user)
      navigate('/dashboard')
    } catch (err: any) {
      message.error(err.response?.data?.message ?? '注册失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div
      style={{
        minHeight: '100vh',
        background: '#f0fdf4',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <Card
        style={{ width: 400 }}
        title={
          <span style={{ color: GREEN, fontSize: 20 }}>
            🌱 儿童成长记录
          </span>
        }
      >
        <Tabs
          items={[
            {
              key: 'login',
              label: '登录',
              children: (
                <Form layout="vertical" onFinish={handleLogin}>
                  <Form.Item
                    name="username"
                    label="用户名"
                    rules={[{ required: true }]}
                  >
                    <Input />
                  </Form.Item>
                  <Form.Item
                    name="password"
                    label="密码"
                    rules={[{ required: true }]}
                  >
                    <Input.Password />
                  </Form.Item>
                  <Button
                    type="primary"
                    htmlType="submit"
                    block
                    loading={loading}
                    style={{ background: GREEN }}
                  >
                    登录
                  </Button>
                </Form>
              ),
            },
            {
              key: 'register',
              label: '注册',
              children: (
                <Form layout="vertical" onFinish={handleRegister}>
                  <Form.Item
                    name="username"
                    label="用户名"
                    rules={[{ required: true, min: 3 }]}
                  >
                    <Input />
                  </Form.Item>
                  <Form.Item
                    name="password"
                    label="密码"
                    rules={[{ required: true, min: 6 }]}
                  >
                    <Input.Password />
                  </Form.Item>
                  <Form.Item
                    name="family_name"
                    label="家庭名称"
                    rules={[{ required: true }]}
                  >
                    <Input placeholder="如：王家" />
                  </Form.Item>
                  <Form.Item
                    name="nickname"
                    label="昵称"
                    rules={[{ required: true }]}
                  >
                    <Input placeholder="如：爸爸" />
                  </Form.Item>
                  <Button
                    type="primary"
                    htmlType="submit"
                    block
                    loading={loading}
                    style={{ background: GREEN }}
                  >
                    注册
                  </Button>
                </Form>
              ),
            },
          ]}
        />
      </Card>
    </div>
  )
}
```

- [ ] **Step 2: 验证编译**

```bash
npm run build 2>&1 | tail -5
```

- [ ] **Step 3: Commit**

```bash
cd /Users/zyb/Own/code/jojoshuai/childcare
git add web/
git commit -m "feat(web): add layout, routing, and login page"
```

---

## Chunk 3: Dashboard 首页

### Task 7: Dashboard 页面

**Files:**
- Modify: `web/src/pages/Dashboard.tsx`

- [ ] **Step 1: 实现 Dashboard.tsx**

```typescript
// web/src/pages/Dashboard.tsx
import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  List,
  Button,
  Modal,
  Form,
  Input,
  Select,
  DatePicker,
  message,
  Skeleton,
  Popconfirm,
} from 'antd'
import {
  PlusOutlined,
  DeleteOutlined,
  RightOutlined,
} from '@ant-design/icons'
import {
  listChildren,
  createChild,
  deleteChild,
  Child,
} from '../api/children'
import { useAuth } from '../context/AuthContext'
import dayjs from 'dayjs'

const GREEN = '#16a34a'

function ageLabel(birthDate: string): string {
  const months = Math.floor(
    dayjs().diff(dayjs(birthDate), 'day') / 30.4375,
  )
  if (months < 12) return `${months}个月`
  const years = Math.floor(months / 12)
  const rem = months % 12
  return rem > 0 ? `${years}岁${rem}个月` : `${years}岁`
}

export default function Dashboard() {
  const navigate = useNavigate()
  const { user } = useAuth()
  const [children, setChildren] = useState<Child[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [form] = Form.useForm()

  const load = async () => {
    try {
      setChildren(await listChildren())
    } catch {
      message.error('加载失败，请刷新重试')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const handleCreate = async (values: {
    name: string
    gender: string
    birth_date: dayjs.Dayjs
  }) => {
    setSubmitting(true)
    try {
      await createChild({
        name: values.name,
        gender: values.gender,
        birth_date: values.birth_date.format('YYYY-MM-DD'),
      })
      setModalOpen(false)
      form.resetFields()
      await load()
    } catch (err: any) {
      message.error(err.response?.data?.message ?? '创建失败')
    } finally {
      setSubmitting(false)
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await deleteChild(id)
      await load()
    } catch (err: any) {
      message.error(err.response?.data?.message ?? '删除失败')
    }
  }

  if (loading) return <Skeleton active paragraph={{ rows: 4 }} />

  return (
    <div style={{ maxWidth: 600 }}>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 16,
        }}
      >
        <h2 style={{ margin: 0, color: '#15803d' }}>我的孩子们</h2>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          style={{ background: GREEN }}
          onClick={() => setModalOpen(true)}
        >
          添加孩子
        </Button>
      </div>

      <List
        bordered
        style={{
          background: '#fff',
          borderColor: '#d1fae5',
          borderRadius: 8,
        }}
        dataSource={children}
        locale={{ emptyText: '还没有孩子，点击右上角添加' }}
        renderItem={child => (
          // 注意：spec 要求显示"最近一次测量摘要"，MVP 阶段暂不实现（需额外 API 调用），
          // 仅显示姓名和月龄，后续可扩展。
          <List.Item
            style={{ cursor: 'pointer', padding: '12px 16px' }}
            onClick={() => navigate(`/children/${child.id}`)}
            actions={[
              user?.role === 'owner' ? (
                <Popconfirm
                  key="delete"
                  title="确认删除？将同时删除所有测量记录。"
                  onConfirm={e => {
                    e?.stopPropagation()
                    handleDelete(child.id)
                  }}
                  onCancel={e => e?.stopPropagation()}
                  okText="删除"
                  cancelText="取消"
                  okButtonProps={{ danger: true }}
                >
                  <DeleteOutlined
                    style={{ color: '#ef4444' }}
                    onClick={e => e.stopPropagation()}
                  />
                </Popconfirm>
              ) : null,
              <RightOutlined key="go" style={{ color: GREEN }} />,
            ].filter(Boolean)}
          >
            <List.Item.Meta
              avatar={
                <span style={{ fontSize: 28 }}>
                  {child.gender === 'male' ? '👦' : '👧'}
                </span>
              }
              title={
                <span style={{ color: '#15803d', fontWeight: 600 }}>
                  {child.name}
                </span>
              }
              description={
                <span style={{ color: '#6b7280' }}>
                  {ageLabel(child.birth_date)}
                </span>
              }
            />
          </List.Item>
        )}
      />

      <Modal
        title="添加孩子"
        open={modalOpen}
        onCancel={() => {
          setModalOpen(false)
          form.resetFields()
        }}
        footer={null}
      >
        <Form form={form} layout="vertical" onFinish={handleCreate}>
          <Form.Item
            name="name"
            label="姓名"
            rules={[{ required: true }]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            name="gender"
            label="性别"
            rules={[{ required: true }]}
          >
            <Select
              options={[
                { value: 'male', label: '男孩 👦' },
                { value: 'female', label: '女孩 👧' },
              ]}
            />
          </Form.Item>
          <Form.Item
            name="birth_date"
            label="出生日期"
            rules={[{ required: true }]}
          >
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>
          <Button
            type="primary"
            htmlType="submit"
            block
            loading={submitting}
            style={{ background: GREEN }}
          >
            保存
          </Button>
        </Form>
      </Modal>
    </div>
  )
}
```

- [ ] **Step 2: 验证编译**

```bash
npm run build 2>&1 | tail -5
```

- [ ] **Step 3: Commit**

```bash
cd /Users/zyb/Own/code/jojoshuai/childcare
git add web/src/pages/Dashboard.tsx
git commit -m "feat(web): add dashboard page with child list"
```

---

## Chunk 4: 孩子详情页（图表 + 抽屉）

### Task 8: GrowthChart 组件

**Files:**
- Create: `web/src/components/GrowthChart.tsx`

图表策略：将孩子数据和 WHO 数据按月龄合并为单一数组传入 `<LineChart data={...}>`，每条线用独立 `dataKey`。

- [ ] **Step 1: 创建 GrowthChart.tsx**

```typescript
// web/src/components/GrowthChart.tsx
import { useMemo } from 'react'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'
import { Measurement } from '../api/measurements'
import { WHOPoint } from '../api/who'
import dayjs from 'dayjs'

interface Props {
  measurements: Measurement[]
  whoData: WHOPoint[]
  birthDate: string
  type: string
}

function calcAgeMonths(birthDate: string, measuredAt: string): number {
  const diffDays = dayjs(measuredAt).diff(dayjs(birthDate), 'day')
  return Math.floor(diffDays / 30.4375)
}

const typeUnit: Record<string, string> = {
  weight: 'kg',
  height: 'cm',
  head_circumference: 'cm',
}

interface ChartPoint {
  month: number
  value?: number
  date?: string
  p3?: number
  p50?: number
  p97?: number
}

export default function GrowthChart({
  measurements,
  whoData,
  birthDate,
  type,
}: Props) {
  const maxAgeMonths = useMemo(() => {
    if (!measurements.length) return 0
    return Math.max(
      ...measurements.map(m => calcAgeMonths(birthDate, m.measured_at)),
    )
  }, [measurements, birthDate])

  const showWHO = maxAgeMonths < 61

  // 合并孩子数据和 WHO 数据，按月龄索引
  const chartData: ChartPoint[] = useMemo(() => {
    const byMonth: Record<number, ChartPoint> = {}

    if (showWHO) {
      whoData.forEach(pt => {
        byMonth[pt.month] = { month: pt.month, p3: pt.p3, p50: pt.p50, p97: pt.p97 }
      })
    }

    measurements.forEach(m => {
      const month = calcAgeMonths(birthDate, m.measured_at)
      byMonth[month] = {
        ...(byMonth[month] ?? { month }),
        value: m.value,
        date: m.measured_at,
      }
    })

    return Object.values(byMonth).sort((a, b) => a.month - b.month)
  }, [measurements, whoData, birthDate, showWHO])

  const unit = typeUnit[type] ?? ''

  const customTooltip = ({ active, payload, label }: any) => {
    if (!active || !payload?.length) return null
    return (
      <div
        style={{
          background: '#fff',
          border: '1px solid #d1fae5',
          borderRadius: 6,
          padding: '8px 12px',
          fontSize: 12,
        }}
      >
        <div style={{ color: '#6b7280', marginBottom: 4 }}>
          月龄 {label} 个月
        </div>
        {payload.map((p: any) => (
          <div key={p.name} style={{ color: p.color }}>
            {p.name}: {p.value} {unit}
          </div>
        ))}
      </div>
    )
  }

  return (
    <div>
      <ResponsiveContainer width="100%" height={320}>
        <LineChart data={chartData} margin={{ top: 8, right: 16, left: 0, bottom: 8 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#d1fae5" />
          <XAxis
            dataKey="month"
            type="number"
            domain={['dataMin', 'dataMax']}
            label={{
              value: '月龄',
              position: 'insideBottomRight',
              offset: -8,
              fontSize: 11,
            }}
            tick={{ fontSize: 11 }}
          />
          <YAxis
            label={{
              value: unit,
              angle: -90,
              position: 'insideLeft',
              offset: 8,
              fontSize: 11,
            }}
            tick={{ fontSize: 11 }}
          />
          <Tooltip content={customTooltip} />
          <Legend />

          <Line
            dataKey="value"
            name="孩子数据"
            stroke="#16a34a"
            strokeWidth={2}
            dot={{ r: 4, fill: '#16a34a' }}
            type="monotone"
            connectNulls
          />

          {showWHO && (
            <>
              <Line
                dataKey="p97"
                name="WHO P97"
                stroke="#9ca3af"
                strokeWidth={1}
                strokeDasharray="4 4"
                dot={false}
                type="monotone"
                connectNulls
              />
              <Line
                dataKey="p50"
                name="WHO P50"
                stroke="#6b7280"
                strokeWidth={1}
                strokeDasharray="4 4"
                dot={false}
                type="monotone"
                connectNulls
              />
              <Line
                dataKey="p3"
                name="WHO P3"
                stroke="#9ca3af"
                strokeWidth={1}
                strokeDasharray="4 4"
                dot={false}
                type="monotone"
                connectNulls
              />
            </>
          )}
        </LineChart>
      </ResponsiveContainer>

      {!showWHO && (
        <div
          style={{
            textAlign: 'center',
            color: '#9ca3af',
            fontSize: 12,
            marginTop: 8,
          }}
        >
          WHO 参考数据覆盖范围为 0-60 个月
        </div>
      )}
    </div>
  )
}
```

---

### Task 9: MeasurementDrawer 组件

**Files:**
- Create: `web/src/components/MeasurementDrawer.tsx`

- [ ] **Step 1: 创建 MeasurementDrawer.tsx**

```typescript
// web/src/components/MeasurementDrawer.tsx
import { useEffect } from 'react'
import {
  Drawer,
  Form,
  InputNumber,
  DatePicker,
  Input,
  Button,
  message,
} from 'antd'
import {
  createMeasurement,
  updateMeasurement,
  Measurement,
} from '../api/measurements'
import dayjs from 'dayjs'

interface Props {
  open: boolean
  childId: string
  type: 'weight' | 'height' | 'head_circumference'
  editing?: Measurement | null
  onClose: () => void
  onSaved: () => void
}

const typeLabel: Record<string, string> = {
  weight: '体重 (kg)',
  height: '身高 (cm)',
  head_circumference: '头围 (cm)',
}

const typeRange: Record<string, [number, number]> = {
  weight: [0.5, 200],
  height: [20, 250],
  head_circumference: [20, 80],
}

export default function MeasurementDrawer({
  open,
  childId,
  type,
  editing,
  onClose,
  onSaved,
}: Props) {
  const [form] = Form.useForm()

  useEffect(() => {
    if (!open) return
    if (editing) {
      form.setFieldsValue({
        value: editing.value,
        measured_at: dayjs(editing.measured_at),
        note: editing.note ?? '',
      })
    } else {
      form.resetFields()
      form.setFieldsValue({ measured_at: dayjs() })
    }
  }, [open, editing])

  const handleSave = async (values: {
    value: number
    measured_at: dayjs.Dayjs
    note: string
  }) => {
    const payload = {
      type,
      value: values.value,
      measured_at: values.measured_at.format('YYYY-MM-DD'),
      note: values.note?.trim() || null,
    }
    try {
      if (editing) {
        await updateMeasurement(childId, editing.id, payload)
      } else {
        await createMeasurement(childId, payload)
      }
      onSaved()
      onClose()
    } catch (err: any) {
      message.error(err.response?.data?.message ?? '保存失败')
    }
  }

  const [min, max] = typeRange[type] ?? [0, 9999]

  return (
    <Drawer
      title={editing ? `编辑${typeLabel[type]}` : `添加${typeLabel[type]}`}
      placement="right"
      open={open}
      onClose={onClose}
      width={360}
    >
      <Form form={form} layout="vertical" onFinish={handleSave}>
        <Form.Item
          name="value"
          label={typeLabel[type]}
          rules={[
            { required: true, message: '请输入数值' },
            {
              type: 'number',
              min,
              max,
              message: `数值应在 ${min}–${max} 之间`,
            },
          ]}
        >
          <InputNumber style={{ width: '100%' }} step={0.1} />
        </Form.Item>
        <Form.Item
          name="measured_at"
          label="测量日期"
          rules={[{ required: true }]}
        >
          <DatePicker style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="note" label="备注（选填）">
          <Input.TextArea rows={3} />
        </Form.Item>
        <Button
          type="primary"
          htmlType="submit"
          block
          style={{ background: '#16a34a' }}
        >
          保存
        </Button>
      </Form>
    </Drawer>
  )
}
```

---

### Task 10: ChildDetail 页面

**Files:**
- Modify: `web/src/pages/ChildDetail.tsx`

- [ ] **Step 1: 实现 ChildDetail.tsx**

```typescript
// web/src/pages/ChildDetail.tsx
import { useState, useEffect, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import {
  Tabs,
  Table,
  Button,
  Popconfirm,
  message,
  Skeleton,
  Space,
} from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import { listChildren, Child } from '../api/children'
import {
  listMeasurements,
  deleteMeasurement,
  Measurement,
} from '../api/measurements'
import { getWHOStandards, WHOPoint } from '../api/who'
import MeasurementDrawer from '../components/MeasurementDrawer'
import GrowthChart from '../components/GrowthChart'
import dayjs from 'dayjs'

type MeasureType = 'weight' | 'height' | 'head_circumference'

const tabs: { key: MeasureType; label: string }[] = [
  { key: 'weight', label: '体重' },
  { key: 'height', label: '身高' },
  { key: 'head_circumference', label: '头围' },
]

function ageLabel(birthDate: string): string {
  const months = Math.floor(
    dayjs().diff(dayjs(birthDate), 'day') / 30.4375,
  )
  if (months < 12) return `${months}个月`
  const years = Math.floor(months / 12)
  const rem = months % 12
  return rem > 0 ? `${years}岁${rem}个月` : `${years}岁`
}

export default function ChildDetail() {
  const { id } = useParams<{ id: string }>()
  const [child, setChild] = useState<Child | null>(null)
  const [type, setType] = useState<MeasureType>('weight')
  const [measurements, setMeasurements] = useState<Measurement[]>([])
  const [whoData, setWhoData] = useState<WHOPoint[]>([])
  const [loading, setLoading] = useState(true)
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [editing, setEditing] = useState<Measurement | null>(null)

  useEffect(() => {
    listChildren().then(cs => {
      setChild(cs.find(c => c.id === id) ?? null)
    })
  }, [id])

  const loadMeasurements = useCallback(async () => {
    if (!id || !child) return
    setLoading(true)
    try {
      const [ms, who] = await Promise.all([
        listMeasurements(id, type),
        getWHOStandards(child.gender, type),
      ])
      setMeasurements(ms)
      setWhoData(who)
    } catch {
      message.error('加载失败，请刷新重试')
    } finally {
      setLoading(false)
    }
  }, [id, child, type])

  useEffect(() => {
    if (child) loadMeasurements()
  }, [child, type, loadMeasurements])

  const handleDelete = async (mid: string) => {
    try {
      await deleteMeasurement(id!, mid)
      await loadMeasurements()
    } catch (err: any) {
      message.error(err.response?.data?.message ?? '删除失败')
    }
  }

  const columns = [
    {
      title: '日期',
      dataIndex: 'measured_at',
      key: 'date',
      render: (v: string) => dayjs(v).format('YYYY-MM-DD'),
    },
    {
      title: '数值',
      dataIndex: 'value',
      key: 'value',
      render: (v: number) => `${v} ${type === 'weight' ? 'kg' : 'cm'}`,
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: unknown, record: Measurement) => (
        <Space>
          <EditOutlined
            style={{ color: '#16a34a', cursor: 'pointer' }}
            onClick={() => {
              setEditing(record)
              setDrawerOpen(true)
            }}
          />
          <Popconfirm
            title="确认删除这条记录？"
            onConfirm={() => handleDelete(record.id)}
            okText="删除"
            cancelText="取消"
            okButtonProps={{ danger: true }}
          >
            <DeleteOutlined style={{ color: '#ef4444', cursor: 'pointer' }} />
          </Popconfirm>
        </Space>
      ),
    },
  ]

  if (!child) return <Skeleton active paragraph={{ rows: 8 }} />

  return (
    <div>
      <h2 style={{ color: '#15803d', marginBottom: 4 }}>{child.name}</h2>
      <p style={{ color: '#6b7280', marginTop: 0, marginBottom: 16 }}>
        {ageLabel(child.birth_date)}
      </p>

      <Tabs
        activeKey={type}
        onChange={k => setType(k as MeasureType)}
        items={tabs.map(t => ({ key: t.key, label: t.label }))}
        style={{ marginBottom: 16 }}
      />

      {loading ? (
        <Skeleton active paragraph={{ rows: 8 }} />
      ) : (
        <div style={{ display: 'flex', gap: 24, alignItems: 'flex-start' }}>
          {/* 左：图表 */}
          <div
            style={{
              flex: 3,
              background: '#fff',
              borderRadius: 8,
              padding: 16,
              border: '1px solid #d1fae5',
            }}
          >
            <GrowthChart
              measurements={measurements}
              whoData={whoData}
              birthDate={child.birth_date}
              type={type}
            />
          </div>

          {/* 右：记录列表 */}
          <div
            style={{
              flex: 2,
              background: '#fff',
              borderRadius: 8,
              border: '1px solid #d1fae5',
              overflow: 'hidden',
            }}
          >
            <div
              style={{
                padding: '12px 16px',
                background: '#f0fdf4',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                borderBottom: '1px solid #d1fae5',
              }}
            >
              <span style={{ fontWeight: 600, color: '#15803d' }}>
                测量记录
              </span>
              <Button
                type="primary"
                size="small"
                icon={<PlusOutlined />}
                style={{ background: '#16a34a' }}
                onClick={() => {
                  setEditing(null)
                  setDrawerOpen(true)
                }}
              >
                添加
              </Button>
            </div>
            <Table
              dataSource={[...measurements].sort(
                (a, b) =>
                  dayjs(b.measured_at).valueOf() -
                  dayjs(a.measured_at).valueOf(),
              )}
              columns={columns}
              rowKey="id"
              size="small"
              pagination={false}
              scroll={{ y: 360 }}
            />
          </div>
        </div>
      )}

      <MeasurementDrawer
        open={drawerOpen}
        childId={id!}
        type={type}
        editing={editing}
        onClose={() => setDrawerOpen(false)}
        onSaved={loadMeasurements}
      />
    </div>
  )
}
```

- [ ] **Step 2: 验证编译**

```bash
npm run build 2>&1 | tail -5
```

- [ ] **Step 3: Commit**

```bash
cd /Users/zyb/Own/code/jojoshuai/childcare
git add web/src/components/ web/src/pages/ChildDetail.tsx
git commit -m "feat(web): add growth chart, measurement drawer, and child detail page"
```

---

## Chunk 5: 家庭页 + 最终验证

### Task 11: Family 页面

**Files:**
- Modify: `web/src/pages/Family.tsx`

- [ ] **Step 1: 实现 Family.tsx**

```typescript
// web/src/pages/Family.tsx
import { useState, useEffect } from 'react'
import { Card, List, Tag, Button, message, Skeleton } from 'antd'
import { getFamily, generateInvite, FamilyInfo } from '../api/family'
import { useAuth } from '../context/AuthContext'

const GREEN = '#16a34a'

export default function Family() {
  const { user } = useAuth()
  const [family, setFamily] = useState<FamilyInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [invite, setInvite] = useState<{
    code: string
    expires_at: string
  } | null>(null)
  const [generating, setGenerating] = useState(false)

  useEffect(() => {
    getFamily()
      .then(setFamily)
      .catch(() => message.error('加载失败，请刷新重试'))
      .finally(() => setLoading(false))
  }, [])

  const handleGenerate = async () => {
    setGenerating(true)
    try {
      setInvite(await generateInvite())
    } catch (err: any) {
      message.error(err.response?.data?.message ?? '生成失败')
    } finally {
      setGenerating(false)
    }
  }

  if (loading) return <Skeleton active paragraph={{ rows: 4 }} />

  return (
    <div style={{ maxWidth: 500 }}>
      <h2 style={{ color: '#15803d', marginBottom: 16 }}>{family?.name}</h2>

      <Card
        title="家庭成员"
        style={{ marginBottom: 16, borderColor: '#d1fae5' }}
      >
        <List
          dataSource={family?.members ?? []}
          renderItem={member => (
            <List.Item>
              <span style={{ fontWeight: 500 }}>{member.nickname}</span>
              <Tag
                color={member.role === 'owner' ? 'green' : 'default'}
                style={{ marginLeft: 8 }}
              >
                {member.role === 'owner' ? '创建者' : '成员'}
              </Tag>
            </List.Item>
          )}
        />
      </Card>

      {user?.role === 'owner' && (
        <Card title="邀请家人加入" style={{ borderColor: '#d1fae5' }}>
          <p style={{ color: '#6b7280', fontSize: 13, marginBottom: 12 }}>
            生成邀请码后，将 6 位码告诉家人，他们在小程序输入即可加入。
          </p>
          <Button
            type="primary"
            style={{ background: GREEN }}
            loading={generating}
            onClick={handleGenerate}
          >
            生成邀请码
          </Button>

          {invite && (
            <div
              style={{
                marginTop: 16,
                padding: 16,
                background: '#f0fdf4',
                borderRadius: 8,
                textAlign: 'center',
                border: '1px solid #d1fae5',
              }}
            >
              <div
                style={{
                  fontSize: 36,
                  fontWeight: 'bold',
                  letterSpacing: 10,
                  color: GREEN,
                  fontFamily: 'monospace',
                }}
              >
                {invite.code}
              </div>
              <div
                style={{ color: '#6b7280', fontSize: 12, marginTop: 8 }}
              >
                有效期至{' '}
                {new Date(invite.expires_at).toLocaleTimeString()}
              </div>
            </div>
          )}
        </Card>
      )}
    </div>
  )
}
```

---

### Task 12: 最终验证

- [ ] **Step 1: 完整编译检查**

```bash
cd /Users/zyb/Own/code/jojoshuai/childcare/web
npm run build 2>&1
```

预期：`✓ built in Xs`，无 TypeScript error。

- [ ] **Step 2: 启动后端（确保有 .env）**

```bash
cd /Users/zyb/Own/code/jojoshuai/childcare/backend
cat .env.example  # 参照格式创建 .env
# 需要设置 MYSQL_DSN、JWT_SECRET、JWT_REFRESH_SECRET
```

- [ ] **Step 3: 启动前端开发服务器**

```bash
cd /Users/zyb/Own/code/jojoshuai/childcare/web
npm run dev
```

打开 http://localhost:5173，手动验证：
- [ ] 注册账号 → 进入 dashboard
- [ ] 添加孩子 → 进入详情页
- [ ] 添加体重记录 → 图表出现数据点
- [ ] 切换身高/头围 Tab → 图表刷新
- [ ] 编辑/删除记录 → 列表更新
- [ ] 家庭页生成邀请码 → 显示 6 位码
- [ ] 退出登录 → 跳转 /login

- [ ] **Step 4: 最终 commit**

```bash
cd /Users/zyb/Own/code/jojoshuai/childcare
git add web/src/pages/Family.tsx
git commit -m "feat(web): add family page and complete web frontend MVP"
```
