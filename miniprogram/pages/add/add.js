// miniprogram/pages/add/add.js
const { request } = require('../../utils/request')
const { today } = require('../../utils/util')

Page({
  data: {
    children: [],
    childIndex: 0,
    type: 'weight',
    value: '',
    date: '',
    note: '',
    loading: false,
    typeLabelMap: { weight: '体重', height: '身高', head_circumference: '头围' },
    unitMap: { weight: 'kg', height: 'cm', head_circumference: 'cm' },
    typeHintMap: {
      weight: '(0.5–200 kg)',
      height: '(20–250 cm)',
      head_circumference: '(20–80 cm)',
    },
    today: today(),
  },

  onShow() {
    const children = getApp().globalData.children || []
    this.setData({ children, date: today() })
    if (children.length === 0) {
      // 尝试重新拉取
      request({ url: '/api/children' }).then(data => {
        getApp().globalData.children = data
        this.setData({ children: data })
      }).catch(() => {})
    }
  },

  onChildChange(e) {
    this.setData({ childIndex: Number(e.detail.value) })
  },

  onTypeSelect(e) {
    this.setData({ type: e.currentTarget.dataset.type, value: '' })
  },

  onValueInput(e) {
    this.setData({ value: e.detail.value })
  },

  onDateChange(e) {
    this.setData({ date: e.detail.value })
  },

  onNoteInput(e) {
    this.setData({ note: e.detail.value })
  },

  onSubmit() {
    const { children, childIndex, type, value, date, note } = this.data
    if (!value) return
    const child = children[childIndex]
    if (!child) return

    this.setData({ loading: true })
    request({
      url: `/api/children/${child.id}/measurements`,
      method: 'POST',
      data: {
        type,
        value: parseFloat(value),
        measured_at: date,
        note: note.trim() || null,
      },
    }).then(() => {
      wx.showToast({ title: '记录成功', icon: 'success' })
      // 重置数值/日期/备注，保留孩子和类型选择
      this.setData({ value: '', date: today(), note: '', loading: false })
    }).catch((err) => {
      this.setData({ loading: false })
      wx.showToast({ title: err.message || '保存失败', icon: 'none' })
    })
  },
})
