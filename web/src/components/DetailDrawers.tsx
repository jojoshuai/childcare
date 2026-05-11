// web/src/components/DetailDrawers.tsx
import { useState, useEffect, useCallback } from 'react'
import { Button, Tag, DatePicker, message, Popconfirm } from 'antd'
import { DeleteOutlined, PlusOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import { listSleep, deleteSleep } from '../api/sleep'
import type { SleepRecord } from '../api/sleep'
import { listDiet, deleteDiet } from '../api/diet'
import type { DietRecord } from '../api/diet'
import { listSupplements, deleteSupplement } from '../api/supplement'
import type { SupplementRecord } from '../api/supplement'
import { listMeasurements, deleteMeasurement, createMeasurement } from '../api/measurements'
import type { Measurement } from '../api/measurements'
import SleepForm from './SleepForm'
import DietForm from './DietForm'
import SupplementForm from './SupplementForm'
import MeasurementDrawer from './MeasurementDrawer'

const GREEN = '#22c55e'
const mealTypeLabels: Record<string, string> = {
  breakfast: '早餐', lunch: '午餐', dinner: '晚餐', snack: '加餐',
}

// ── Shared Drawer Wrapper ──
function DrawerShell({
  open, onClose, title, children,
}: { open: boolean; onClose: () => void; title: string; children: React.ReactNode }) {
  return (
    <div style={{
      position: 'fixed', top: 0, right: 0, bottom: 0, width: 480, maxWidth: '90vw',
      background: '#fff', boxShadow: '-4px 0 24px rgba(0,0,0,.12)',
      zIndex: 1000, transform: open ? 'translateX(0)' : 'translateX(100%)',
      transition: 'transform .3s cubic-bezier(.4,0,.2,1)',
      display: 'flex', flexDirection: 'column', overflow: 'hidden',
    }}>
      <div style={{
        padding: '16px 20px', borderBottom: '1px solid #f1f5f9',
        display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexShrink: 0,
      }}>
        <div style={{ fontSize: 15, fontWeight: 700, color: '#0f172a' }}>{title}</div>
        <button onClick={onClose} style={{
          width: 28, height: 28, borderRadius: 8, border: 'none', background: 'transparent',
          cursor: 'pointer', fontSize: 18, color: '#94a3b8', display: 'flex',
          alignItems: 'center', justifyContent: 'center',
        }}>×</button>
      </div>
      <div style={{ flex: 1, overflowY: 'auto', padding: '16px 20px 20px' }}>
        {children}
      </div>
    </div>
  )
}

// ── Overlay ──
function Overlay({ show, onClick }: { show: boolean; onClick: () => void }) {
  if (!show) return null
  return (
    <div style={{
      position: 'fixed', inset: 0, background: 'rgba(0,0,0,.3)',
      backdropFilter: 'blur(3px)', zIndex: 999,
    }} onClick={onClick} />
  )
}

// ── Sleep Detail ──
export function SleepDetailDrawer({ open, onClose, childId }: { open: boolean; onClose: () => void; childId: string }) {
  const [records, setRecords] = useState<SleepRecord[]>([])
  const [selectedDate, setSelectedDate] = useState(dayjs())
  const [loading, setLoading] = useState(true)
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [editing, setEditing] = useState<SleepRecord | null>(null)

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listSleep(childId)
      setRecords(data)
    } catch { message.error('加载失败') }
    finally { setLoading(false) }
  }, [childId])

  useEffect(() => { if (open && childId) loadData() }, [open, childId, loadData])

  const handleDelete = async (sid: string) => {
    try { await deleteSleep(childId, sid); await loadData() }
    catch { message.error('删除失败') }
  }

  const filtered = records.filter(r => dayjs(r.start_time).isSame(selectedDate, 'day'))

  const calcDuration = (start: string, end: string | null) => {
    if (!end) return '进行中'
    const diff = dayjs(end).diff(dayjs(start), 'minute')
    if (diff <= 0) return '--'
    const h = Math.floor(diff / 60); const m = diff % 60
    return `${h > 0 ? `${h}小时` : ''}${m}分钟`
  }

  // Group by date
  const grouped: Record<string, SleepRecord[]> = {}
  filtered.forEach(r => {
    const key = dayjs(r.start_time).format('YYYY-MM-DD')
    if (!grouped[key]) grouped[key] = []
    grouped[key].push(r)
  })
  const sortedDates = Object.keys(grouped).sort((a, b) => b.localeCompare(a))

  return (
    <>
      <Overlay show={open} onClick={onClose} />
      <DrawerShell open={open} onClose={onClose} title="🌙 睡眠记录">
        <div style={{ marginBottom: 12 }}>
          <DatePicker value={selectedDate} onChange={v => v && setSelectedDate(v)} format="YYYY-MM-DD" />
        </div>
        <Button type="primary" icon={<PlusOutlined />} block style={{ background: GREEN, marginBottom: 16 }} onClick={() => { setEditing(null); setDrawerOpen(true) }}>
          添加睡眠记录
        </Button>

        {loading ? <div style={{ textAlign: 'center', padding: 40, color: '#94a3b8' }}>加载中...</div> : (
          sortedDates.length === 0
            ? <div style={{ textAlign: 'center', padding: 40, color: '#94a3b8' }}>暂无记录</div>
            : sortedDates.map(date => (
              <div key={date}>
                <div style={{
                  fontSize: 12, fontWeight: 700, color: '#475569',
                  margin: '16px 0 8px', paddingBottom: 6,
                  borderBottom: '1px solid #f1f5f9',
                }}>
                  {date === dayjs().format('YYYY-MM-DD') ? '今天' : date}
                </div>
                {grouped[date].map(record => (
                  <div key={record.id} style={{
                    display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                    padding: '10px 0', borderBottom: '1px solid #f8fafc',
                  }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                      <div style={{
                        width: 8, height: 8, borderRadius: '50%',
                        background: record.woke_up ? '#a78bfa' : '#8b5cf6',
                      }} />
                      <div>
                        <div style={{ fontSize: 13, fontWeight: 500, color: '#334155' }}>
                          {dayjs(record.start_time).format('HH:mm')} → {record.end_time ? dayjs(record.end_time).format('HH:mm') : '进行中'}
                        </div>
                        <div style={{ fontSize: 11, color: '#94a3b8', marginTop: 1 }}>
                          {calcDuration(record.start_time, record.end_time)}
                        </div>
                      </div>
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                      {record.woke_up && <Tag color="orange">醒 {record.wake_count}次</Tag>}
                      <Popconfirm title="确认删除？" onConfirm={() => handleDelete(record.id)} okText="删除" cancelText="取消" okButtonProps={{ danger: true }}>
                        <DeleteOutlined style={{ color: '#ef4444', cursor: 'pointer' }} />
                      </Popconfirm>
                    </div>
                  </div>
                ))}
              </div>
            ))
        )}

        <SleepForm open={drawerOpen} childId={childId} editing={editing} onClose={() => setDrawerOpen(false)} onSaved={loadData} />
      </DrawerShell>
    </>
  )
}

