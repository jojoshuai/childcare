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
import { listChildren } from '../api/children'
import type { Child } from '../api/children'
import {
  listMeasurements,
  deleteMeasurement,
} from '../api/measurements'
import type { Measurement } from '../api/measurements'
import { getWHOStandards } from '../api/who'
import type { WHOPoint } from '../api/who'
import MeasurementDrawer from '../components/MeasurementDrawer'
import GrowthChart from '../components/GrowthChart'
import SleepRecords from '../components/SleepRecords'
import DietRecords from '../components/DietRecords'
import SupplementRecords from '../components/SupplementRecords'
import dayjs from 'dayjs'

type MeasureType = 'weight' | 'height' | 'head_circumference'
type TabKey = MeasureType | 'sleep' | 'diet' | 'supplement'

const tabs: { key: TabKey; label: string }[] = [
  { key: 'weight', label: '体重' },
  { key: 'height', label: '身高' },
  { key: 'head_circumference', label: '头围' },
  { key: 'sleep', label: '睡眠' },
  { key: 'diet', label: '饮食' },
  { key: 'supplement', label: '补剂' },
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
  const [activeTab, setActiveTab] = useState<TabKey>('weight')
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
        listMeasurements(id, activeTab as MeasureType),
        getWHOStandards(child.gender, activeTab as MeasureType),
      ])
      setMeasurements(ms)
      setWhoData(who)
    } catch {
      message.error('加载失败，请刷新重试')
    } finally {
      setLoading(false)
    }
  }, [id, child, activeTab])

  useEffect(() => {
    if (child && ['weight', 'height', 'head_circumference'].includes(activeTab)) {
      loadMeasurements()
    }
  }, [child, activeTab, loadMeasurements])

  const handleDelete = async (mid: string) => {
    try {
      await deleteMeasurement(id!, mid)
      await loadMeasurements()
    } catch (err: any) {
      message.error(err.response?.data?.message ?? '删除失败')
    }
  }

  const isMeasureTab = ['weight', 'height', 'head_circumference'].includes(activeTab)

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
      render: (v: number) => `${v} ${activeTab === 'weight' ? 'kg' : 'cm'}`,
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
        activeKey={activeTab}
        onChange={k => setActiveTab(k as TabKey)}
        items={tabs.map(t => ({ key: t.key, label: t.label }))}
        style={{ marginBottom: 16 }}
      />

      {loading && isMeasureTab ? (
        <Skeleton active paragraph={{ rows: 8 }} />
      ) : isMeasureTab ? (
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
              type={activeTab as MeasureType}
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
      ) : (
        /* 睡眠/饮食/补剂 tabs */
        <div
          style={{
            background: '#fff',
            borderRadius: 8,
            padding: 16,
            border: '1px solid #d1fae5',
          }}
        >
          {activeTab === 'sleep' && <SleepRecords childId={id!} />}
          {activeTab === 'diet' && <DietRecords childId={id!} />}
          {activeTab === 'supplement' && <SupplementRecords childId={id!} />}
        </div>
      )}

      <MeasurementDrawer
        open={drawerOpen}
        childId={id!}
        type={activeTab as MeasureType}
        editing={editing}
        onClose={() => setDrawerOpen(false)}
        onSaved={loadMeasurements}
      />
    </div>
  )
}
