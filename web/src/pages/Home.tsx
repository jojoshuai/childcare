// web/src/pages/Home.tsx
import { useState, useEffect, useCallback, useMemo } from 'react'
import { listChildren } from '../api/children'
import type { Child } from '../api/children'
import { listMeasurements } from '../api/measurements'
import type { Measurement } from '../api/measurements'
import { listSleep } from '../api/sleep'
import type { SleepRecord } from '../api/sleep'
import { listDiet } from '../api/diet'
import type { DietRecord } from '../api/diet'
import { listSupplements } from '../api/supplement'
import type { SupplementRecord } from '../api/supplement'
import { getWHOStandards } from '../api/who'
import type { WHOPoint } from '../api/who'
import ChildSwitcher from '../components/ChildSwitcher'
import StatCards from '../components/StatCards'
import { Button, Empty } from 'antd'
import GrowthChart from '../components/GrowthChart'
import SleepChart from '../components/SleepChart'
import DietNutritionHeatmap from '../components/DietNutritionHeatmap'
import SupplementHeatmap from '../components/SupplementHeatmap'
import {
  SleepDetailDrawer,
  DietDetailDrawer,
  SupplementDetailDrawer,
  MeasurementDetailDrawer,
} from '../components/DetailDrawers'
import dayjs from 'dayjs'

type DetailType = 'sleep' | 'diet' | 'supplement' | 'weight' | 'height' | null

export default function Home() {
  const [child, setChild] = useState<Child | null>(null)
  const [hasChildrenLoaded, setHasChildrenLoaded] = useState(false)
  const [detailType, setDetailType] = useState<DetailType>(null)

  // Load children on mount
  useEffect(() => {
    listChildren().then(cs => {
      setHasChildrenLoaded(true)
      if (cs.length > 0) setChild(cs[0])
    })
  }, [])

  // ── No children empty state ──
  if (hasChildrenLoaded && !child) {
    return (
      <div style={{ maxWidth: 1120, margin: '0 auto', padding: '80px 24px' }}>
        <div style={{ marginBottom: 16 }}>
          <ChildSwitcher selectedId={null} onSelect={setChild} />
        </div>
        <Empty description="还没有孩子，点击右侧 + 添加" />
      </div>
    )
  }

  // ── Children still loading ──
  if (!hasChildrenLoaded) {
    return (
      <div style={{ maxWidth: 1120, margin: '0 auto', padding: '80px 24px' }}>
        <Empty description="加载中..." />
      </div>
    )
  }

  // ── Main content keyed by child.id to force remount on switch ──
  return (
    <div style={{ maxWidth: 1120, margin: '0 auto', padding: '20px 24px 24px' }}>
      {/* Child switcher */}
      <div style={{ marginBottom: 16 }}>
        <ChildSwitcher selectedId={child.id} onSelect={setChild} />
      </div>

      <ChildData child={child} detailType={detailType} onDetailChange={setDetailType} />
    </div>
  )
}

