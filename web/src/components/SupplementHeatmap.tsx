// web/src/components/SupplementHeatmap.tsx
import { useState, useEffect, useMemo } from 'react'
import { Tooltip, Tag, Skeleton } from 'antd'
import { listSupplements } from '../api/supplement'
import type { SupplementRecord } from '../api/supplement'
import TimeRangePicker, { rangeDays } from './TimeRangePicker'
import dayjs from 'dayjs'

type Mode = 'week' | 'month'

interface Props {
  childId: string
}

const noData = '#f3f4f6'
const lightGreen = '#bbf7d0'
const midGreen = '#4ade80'
const darkGreen = '#16a34a'

function getColor(count: number, hasSupps: boolean) {
  if (!hasSupps) return noData
  if (count === 1) return lightGreen
  if (count === 2) return midGreen
  return darkGreen
}

export default function SupplementHeatmap({ childId }: Props) {
  const [records, setRecords] = useState<SupplementRecord[]>([])
  const [loading, setLoading] = useState(true)
  const [mode, setMode] = useState<Mode>('week')
  const [baseDate, setBaseDate] = useState(dayjs())

  useEffect(() => {
    setLoading(true)
    listSupplements(childId)
      .then(setRecords)
      .catch(() => setRecords([]))
      .finally(() => setLoading(false))
  }, [childId])

  const days = useMemo(() => rangeDays(mode, baseDate), [mode, baseDate])
  const today = dayjs().format('YYYY-MM-DD')

  // Get unique supplement names from ALL records (not just current range)
  // so the rows are always visible
  const suppNames = useMemo(() => {
    const set = new Set<string>()
    records.forEach(r => set.add(r.supplement_name))
    return Array.from(set)
  }, [records])

  // Build count map: suppName -> date -> count (only for current range days)
  const countMap = useMemo(() => {
    const map: Record<string, Record<string, number>> = {}
    const daySet = new Set(days)
    records.forEach(r => {
      const date = dayjs(r.taken_at).format('YYYY-MM-DD')
      if (!daySet.has(date)) return
      if (!map[r.supplement_name]) map[r.supplement_name] = {}
      map[r.supplement_name][date] = (map[r.supplement_name][date] || 0) + 1
    })
    return map
  }, [records, days])

  // Calculate streak (consecutive days from today backwards, within range)
  const streak = useMemo(() => {
    let count = 0
    for (let i = 0; i < days.length; i++) {
      const d = days[days.length - 1 - i] // from most recent backwards
      if (d === today || dayjs(d).isBefore(today)) {
        const anyTaken = suppNames.some(n => (countMap[n]?.[d] ?? 0) > 0)
        if (anyTaken) count++
        else break
      }
    }
    return count
  }, [suppNames, countMap, days, today])

  if (loading) return <Skeleton active paragraph={{ rows: 2 }} />
  if (suppNames.length === 0) return null

  const cellSize = 28

  return (
    <div>
      <TimeRangePicker
        mode={mode}
        onModeChange={setMode}
        baseDate={baseDate}
        onBaseDateChange={setBaseDate}
      />

      {streak > 0 && (
        <div style={{ marginBottom: 12 }}>
          <Tag color="green">连续 {streak} 天</Tag>
        </div>
      )}

      <div style={{ overflowX: 'auto' }}>
        <div style={{ display: 'inline-block' }}>
          {suppNames.map(name => (
            <div key={name} style={{ display: 'flex', height: cellSize, marginBottom: 2 }}>
              <div
                style={{
                  width: 60,
                  flexShrink: 0,
                  fontSize: 11,
                  color: '#6b7280',
                  lineHeight: `${cellSize}px`,
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                }}
              >
                {name}
              </div>
              {days.map(d => {
                const count = countMap[name]?.[d] ?? 0
                const hasSupps = count > 0
                const color = getColor(count, hasSupps)
                const isToday = d === today

                return (
                  <Tooltip
                    key={d}
                    title={hasSupps ? `${name} · ${d} 打卡 ${count} 次` : `${name} · ${d} 未打卡`}
                  >
                    <div
                      style={{
                        width: cellSize,
                        height: cellSize,
                        borderRadius: 3,
                        background: color,
                        marginRight: 2,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        fontSize: 9,
                        color: hasSupps ? '#fff' : '#d1d5db',
                        border: isToday ? '1px solid #16a34a' : '1px solid transparent',
                        cursor: 'default',
                      }}
                    >
                      {count > 0 ? count : ''}
                    </div>
                  </Tooltip>
                )
              })}
            </div>
          ))}
        </div>
      </div>

      <div style={{ display: 'flex', alignItems: 'center', marginTop: 8, gap: 4, fontSize: 11, color: '#9ca3af' }}>
        <span>少</span>
        <div style={{ width: 14, height: 14, borderRadius: 2, background: lightGreen }} />
        <div style={{ width: 14, height: 14, borderRadius: 2, background: midGreen }} />
        <div style={{ width: 14, height: 14, borderRadius: 2, background: darkGreen }} />
        <span>多</span>
      </div>
    </div>
  )
}
