// web/src/components/DietNutritionHeatmap.tsx
import { useState, useEffect, useMemo } from 'react'
import { Tooltip, Tag, Skeleton } from 'antd'
import { listDiet } from '../api/diet'
import type { DietRecord } from '../api/diet'
import TimeRangePicker, { rangeDays } from './TimeRangePicker'
import dayjs from 'dayjs'

type Mode = 'week' | 'month'

interface Props {
  childId: string
}

const foodTypes = [
  { key: 'staple', label: '主食' },
  { key: 'vegetable', label: '蔬菜' },
  { key: 'fruit', label: '水果' },
  { key: 'protein', label: '肉蛋' },
  { key: 'dairy', label: '奶' },
  { key: 'snack', label: '零食' },
]

const typeColors: Record<string, string> = {
  staple: '#f59e0b',
  vegetable: '#22c55e',
  fruit: '#ec4899',
  protein: '#3b82f6',
  dairy: '#06b6d4',
  snack: '#f97316',
}

export default function DietNutritionHeatmap({ childId }: Props) {
  const [records, setRecords] = useState<DietRecord[]>([])
  const [loading, setLoading] = useState(true)
  const [mode, setMode] = useState<Mode>('week')
  const [baseDate, setBaseDate] = useState(dayjs())

  useEffect(() => {
    setLoading(true)
    listDiet(childId)
      .then(setRecords)
      .catch(() => setRecords([]))
      .finally(() => setLoading(false))
  }, [childId])

  const days = useMemo(() => rangeDays(mode, baseDate), [mode, baseDate])
  const today = dayjs().format('YYYY-MM-DD')

  // Build map: date -> Set<food_type>
  const typeSetMap = useMemo(() => {
    const map: Record<string, Set<string>> = {}
    const daySet = new Set(days)
    records.forEach(r => {
      const date = dayjs(r.record_time).format('YYYY-MM-DD')
      if (!daySet.has(date)) return
      if (!map[date]) map[date] = new Set()
      map[date].add(r.food_type)
    })
    return map
  }, [records, days])

  // Count covered types today
  const todayTypes = typeSetMap[today] ?? new Set<string>()
  const coveredCount = foodTypes.filter(t => todayTypes.has(t.key) && t.key !== 'snack').length

  if (loading) return <Skeleton active paragraph={{ rows: 2 }} />
  if (records.length === 0) return null

  const cellSize = 28

  return (
    <div>
      <TimeRangePicker
        mode={mode}
        onModeChange={setMode}
        baseDate={baseDate}
        onBaseDateChange={setBaseDate}
      />

      <div style={{ marginBottom: 12 }}>
        <Tag color={coveredCount >= 4 ? 'green' : coveredCount >= 2 ? 'orange' : 'red'}>
          今天 {coveredCount}/5 类
        </Tag>
      </div>

      <div style={{ overflowX: 'auto' }}>
        <div style={{ display: 'inline-block' }}>
          {foodTypes.map(ft => (
            <div key={ft.key} style={{ display: 'flex', height: cellSize, marginBottom: 2 }}>
              <div
                style={{
                  width: 48,
                  flexShrink: 0,
                  fontSize: 11,
                  color: '#6b7280',
                  lineHeight: `${cellSize}px`,
                  textAlign: 'right',
                  paddingRight: 8,
                }}
              >
                {ft.label}
              </div>
              {days.map(d => {
                const types = typeSetMap[d]
                const hasType = types?.has(ft.key) ?? false
                const isToday = d === today
                const bgColor = hasType ? typeColors[ft.key] : '#f3f4f6'
                const textColor = hasType ? '#fff' : '#d1d5db'

                return (
                  <Tooltip
                    key={d}
                    title={hasType ? `${ft.label} · ${d} ✓` : `${ft.label} · ${d} 未吃`}
                  >
                    <div
                      style={{
                        width: cellSize,
                        height: cellSize,
                        borderRadius: 3,
                        background: bgColor,
                        marginRight: 2,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        fontSize: 10,
                        color: textColor,
                        border: isToday ? '2px solid #15803d' : '1px solid transparent',
                        cursor: 'default',
                      }}
                    >
                      {hasType ? '✓' : ''}
                    </div>
                  </Tooltip>
                )
              })}
            </div>
          ))}
        </div>
      </div>

      <div style={{ display: 'flex', alignItems: 'center', marginTop: 8, gap: 8, fontSize: 11, color: '#9ca3af' }}>
        <span>食物类型：</span>
        {foodTypes.map(ft => (
          <div key={ft.key} style={{ display: 'flex', alignItems: 'center', gap: 2 }}>
            <div style={{ width: 10, height: 10, borderRadius: 2, background: typeColors[ft.key] }} />
            <span>{ft.label}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
