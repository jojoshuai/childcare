// web/src/components/GrowthChart.tsx
import { useMemo } from 'react'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'
import type { Measurement } from '../api/measurements'
import type { WHOPoint } from '../api/who'
import dayjs from 'dayjs'

interface Props {
  measurements: Measurement[]
  whoData: WHOPoint[]
  birthDate: string
  type: string
}

function calcAgeMonths(birthDate: string, measuredAt: string): number {
  const diffDays = dayjs(measuredAt).diff(dayjs(birthDate), 'day')
  return Math.floor(diffDays / 30.4375)
}

const typeUnit: Record<string, string> = {
  weight: 'kg',
  height: 'cm',
}

interface ChartPoint {
  date: string
  displayDate: string
  value?: number
  monthAge?: number
  p3?: number
  p50?: number
  p97?: number
}

export default function GrowthChart({
  measurements,
  whoData,
  birthDate,
  type,
}: Props) {
  const maxAgeMonths = useMemo(() => {
    if (!measurements.length) return 0
    return Math.max(
      ...measurements.map(m => calcAgeMonths(birthDate, m.measured_at)),
    )
  }, [measurements, birthDate])

  const showWHO = maxAgeMonths < 61

  // Use actual date as X-axis for fine granularity
  const chartData: ChartPoint[] = useMemo(() => {
    const byDate: Record<string, ChartPoint> = {}

    measurements.forEach(m => {
      const date = dayjs(m.measured_at).format('YYYY-MM-DD')
      const monthAge = calcAgeMonths(birthDate, m.measured_at)
      byDate[date] = {
        ...(byDate[date] ?? { date, displayDate: dayjs(m.measured_at).format('MM/DD'), monthAge }),
        value: m.value,
      }
    })

    // Add WHO reference points for nearby months
    if (showWHO) {
      whoData.forEach(pt => {
        // Find the date closest to this month age
        const targetDate = dayjs(birthDate).add(pt.month, 'month').format('YYYY-MM-DD')
        if (!byDate[targetDate]) {
          byDate[targetDate] = {
            date: targetDate,
            displayDate: `${pt.month}m`,
            p3: pt.p3,
            p50: pt.p50,
            p97: pt.p97,
          }
        }
      })
    }

    return Object.values(byDate).sort((a, b) => a.date.localeCompare(b.date))
  }, [measurements, whoData, birthDate, showWHO])

  const unit = typeUnit[type] ?? ''

  const customTooltip = ({ active, payload }: any) => {
    if (!active || !payload?.length) return null
    const point = payload[0].payload as ChartPoint
    return (
      <div
        style={{
          background: '#fff',
          border: '1px solid #d1fae5',
          borderRadius: 6,
          padding: '8px 12px',
          fontSize: 12,
        }}
      >
        <div style={{ color: '#6b7280', marginBottom: 4 }}>
          {point.displayDate}
          {point.monthAge !== undefined && (
            <span style={{ marginLeft: 8 }}>月龄 {point.monthAge} 个月</span>
          )}
        </div>
        {payload.map((p: any) => (
          <div key={p.name} style={{ color: p.color }}>
            {p.name}: {p.value} {unit}
          </div>
        ))}
      </div>
    )
  }

  return (
    <div>
      <ResponsiveContainer width="100%" height={320}>
        <LineChart data={chartData} margin={{ top: 8, right: 16, left: 0, bottom: 8 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#d1fae5" />
          <XAxis
            dataKey="displayDate"
            label={{
              value: '日期',
              position: 'insideBottomRight',
              offset: -8,
              fontSize: 11,
            }}
            tick={{ fontSize: 11 }}
          />
          <YAxis
            label={{
              value: unit,
              angle: -90,
              position: 'insideLeft',
              offset: 8,
              fontSize: 11,
            }}
            tick={{ fontSize: 11 }}
          />
          <Tooltip content={customTooltip} />
          <Legend />

          <Line
            dataKey="value"
            name="孩子数据"
            stroke="#16a34a"
            strokeWidth={2}
            dot={{ r: 4, fill: '#16a34a' }}
            type="monotone"
            connectNulls
          />

          {showWHO && (
            <>
              <Line
                dataKey="p97"
                name="WHO P97"
                stroke="#9ca3af"
                strokeWidth={1}
                strokeDasharray="4 4"
                dot={false}
                type="monotone"
                connectNulls
              />
              <Line
                dataKey="p50"
                name="WHO P50"
                stroke="#6b7280"
                strokeWidth={1}
                strokeDasharray="4 4"
                dot={false}
                type="monotone"
                connectNulls
              />
              <Line
                dataKey="p3"
                name="WHO P3"
                stroke="#9ca3af"
                strokeWidth={1}
                strokeDasharray="4 4"
                dot={false}
                type="monotone"
                connectNulls
              />
            </>
          )}
        </LineChart>
      </ResponsiveContainer>

      {!showWHO && (
        <div
          style={{
            textAlign: 'center',
            color: '#9ca3af',
            fontSize: 12,
            marginTop: 8,
          }}
        >
          WHO 参考数据覆盖范围为 0-60 个月
        </div>
      )}
    </div>
  )
}
