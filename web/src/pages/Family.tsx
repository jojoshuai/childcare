// web/src/pages/Family.tsx
import { useState, useEffect } from 'react'
import { Card, List, Tag, Button, message, Skeleton } from 'antd'
import { getFamily, generateInvite } from '../api/family'
import type { FamilyInfo } from '../api/family'
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