// ── Diet Detail ──
interface MealGroup {
  mealGroupId: string | null
  mealType: string
  recordTime: string
  items: DietRecord[]
}

export function DietDetailDrawer({ open, onClose, childId }: { open: boolean; onClose: () => void; childId: string }) {
  const [records, setRecords] = useState<DietRecord[]>([])
  const [selectedDate, setSelectedDate] = useState(dayjs())
  const [loading, setLoading] = useState(true)
  const [drawerOpen, setDrawerOpen] = useState(false)

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listDiet(childId)
      setRecords(data)
    } catch { message.error('加载失败') }
    finally { setLoading(false) }
  }, [childId])

  useEffect(() => { if (open && childId) loadData() }, [open, childId, loadData])

  const handleDelete = async (did: string) => {
    try { await deleteDiet(childId, did); await loadData() }
    catch { message.error('删除失败') }
  }

  const filtered = records.filter(r => dayjs(r.record_time).isSame(selectedDate, 'day'))
    .sort((a, b) => dayjs(b.record_time).valueOf() - dayjs(a.record_time).valueOf())

  const groups: MealGroup[] = []
  const seen = new Map<string, MealGroup>()
  for (const r of filtered) {
    const gid = r.meal_group_id
    if (gid && seen.has(gid)) { seen.get(gid)!.items.push(r) }
    else {
      const g: MealGroup = { mealGroupId: gid ?? null, mealType: r.meal_type || '', recordTime: r.record_time, items: [r] }
      groups.push(g)
      if (gid) seen.set(gid, g)
    }
  }

  const foodTypeLabels: Record<string, string> = { staple: '主食', vegetable: '蔬菜', fruit: '水果', protein: '肉蛋', dairy: '奶', snack: '零食' }
  const typeColors: Record<string, string> = { staple: 'gold', vegetable: 'green', fruit: 'magenta', protein: 'blue', dairy: 'cyan', snack: 'orange' }
  const amountEmoji = ['', '🥄', '🍽️', '🍴']

  return (
    <>
      <Overlay show={open} onClick={onClose} />
      <DrawerShell open={open} onClose={onClose} title="🥗 饮食记录">
        <div style={{ marginBottom: 12 }}>
          <DatePicker value={selectedDate} onChange={v => v && setSelectedDate(v)} format="YYYY-MM-DD" />
        </div>
        <Button type="primary" icon={<PlusOutlined />} block style={{ background: GREEN, marginBottom: 16 }} onClick={() => setDrawerOpen(true)}>
          添加饮食
        </Button>

        {loading ? <div style={{ textAlign: 'center', padding: 40, color: '#94a3b8' }}>加载中...</div> : (
          groups.length === 0
            ? <div style={{ textAlign: 'center', padding: 40, color: '#94a3b8' }}>暂无记录</div>
            : groups.map((group, gi) => (
              <div key={gi} style={{ marginBottom: 14 }}>
                <div style={{
                  padding: '8px 0', borderBottom: '1px solid #f1f5f9',
                  display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4,
                }}>
                  {group.mealType && (
                    <Tag color="green">{mealTypeLabels[group.mealType] || group.mealType}</Tag>
                  )}
                  <span style={{ fontSize: 11, color: '#94a3b8' }}>
                    {dayjs(group.recordTime).format('HH:mm')} · 共 {group.items.length} 项
                  </span>
                </div>
                {group.items.map(record => (
                  <div key={record.id} style={{
                    display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                    padding: '8px 0', borderBottom: '1px solid #f8fafc',
                  }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                      <span style={{ fontSize: 13, fontWeight: 500, color: '#334155' }}>{record.food_name}</span>
                      <Tag color={typeColors[record.food_type] || 'default'}>{foodTypeLabels[record.food_type] || record.food_type}</Tag>
                      <span style={{ fontSize: 13 }}>{amountEmoji[record.amount_level] || ''}</span>
                    </div>
                    <Popconfirm title="确认删除？" onConfirm={() => handleDelete(record.id)} okText="删除" cancelText="取消" okButtonProps={{ danger: true }}>
                      <DeleteOutlined style={{ color: '#ef4444', cursor: 'pointer' }} />
                    </Popconfirm>
                  </div>
                ))}
              </div>
            ))
        )}

        <DietForm open={drawerOpen} childId={childId} onClose={() => setDrawerOpen(false)} onSaved={loadData} />
      </DrawerShell>
    </>
  )
}

