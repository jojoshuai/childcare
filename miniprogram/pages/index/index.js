// miniprogram/pages/index/index.js
const { request } = require('../../utils/request')
const { ageLabel } = require('../../utils/util')

Page({
  data: {
    children: [],
    loading: true,
  },

  onShow() {
    this.loadChildren()
  },

  onPullDownRefresh() {
    this.loadChildren().finally(() => wx.stopPullDownRefresh())
  },

  loadChildren() {
    this.setData({ loading: true })
    return request({ url: '/api/children' }).then((data) => {
      const children = data.map(c => ({ ...c, _ageLabel: ageLabel(c.birth_date) }))
      getApp().globalData.children = children
      this.setData({ children, loading: false })
    }).catch(() => {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败，请下拉刷新', icon: 'none' })
    })
  },

  onChildTap(e) {
    wx.navigateTo({ url: `/pages/chart/chart?id=${e.currentTarget.dataset.id}` })
  },
})
