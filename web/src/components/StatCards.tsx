// web/src/components/StatCards.tsx
import { Card, Skeleton } from 'antd'

interface Props {
  sleepHours: string
  sleepTrend: string
  dietCount: string
  dietTrend: string
  weightValue: string
  heightValue: string
  percentile: string
  loading?: boolean
}

export default function StatCards({
  sleepHours, sleepTrend,
  dietCount, dietTrend,
  weightValue, heightValue, percentile,
  loading,
}: Props) {
  if (loading) {
    return <Skeleton active paragraph={{ rows: 1 }} />
  }

  return (
    <div style={{
      display: 'grid',
      gridTemplateColumns: 'repeat(3, 1fr)',
      gap: 10,
      marginBottom: 16,
    }}>
      {/* Sleep */}
      <Card size="small" style={{ borderRadius: 12, border: '1px solid #f1f5f9' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 6 }}>
          <div style={{
            width: 30, height: 30, borderRadius: 10,
            background: '#f5f3ff', display: 'flex', alignItems: 'center',
            justifyContent: 'center', fontSize: 15,
          }}>
            🌙
          </div>
          <span style={{
            fontSize: 11, fontWeight: 600, padding: '2px 6px', borderRadius: 4,
            background: '#f0fdf4', color: '#16a34a',
          }}>
            {sleepTrend}
          </span>
        </div>
        <div style={{ fontSize: 22, fontWeight: 800, color: '#0f172a', letterSpacing: '-.02em' }}>
          {sleepHours}
        </div>
        <div style={{ fontSize: 11, color: '#94a3b8', marginTop: 1 }}>今日睡眠</div>
      </Card>

      {/* Diet */}
      <Card size="small" style={{ borderRadius: 12, border: '1px solid #f1f5f9' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 6 }}>
          <div style={{
            width: 30, height: 30, borderRadius: 10,
            background: '#dcfce7', display: 'flex', alignItems: 'center',
            justifyContent: 'center', fontSize: 15,
          }}>
            🥗
          </div>
          <span style={{
            fontSize: 11, fontWeight: 600, padding: '2px 6px', borderRadius: 4,
            background: '#f0fdf4', color: '#16a34a',
          }}>
            {dietTrend}
          </span>
        </div>
        <div style={{ fontSize: 22, fontWeight: 800, color: '#0f172a', letterSpacing: '-.02em' }}>
          {dietCount}
        </div>
        <div style={{ fontSize: 11, color: '#94a3b8', marginTop: 1 }}>今日饮食种类</div>
      </Card>

      {/* Weight + Height combined */}
      <Card size="small" style={{ borderRadius: 12, border: '1px solid #f1f5f9' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 6 }}>
          <div style={{
            width: 30, height: 30, borderRadius: 10,
            background: '#eff6ff', display: 'flex', alignItems: 'center',
            justifyContent: 'center', fontSize: 15,
          }}>
            📏
          </div>
          <span style={{
            fontSize: 11, fontWeight: 600, padding: '2px 6px', borderRadius: 4,
            background: '#f1f5f9', color: '#64748b',
          }}>
            {percentile}
          </span>
        </div>
        <div style={{ display: 'flex', gap: 16 }}>
          <div style={{ flex: 1 }}>
            <div style={{ fontSize: 10, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '.04em', fontWeight: 600 }}>
              体重
            </div>
            <div style={{ fontSize: 15, fontWeight: 700, color: '#15803d' }}>{weightValue}</div>
          </div>
          <div style={{ flex: 1 }}>
            <div style={{ fontSize: 10, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '.04em', fontWeight: 600 }}>
              身高
            </div>
            <div style={{ fontSize: 15, fontWeight: 700, color: '#15803d' }}>{heightValue}</div>
          </div>
        </div>
      </Card>
    </div>
  )
}
