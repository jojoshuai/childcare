// web/src/components/SleepChart.tsx
import { useState, useEffect, useMemo, useCallback } from 'react'
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  ReferenceLine,
} from 'recharts'
import { Skeleton } from 'antd'
import { listSleep } from '../api/sleep'
import type { SleepRecord } from '../api/sleep'
import TimeRangePicker, { rangeDays } from './TimeRangePicker'
import dayjs from 'dayjs'

type Mode = 'week' | 'month'

interface Props {
  childId: string
}

// Belonging day rule:
// start hour >= 2 -> belongs to start date
// start hour 0-2 -> belongs to previous date
function belongingDay(record: SleepRecord): string {
  const start = dayjs(record.start_time)
  const hour = start.hour()
  if (hour >= 2) return start.format('YYYY-MM-DD')
  return start.subtract(1, 'day').format('YYYY-MM-DD')
}

function durationMinutes(start: string, end: string | null): number {
  if (!end) return 0
  return dayjs(end).diff(dayjs(start), 'minute')
}

function isNightSleep(record: SleepRecord): boolean {
  const hour = dayjs(record.start_time).hour()
  return hour >= 18 || hour < 6
}

interface DayData {
  date: string
  displayDate: string
  nightMinutes: number
  dayMinutes: number
  totalHours: number
  wokeUp: boolean
  wakeCount: number
}

export default function SleepChart({ childId }: Props) {
  const [records, setRecords] = useState<SleepRecord[]>([])
  const [loading, setLoading] = useState(true)
  const [mode, setMode] = useState<Mode>('week')
  const [baseDate, setBaseDate] = useState(dayjs())

  useEffect(() => {
    setLoading(true)
    listSleep(childId)
      .then(setRecords)
      .catch(() => setRecords([]))
      .finally(() => setLoading(false))
  }, [childId])

  const days = useMemo(() => rangeDays(mode, baseDate), [mode, baseDate])

  const chartData: DayData[] = useMemo(() => {
    const dayList: DayData[] = days.map(d => ({
      date: d,
      displayDate: dayjs(d).format('MM/DD'),
      nightMinutes: 0,
      dayMinutes: 0,
      totalHours: 0,
      wokeUp: false,
      wakeCount: 0,
    }))

    const dayMap = new Map<string, DayData>()
    dayList.forEach(d => dayMap.set(d.date, d))

    for (const r of records) {
      if (!r.end_time) continue
      const bDay = belongingDay(r)
      const day = dayMap.get(bDay)
      if (!day) continue

      const mins = durationMinutes(r.start_time, r.end_time)
      if (isNightSleep(r)) {
        day.nightMinutes += mins
      } else {
        day.dayMinutes += mins
      }
      day.totalHours = (day.nightMinutes + day.dayMinutes) / 60
      if (r.woke_up) {
        day.wokeUp = true
        day.wakeCount += r.wake_count
      }
    }

    return dayList
  }, [records, days])

  const whoLine = 11.5

  if (loading) return <Skeleton active paragraph={{ rows: 1 }} />

  return (
    <div>
      <TimeRangePicker
        mode={mode}
        onModeChange={setMode}
        baseDate={baseDate}
        onBaseDateChange={setBaseDate}
      />
      <ResponsiveContainer width="100%" height={240}>
        <BarChart data={chartData} margin={{ top: 8, right: 16, left: 0, bottom: 8 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#d1fae5" />
          <XAxis
            dataKey="displayDate"
            tick={{ fontSize: 10 }}
          />
          <YAxis
            label={{ value: '小时', angle: -90, position: 'insideLeft', offset: 8, fontSize: 11 }}
            tick={{ fontSize: 10 }}
            domain={[0, 16]}
          />
          <Tooltip
            content={({ active, payload }) => {
              if (!active || !payload?.length) return null
              const d = payload[0].payload as DayData
              return (
                <div style={{ background: '#fff', border: '1px solid #d1fae5', borderRadius: 6, padding: '8px 12px', fontSize: 12 }}>
                  <div style={{ fontWeight: 600, marginBottom: 4 }}>{d.displayDate}</div>
                  <div>夜间: {(d.nightMinutes / 60).toFixed(1)}h</div>
                  <div>白天: {(d.dayMinutes / 60).toFixed(1)}h</div>
                  <div>总计: {d.totalHours.toFixed(1)}h</div>
                  {d.wokeUp && <div style={{ color: '#ef4444' }}>醒来 {d.wakeCount} 次</div>}
                </div>
              )
            }}
          />
          <Legend />
          <ReferenceLine
            y={whoLine}
            stroke="#9ca3af"
            strokeDasharray="4 4"
            label={{ value: `WHO ${whoLine}h`, position: 'right', fill: '#9ca3af', fontSize: 10 }}
          />
          <Bar dataKey="nightMinutes" name="夜间睡眠" stackId="sleep" fill="#2563eb" radius={[0, 0, 0, 0]} />
          <Bar dataKey="dayMinutes" name="白天小睡" stackId="sleep" fill="#60a5fa" radius={[4, 4, 0, 0]} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  )
}
