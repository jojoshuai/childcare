// web/src/components/DietRecords.tsx
import { useState, useEffect, useCallback } from 'react'
import { DatePicker, Button, Popconfirm, message, Skeleton, Tag, Space } from 'antd'
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons'
import { listDiet, deleteDiet } from '../api/diet'
import type { DietRecord } from '../api/diet'
import DietForm from './DietForm'
import dayjs from 'dayjs'

const GREEN = '#16a34a'

const foodTypeLabels: Record<string, string> = {
  staple: '主食',
  vegetable: '蔬菜',
  fruit: '水果',
  protein: '肉蛋',
  dairy: '奶',
  snack: '零食',
}

const typeColors: Record<string, string> = {
  staple: 'gold',
  vegetable: 'green',
  fruit: 'magenta',
  protein: 'blue',
  dairy: 'cyan',
  snack: 'orange',
}

const amountLabels = ['', '少吃了一点', '正常量', '吃了很多']
const amountEmoji = ['', '🥄', '🍽️', '🍴']

interface Props {
  childId: string
}

export default function DietRecords({ childId }: Props) {
  const [records, setRecords] = useState<DietRecord[]>([])
  const [loading, setLoading] = useState(true)
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [selectedDate, setSelectedDate] = useState<dayjs.Dayjs>(dayjs())

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listDiet(childId)
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

  const handleDelete = async (did: string) => {
    try {
      await deleteDiet(childId, did)
      await loadData()
    } catch (err: any) {
      message.error(err.response?.data?.message ?? '删除失败')
    }
  }

  // Filter and group by time for the selected date
  const filtered = records.filter(r =>
    dayjs(r.record_time).isSame(selectedDate, 'day'),
  ).sort((a, b) => dayjs(b.record_time).valueOf() - dayjs(a.record_time).valueOf())

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
          onClick={() => setDrawerOpen(true)}
        >
          添加饮食
        </Button>
      </div>

      {loading ? (
        <Skeleton active paragraph={{ rows: 4 }} />
      ) : filtered.length === 0 ? (
        <div style={{ textAlign: 'center', padding: 40, color: '#9ca3af' }}>
          当天暂无饮食记录
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
              <span style={{ fontWeight: 600, color: '#15803d', marginRight: 8 }}>
                {record.food_name}
              </span>
              <Tag color={typeColors[record.food_type] || 'default'}>
                {foodTypeLabels[record.food_type] || record.food_type}
              </Tag>
              <span style={{ marginLeft: 8, fontSize: 13 }}>
                {amountEmoji[record.amount_level] || ''} {amountLabels[record.amount_level]}
              </span>
            </div>
            <Space>
              <span style={{ fontSize: 12, color: '#9ca3af' }}>
                {dayjs(record.record_time).format('HH:mm')}
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

      <DietForm
        open={drawerOpen}
        childId={childId}
        onClose={() => setDrawerOpen(false)}
        onSaved={loadData}
      />
    </div>
  )
}
