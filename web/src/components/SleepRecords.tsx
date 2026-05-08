// web/src/components/SleepRecords.tsx
import { useState, useEffect, useCallback } from 'react'
import { DatePicker, Button, Popconfirm, message, Skeleton, Tag, Space } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import { listSleep, deleteSleep } from '../api/sleep'
import type { SleepRecord } from '../api/sleep'
import SleepForm from './SleepForm'
import dayjs from 'dayjs'

const GREEN = '#16a34a'

interface Props {
  childId: string
}

export default function SleepRecords({ childId }: Props) {
  const [records, setRecords] = useState<SleepRecord[]>([])
  const [loading, setLoading] = useState(true)
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [editing, setEditing] = useState<SleepRecord | null>(null)
  const [selectedDate, setSelectedDate] = useState<dayjs.Dayjs>(dayjs())

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listSleep(childId)
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

  const handleDelete = async (sid: string) => {
    try {
      await deleteSleep(childId, sid)
      await loadData()
    } catch (err: any) {
      message.error(err.response?.data?.message ?? '删除失败')
    }
  }

  // Filter records for the selected date
  const filtered = records.filter(r =>
    dayjs(r.start_time).isSame(selectedDate, 'day'),
  )

  const calcDuration = (start: string, end: string | null) => {
    if (!end) return '进行中'
    const diff = dayjs(end).diff(dayjs(start), 'minute')
    if (diff <= 0) return '--'
    const h = Math.floor(diff / 60)
    const m = diff % 60
    return `${h > 0 ? `${h}小时` : ''}${m}分钟`
  }

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
          添加睡眠
        </Button>
      </div>

      {loading ? (
        <Skeleton active paragraph={{ rows: 4 }} />
      ) : filtered.length === 0 ? (
        <div style={{ textAlign: 'center', padding: 40, color: '#9ca3af' }}>
          当天暂无睡眠记录
        </div>
      ) : (
        filtered.map(record => (
          <div
            key={record.id}
            style={{
              background: '#fff',
              borderRadius: 8,
              padding: 16,
              border: '1px solid #d1fae5',
              marginBottom: 12,
            }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <span style={{ fontSize: 16, fontWeight: 600, color: '#15803d' }}>
                  {dayjs(record.start_time).format('HH:mm')}
                  {' → '}
                  {record.end_time
                    ? dayjs(record.end_time).format('HH:mm')
                    : '进行中'}
                </span>
                <span style={{ marginLeft: 12, color: '#6b7280', fontSize: 13 }}>
                  {calcDuration(record.start_time, record.end_time)}
                </span>
              </div>
              <Space>
                {record.woke_up && <Tag color="orange">中途醒来 {record.wake_count > 0 ? `${record.wake_count}次` : ''}</Tag>}
                <EditOutlined
                  style={{ color: GREEN, cursor: 'pointer' }}
                  onClick={() => {
                    setEditing(record)
                    setDrawerOpen(true)
                  }}
                />
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
          </div>
        ))
      )}

      <SleepForm
        open={drawerOpen}
        childId={childId}
        editing={editing}
        onClose={() => setDrawerOpen(false)}
        onSaved={loadData}
      />
    </div>
  )
}
