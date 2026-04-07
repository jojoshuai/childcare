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
