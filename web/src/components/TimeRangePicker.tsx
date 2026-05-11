// web/src/components/TimeRangePicker.tsx
import { useCallback } from 'react'
import { Segmented, Button } from 'antd'
import { LeftOutlined, RightOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'

type Mode = 'week' | 'month'

interface Props {
  mode: Mode
  onModeChange: (m: Mode) => void
  baseDate: dayjs.Dayjs
  onBaseDateChange: (d: dayjs.Dayjs) => void
}

/** Get the label string for the current range */
export function rangeLabel(mode: Mode, baseDate: dayjs.Dayjs): string {
  if (mode === 'week') {
    const ws = baseDate.startOf('week').add(1, 'day') // Monday
    const we = ws.add(6, 'day') // Sunday
    return `${ws.format('M/D')} ~ ${we.format('M/D')}`
  }
  return baseDate.format('YYYY 年 M 月')
}

/** Get the list of dates (YYYY-MM-DD) in the current range */
export function rangeDays(mode: Mode, baseDate: dayjs.Dayjs): string[] {
  const days: string[] = []
  if (mode === 'week') {
    const ws = baseDate.startOf('week').add(1, 'day')
    for (let i = 0; i < 7; i++) {
      days.push(ws.add(i, 'day').format('YYYY-MM-DD'))
    }
  } else {
    const monthStart = baseDate.startOf('month')
    const monthEnd = baseDate.endOf('month')
    for (let d = monthStart.clone(); d.isBefore(monthEnd) || d.isSame(monthEnd, 'day'); d = d.add(1, 'day')) {
      days.push(d.format('YYYY-MM-DD'))
    }
  }
  return days
}

export default function TimeRangePicker({ mode, onModeChange, baseDate, onBaseDateChange }: Props) {
  const step = useCallback(
    (dir: 1 | -1) => {
      onBaseDateChange(baseDate.add(dir, mode === 'week' ? 'week' : 'month'))
    },
    [mode, baseDate, onBaseDateChange],
  )

  const isFuture = mode === 'month'
    ? baseDate.isAfter(dayjs(), 'month')
    : baseDate.startOf('week').add(6, 'day').isAfter(dayjs(), 'day')

  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 12 }}>
      <Button size="small" icon={<LeftOutlined />} onClick={() => step(-1)} />
      <Segmented
        size="small"
        value={mode}
        onChange={v => onModeChange(v as Mode)}
        options={[
          { label: '周', value: 'week' },
          { label: '月', value: 'month' },
        ]}
      />
      <span style={{ fontSize: 13, color: '#6b7280', minWidth: 110, textAlign: 'center' }}>
        {rangeLabel(mode, baseDate)}
      </span>
      <Button size="small" icon={<RightOutlined />} onClick={() => step(1)} disabled={isFuture} />
    </div>
  )
}
