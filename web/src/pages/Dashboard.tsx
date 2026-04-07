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
} from '../api/children'
import type { Child } from '../api/children'
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
