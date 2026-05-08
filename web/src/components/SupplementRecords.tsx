// web/src/components/SupplementRecords.tsx
import { useState, useEffect, useCallback } from 'react'
import { DatePicker, Button, Popconfirm, message, Skeleton, Tag, Space } from 'antd'
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons'
import { listSupplements, deleteSupplement } from '../api/supplement'
import type { SupplementRecord } from '../api/supplement'
import SupplementForm from './SupplementForm'
import dayjs from 'dayjs'

const GREEN = '#16a34a'

interface Props {
  childId: string
}

export default function SupplementRecords({ childId }: Props) {
  const [records, setRecords] = useState<SupplementRecord[]>([])
  const [loading, setLoading] = useState(true)
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [editing, setEditing] = useState<SupplementRecord | null>(null)
  const [selectedDate, setSelectedDate] = useState<dayjs.Dayjs>(dayjs())

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listSupplements(childId)
      setRecords(data)
    } catch {
      message.error('加载失败')
    } finally {
      setLoading(false)
    }
  }, [childId])

  useEffect(() => {
    loadData()
  }, [loadData])

  const handleDelete = async (spid: string) => {
    try {
      await deleteSupplement(childId, spid)
      await loadData()
    } catch (err: any) {
      message.error(err.response?.data?.message ?? '删除失败')
    }
  }

  // Filter records for the selected date
  const filtered = records.filter(r =>
    dayjs(r.taken_at).isSame(selectedDate, 'day'),
  ).sort((a, b) => dayjs(b.taken_at).valueOf() - dayjs(a.taken_at).valueOf())

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <DatePicker
          value={selectedDate}
          onChange={v => v && setSelectedDate(v)}
          format="YYYY-MM-DD"
        />
        <Button
          type="primary"
          icon={<PlusOutlined />}
          style={{ background: GREEN }}
          onClick={() => {
            setEditing(null)
            setDrawerOpen(true)
          }}
        >
          添加补剂
        </Button>
      </div>

      {loading ? (
        <Skeleton active paragraph={{ rows: 4 }} />
      ) : filtered.length === 0 ? (
        <div style={{ textAlign: 'center', padding: 40, color: '#9ca3af' }}>
          当天暂无补剂记录
        </div>
      ) : (
        filtered.map(record => (
          <div
            key={record.id}
            style={{
              background: '#fff',
              borderRadius: 8,
              padding: 12,
              border: '1px solid #d1fae5',
              marginBottom: 8,
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}
          >
            <div>
              <Tag color="blue" style={{ marginRight: 8 }}>{record.supplement_name}</Tag>
              {record.dose && (
                <span style={{ fontSize: 13, color: '#6b7280', marginRight: 8 }}>
                  {record.dose}
                </span>
              )}
            </div>
            <Space>
              <span style={{ fontSize: 12, color: '#9ca3af' }}>
                {dayjs(record.taken_at).format('HH:mm')}
              </span>
              <Popconfirm
                title="确认删除？"
                onConfirm={() => handleDelete(record.id)}
                okText="删除"
                cancelText="取消"
                okButtonProps={{ danger: true }}
              >
                <DeleteOutlined style={{ color: '#ef4444', cursor: 'pointer' }} />
              </Popconfirm>
            </Space>
          </div>
        ))
      )}

      <SupplementForm
        open={drawerOpen}
        childId={childId}
        editing={editing}
        onClose={() => setDrawerOpen(false)}
        onSaved={loadData}
      />
    </div>
  )
}