// ── Supplement Detail ──
export function SupplementDetailDrawer({ open, onClose, childId }: { open: boolean; onClose: () => void; childId: string }) {
  const [records, setRecords] = useState<SupplementRecord[]>([])
  const [selectedDate, setSelectedDate] = useState(dayjs())
  const [loading, setLoading] = useState(true)
  const [drawerOpen, setDrawerOpen] = useState(false)

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listSupplements(childId)
      setRecords(data)
    } catch { message.error('加载失败') }
    finally { setLoading(false) }
  }, [childId])

  useEffect(() => { if (open && childId) loadData() }, [open, childId, loadData])

  const handleDelete = async (sid: string) => {
    try { await deleteSupplement(childId, sid); await loadData() }
    catch { message.error('删除失败') }
  }

  const filtered = records.filter(r => dayjs(r.taken_at).isSame(selectedDate, 'day'))
    .sort((a, b) => dayjs(b.taken_at).valueOf() - dayjs(a.taken_at).valueOf())

  // Streak
  const allDates = [...new Set(records.map(r => dayjs(r.taken_at).format('YYYY-MM-DD')))].sort().reverse()
  let streak = 0
  for (let i = 0; i < 30; i++) {
    const d = dayjs().subtract(i, 'day').format('YYYY-MM-DD')
    if (allDates.includes(d)) streak++
    else if (i > 0) break
  }

  return (
    <>
      <Overlay show={open} onClick={onClose} />
      <DrawerShell open={open} onClose={onClose} title="💊 补剂记录">
        <div style={{ marginBottom: 12 }}>
          <DatePicker value={selectedDate} onChange={v => v && setSelectedDate(v)} format="YYYY-MM-DD" />
        </div>
        <Button type="primary" icon={<PlusOutlined />} block style={{ background: GREEN, marginBottom: 16 }} onClick={() => setDrawerOpen(true)}>
          添加补剂
        </Button>
        {streak > 0 && (
          <div style={{ marginBottom: 12 }}>
            <Tag color="green">🔥 连续 {streak} 天</Tag>
          </div>
        )}

        {loading ? <div style={{ textAlign: 'center', padding: 40, color: '#94a3b8' }}>加载中...</div> : (
          filtered.length === 0
            ? <div style={{ textAlign: 'center', padding: 40, color: '#94a3b8' }}>暂无记录</div>
            : filtered.map(record => (
              <div key={record.id} style={{
                display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                padding: '10px 0', borderBottom: '1px solid #f8fafc',
              }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                  <div style={{ width: 8, height: 8, borderRadius: '50%', background: '#f59e0b' }} />
                  <div>
                    <div style={{ fontSize: 13, fontWeight: 500, color: '#334155' }}>{record.supplement_name}</div>
                    <div style={{ fontSize: 11, color: '#94a3b8', marginTop: 1 }}>
                      {dayjs(record.taken_at).format('HH:mm')} · {record.dose || '—'}
                    </div>
                  </div>
                </div>
                <Popconfirm title="确认删除？" onConfirm={() => handleDelete(record.id)} okText="删除" cancelText="取消" okButtonProps={{ danger: true }}>
                  <DeleteOutlined style={{ color: '#ef4444', cursor: 'pointer' }} />
                </Popconfirm>
              </div>
            ))
        )}

        <SupplementForm open={drawerOpen} childId={childId} editing={null} onClose={() => setDrawerOpen(false)} onSaved={loadData} />
      </DrawerShell>
    </>
  )
}