// ── Data section: remounts when child.id changes ──
function ChildData({
  child,
  detailType,
  onDetailChange,
}: {
  child: Child
  detailType: DetailType
  onDetailChange: (d: DetailType) => void
}) {
  const [weightData, setWeightData] = useState<Measurement[]>([])
  const [heightData, setHeightData] = useState<Measurement[]>([])
  const [whoWeight, setWhoWeight] = useState<WHOPoint[]>([])
  const [whoHeight, setWhoHeight] = useState<WHOPoint[]>([])
  const [sleepRecords, setSleepRecords] = useState<SleepRecord[]>([])
  const [dietRecords, setDietRecords] = useState<DietRecord[]>([])
  const [suppRecords, setSuppRecords] = useState<SupplementRecord[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    Promise.all([
      listMeasurements(child.id, 'weight'),
      listMeasurements(child.id, 'height'),
      getWHOStandards(child.gender, 'weight'),
      getWHOStandards(child.gender, 'height'),
      listSleep(child.id),
      listDiet(child.id),
      listSupplements(child.id),
    ]).then(([w, h, wr, hr, sleeps, diets, supps]) => {
      if (cancelled) return
      setWeightData(w)
      setHeightData(h)
      setWhoWeight(wr)
      setWhoHeight(hr)
      setSleepRecords(sleeps)
      setDietRecords(diets)
      setSuppRecords(supps)
    }).finally(() => {
      if (!cancelled) setLoading(false)
    })
    return () => { cancelled = true }
  }, [child.id, child.gender])

  // ── Stat computations ──
  const today = dayjs().format('YYYY-MM-DD')

  const todaySleep = useMemo(() => {
    const todayRecords = sleepRecords.filter(r => dayjs(r.start_time).isSame(today, 'day'))
    let totalMin = 0
    todayRecords.forEach(r => {
      if (r.end_time) totalMin += dayjs(r.end_time).diff(dayjs(r.start_time), 'minute')
    })
    const h = Math.floor(totalMin / 60)
    const m = totalMin % 60
    return `${h}h${m > 0 ? `${m}m` : ''}`
  }, [sleepRecords, today])

  const yesterdaySleepMin = useMemo(() => {
    const yday = dayjs().subtract(1, 'day').format('YYYY-MM-DD')
    let totalMin = 0
    sleepRecords.filter(r => dayjs(r.start_time).isSame(yday, 'day')).forEach(r => {
      if (r.end_time) totalMin += dayjs(r.end_time).diff(dayjs(r.start_time), 'minute')
    })
    return totalMin
  }, [sleepRecords])

  const sleepTrend = useMemo(() => {
    const todayMin = sleepRecords
      .filter(r => dayjs(r.start_time).isSame(today, 'day') && r.end_time)
      .reduce((sum, r) => sum + dayjs(r.end_time!).diff(dayjs(r.start_time), 'minute'), 0)
    const diff = todayMin - yesterdaySleepMin
    if (diff === 0) return '--'
    return diff > 0 ? `+${Math.round(diff / 60 * 10) / 10}h` : `${Math.round(diff / 60 * 10) / 10}h`
  }, [sleepRecords, today, yesterdaySleepMin])

  const todayDietTypes = useMemo(() => {
    const set = new Set<string>()
    dietRecords.filter(r => dayjs(r.record_time).isSame(today, 'day')).forEach(r => set.add(r.food_type))
    const core = ['staple', 'vegetable', 'fruit', 'protein', 'dairy']
    const count = core.filter(t => set.has(t)).length
    return `${count}/5`
  }, [dietRecords, today])

  const dietTrend = useMemo(() => {
    const yday = dayjs().subtract(1, 'day').format('YYYY-MM-DD')
    const yset = new Set<string>()
    dietRecords.filter(r => dayjs(r.record_time).isSame(yday, 'day')).forEach(r => yset.add(r.food_type))
    const core = ['staple', 'vegetable', 'fruit', 'protein', 'dairy']
    const yCount = core.filter(t => yset.has(t)).length
    const tCount = parseInt(todayDietTypes.split('/')[0])
    const diff = tCount - yCount
    if (diff === 0) return '--'
    return diff > 0 ? `+${diff}` : `${diff}`
  }, [dietRecords, todayDietTypes])

  const latestWeight = weightData.length > 0
    ? [...weightData].sort((a, b) => dayjs(b.measured_at).valueOf() - dayjs(a.measured_at).valueOf())[0]
    : null
  const latestHeight = heightData.length > 0
    ? [...heightData].sort((a, b) => dayjs(b.measured_at).valueOf() - dayjs(a.measured_at).valueOf())[0]
    : null

  return (
    <>
      {/* Stat cards */}
      <StatCards
        sleepHours={todaySleep || '--'}
        sleepTrend={sleepTrend}
        dietCount={todayDietTypes}
        dietTrend={dietTrend}
        weightValue={latestWeight ? `${latestWeight.value} kg` : '--'}
        heightValue={latestHeight ? `${latestHeight.value} cm` : '--'}
        percentile="P50+"
        loading={loading}
      />

      {/* Section: 记录 */}
      <div style={{
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        marginBottom: 10,
      }}>
        <div style={{ fontSize: 13, fontWeight: 700, color: '#1e293b', display: 'flex', alignItems: 'center', gap: 8 }}>
          <div style={{ width: 8, height: 8, borderRadius: '50%', background: '#22c55e' }} />
          记录
        </div>
      </div>

      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(3, 1fr)',
        gap: 14,
        marginBottom: 16,
      }}>
        {/* Sleep card */}
        <CardShell>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
            <span style={{ fontSize: 13, fontWeight: 700, color: '#1e293b' }}>睡眠</span>
            <Button size="small" type="text" onClick={() => onDetailChange('sleep')} style={{ fontSize: 11, color: '#475569' }}>
              详情
            </Button>
          </div>
          <SleepChart childId={child.id} />
        </CardShell>

        {/* Diet card */}
        <CardShell>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
            <span style={{ fontSize: 13, fontWeight: 700, color: '#1e293b' }}>饮食</span>
            <Button size="small" type="text" onClick={() => onDetailChange('diet')} style={{ fontSize: 11, color: '#475569' }}>
              详情
            </Button>
          </div>
          <DietNutritionHeatmap childId={child.id} />
        </CardShell>

        {/* Supplement card */}
        <CardShell>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
            <span style={{ fontSize: 13, fontWeight: 700, color: '#1e293b' }}>补剂</span>
            <Button size="small" type="text" onClick={() => onDetailChange('supplement')} style={{ fontSize: 11, color: '#475569' }}>
              详情
            </Button>
          </div>
          <SupplementHeatmap childId={child.id} />
        </CardShell>
      </div>

      {/* Section: 生长曲线 */}
      <div style={{
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        marginBottom: 10,
      }}>
        <div style={{ fontSize: 13, fontWeight: 700, color: '#1e293b', display: 'flex', alignItems: 'center', gap: 8 }}>
          <div style={{ width: 8, height: 8, borderRadius: '50%', background: '#60a5fa' }} />
          生长曲线
        </div>
      </div>

      <div style={{
        display: 'grid',
        gridTemplateColumns: '1fr 1fr',
        gap: 14,
        marginBottom: 16,
      }}>
        {/* Height */}
        <CardShell>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 10 }}>
            <span style={{ fontSize: 13, fontWeight: 700, color: '#1e293b' }}>身高</span>
            <Button size="small" type="text" onClick={() => onDetailChange('height')} style={{ fontSize: 11, color: '#475569' }}>
              详情
            </Button>
          </div>
          <GrowthChart
            measurements={heightData}
            whoData={whoHeight}
            birthDate={child.birth_date}
            type="height"
          />
        </CardShell>

        {/* Weight */}
        <CardShell>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 10 }}>
            <span style={{ fontSize: 13, fontWeight: 700, color: '#1e293b' }}>体重</span>
            <Button size="small" type="text" onClick={() => onDetailChange('weight')} style={{ fontSize: 11, color: '#475569' }}>
              详情
            </Button>
          </div>
          <GrowthChart
            measurements={weightData}
            whoData={whoWeight}
            birthDate={child.birth_date}
            type="weight"
          />
        </CardShell>
      </div>

      {/* Detail Drawers */}
      <SleepDetailDrawer
        open={detailType === 'sleep'}
        onClose={() => onDetailChange(null)}
        childId={child.id}
      />
      <DietDetailDrawer
        open={detailType === 'diet'}
        onClose={() => onDetailChange(null)}
        childId={child.id}
      />
      <SupplementDetailDrawer
        open={detailType === 'supplement'}
        onClose={() => onDetailChange(null)}
        childId={child.id}
      />
      <MeasurementDetailDrawer
        open={detailType === 'weight'}
        onClose={() => onDetailChange(null)}
        childId={child.id}
        type="weight"
      />
      <MeasurementDetailDrawer
        open={detailType === 'height'}
        onClose={() => onDetailChange(null)}
        childId={child.id}
        type="height"
      />
    </>
  )
}

// ── Card Shell helper ──
function CardShell({ children }: { children: React.ReactNode }) {
  return (
    <div style={{
      background: '#fff', borderRadius: 12, padding: 16,
      boxShadow: '0 1px 2px rgba(0,0,0,.04), 0 1px 3px rgba(0,0,0,.06)',
      border: '1px solid #f1f5f9',
    }}>
      {children}
    </div>
  )
}
