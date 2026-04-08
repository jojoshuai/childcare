// miniprogram/pages/chart/chart.js
const { request } = require('../../utils/request')
const { calcAgeMonths, ageLabel } = require('../../utils/util')

// 注意：安装 wx-charts 并在微信开发者工具中"构建 npm"后，此 require 才能生效
const WxCharts = require('wx-charts')

let chartInstance = null

Page({
  data: {
    childId: '',
    childName: '',
    childAge: '',
    child: null,        // 完整 child 对象（含 birth_date, gender）
    type: 'weight',
    measurements: [],   // 已按 measured_at 倒序排列
    showWHO: true,
    chartLoading: true,
    chartWidth: 320,
    typeLabelMap: { weight: '体重', height: '身高', head_circumference: '头围' },
    unitMap: { weight: 'kg', height: 'cm', head_circumference: 'cm' },
  },

  onLoad(options) {
    const childId = options.id
    const { windowWidth } = wx.getSystemInfoSync()
    this.setData({ childId, chartWidth: windowWidth - 64 }) // 减去 padding

    // 从缓存或 API 取孩子信息
    const children = getApp().globalData.children || []
    const child = children.find(c => c.id === childId)
    if (child) {
      this.setData({
        child,
        childName: child.name,
        childAge: ageLabel(child.birth_date),
        showWHO: calcAgeMonths(child.birth_date) < 61,
      })
      this.loadData()
    } else {
      // 缓存没有，重新拉
      request({ url: '/api/children' }).then(data => {
        getApp().globalData.children = data
        const c = data.find(x => x.id === childId)
        if (!c) return
        this.setData({
          child: c,
          childName: c.name,
          childAge: ageLabel(c.birth_date),
          showWHO: calcAgeMonths(c.birth_date) < 61,
        })
        this.loadData()
      })
    }
  },

  onTypeChange(e) {
    chartInstance = null // 重置图表实例
    this.setData({ type: e.currentTarget.dataset.type, chartLoading: true })
    this.loadData()
  },

  loadData() {
    const { childId, type, child, showWHO } = this.data
    if (!child) return

    const measureReq = request({ url: `/api/children/${childId}/measurements?type=${type}` })
    const whoReq = showWHO
      ? request({ url: `/api/who-standards?gender=${child.gender}&type=${type}` })
      : Promise.resolve({ data: [] })

    Promise.all([measureReq, whoReq]).then(([measureData, whoResp]) => {
      // 按 measured_at 倒序排列（列表显示用）
      const measurements = [...measureData].sort(
        (a, b) => new Date(b.measured_at) - new Date(a.measured_at)
      )
      this.setData({ measurements, chartLoading: false })

      if (measurements.length > 0) {
        this.drawChart(measureData, whoResp.data, child)
      }
    }).catch(() => {
      this.setData({ chartLoading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    })
  },

  drawChart(measureData, whoData, child) {
    const { type, chartWidth } = this.data
    const unitMap = { weight: 'kg', height: 'cm', head_circumference: 'cm' }

    // 计算每条测量记录的月龄（用测量日期而非今天）
    const childPoints = measureData.map(m => ({
      month: calcAgeMonths(child.birth_date, m.measured_at),
      value: m.value,
    }))

    // 构建月龄并集作为 X 轴 categories
    const childMonths = childPoints.map(p => p.month)
    const whoMonths = whoData.map(w => w.month)
    const allMonths = [...new Set([...childMonths, ...whoMonths])].sort((a, b) => a - b)

    // 建立快速查找表
    const childByMonth = {}
    childPoints.forEach(p => { childByMonth[p.month] = p.value })
    const whoByMonth = {}
    whoData.forEach(w => { whoByMonth[w.month] = w })

    // 生成 categories 标签（每 6 个月显示一次，减少拥挤）
    const categories = allMonths.map(m => m % 6 === 0 ? String(m) : '')

    // 生成各系列数据（null 表示该月无数据）
    const childValues = allMonths.map(m => childByMonth[m] !== undefined ? childByMonth[m] : null)
    const p3Values    = allMonths.map(m => whoByMonth[m] ? whoByMonth[m].p3  : null)
    const p50Values   = allMonths.map(m => whoByMonth[m] ? whoByMonth[m].p50 : null)
    const p97Values   = allMonths.map(m => whoByMonth[m] ? whoByMonth[m].p97 : null)

    const series = [
      { name: '孩子', data: childValues, color: '#16a34a', format: v => v !== null ? String(v) : '' },
    ]
    if (whoData.length > 0) {
      series.push({ name: 'P97', data: p97Values, color: '#d1d5db', format: () => '' })
      series.push({ name: 'P50', data: p50Values, color: '#9ca3af', format: () => '' })
      series.push({ name: 'P3',  data: p3Values,  color: '#d1d5db', format: () => '' })
    }

    // wx-charts 需在 nextTick 后初始化（确保 canvas 已渲染）
    wx.nextTick(() => {
      chartInstance = new WxCharts({
        canvasId: 'growthChart',
        type: 'line',
        categories,
        series,
        yAxis: { title: unitMap[type], titleFontSize: 10 },
        width: chartWidth,
        height: 250,
        dataLabel: false,
        dataPointShape: true,
        extra: { lineStyle: 'straight' },
      })
    })
  },

  onDelete(e) {
    const mid = e.currentTarget.dataset.id
    const { childId } = this.data
    wx.showModal({
      title: '确认删除',
      content: '确认删除这条记录？',
      confirmText: '删除',
      confirmColor: '#ef4444',
      success: (res) => {
        if (!res.confirm) return
        request({
          url: `/api/children/${childId}/measurements/${mid}`,
          method: 'DELETE',
        }).then(() => {
          wx.showToast({ title: '已删除', icon: 'success' })
          chartInstance = null
          this.setData({ chartLoading: true })
          this.loadData()
        }).catch((err) => {
          wx.showToast({ title: err.message || '删除失败', icon: 'none' })
        })
      },
    })
  },
})