// ── Measurement Detail (Weight/Height) ──
export function MeasurementDetailDrawer({
  open, onClose, childId, type,
}: { open: boolean; onClose: () => void; childId: string; type: 'weight' | 'height' }) {
  const [measurements, setMeasurements] = useState<Measurement[]>([])
  const [loading, setLoading] = useState(true)
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [editing, setEditing] = useState<Measurement | null>(null)

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listMeasurements(childId, type)
      setMeasurements(data)
    } catch { message.error('加载失败') }
    finally { setLoading(false) }
  }, [childId, type])

  useEffect(() => { if (open && childId) loadData() }, [open, childId, type, loadData])

  const handleDelete = async (mid: string) => {
    try { await deleteMeasurement(childId, mid); await loadData() }
    catch { message.error('删除失败') }
  }

  const sorted = [...measurements].sort((a, b) => dayjs(b.measured_at).valueOf() - dayjs(a.measured_at).valueOf())
  const unit = type === 'weight' ? 'kg' : 'cm'

  return (
    <>
      <Overlay show={open} onClick={onClose} />
      <DrawerShell open={open} onClose={onClose} title={type === 'weight' ? '⚖️ 体重记录' : '📏 身高记录'}>
        <Button type="primary" icon={<PlusOutlined />} block style={{ background: GREEN, marginBottom: 16 }} onClick={() => { setEditing(null); setDrawerOpen(true) }}>
          添加{type === 'weight' ? '体重' : '身高'}
        </Button>

        {loading ? <div style={{ textAlign: 'center', padding: 40, color: '#94a3b8' }}>加载中...</div> : (
          sorted.length === 0
            ? <div style={{ textAlign: 'center', padding: 40, color: '#94a3b8' }}>暂无记录</div>
            : sorted.map(m => (
              <div key={m.id} style={{
                display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                padding: '12px 0', borderBottom: '1px solid #f8fafc',
              }}>
                <span style={{ fontSize: 12, fontWeight: 700, color: '#475569' }}>
                  {dayjs(m.measured_at).format('YYYY-MM-DD')}
                </span>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <span style={{ fontSize: 13, fontWeight: 700, color: '#15803d' }}>
                    {m.value} {unit}
                  </span>
                  <Popconfirm title="确认删除？" onConfirm={() => handleDelete(m.id)} okText="删除" cancelText="取消" okButtonProps={{ danger: true }}>
                    <DeleteOutlined style={{ color: '#ef4444', cursor: 'pointer' }} />
                  </Popconfirm>
                </div>
              </div>
            ))
        )}

        <MeasurementDrawer open={drawerOpen} childId={childId} type={type} editing={editing} onClose={() => setDrawerOpen(false)} onSaved={loadData} />
      </DrawerShell>
    </>
  )
}
