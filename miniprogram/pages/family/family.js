// miniprogram/pages/family/family.js
const { request } = require('../../utils/request')

Page({
  data: {
    family: null,
    loading: true,
    isOwner: false,
    invite: null,
    generating: false,
  },

  onShow() {
    this.loadFamily()
  },

  loadFamily() {
    this.setData({ loading: true })
    request({ url: '/api/family' }).then((data) => {
      const user = getApp().globalData.user || {}
      this.setData({
        family: data,
        isOwner: user.role === 'owner',
        loading: false,
      })
    }).catch(() => {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    })
  },

  onGenerate() {
    this.setData({ generating: true })
    request({ url: '/api/family/invite', method: 'POST' }).then((data) => {
      // expires_at 为 RFC3339，转为本地 HH:mm
      const expireTime = new Date(data.expires_at).toLocaleTimeString(
        [], { hour: '2-digit', minute: '2-digit' }
      )
      this.setData({
        invite: { code: data.code, expireTime },
        generating: false,
      })
    }).catch((err) => {
      this.setData({ generating: false })
      wx.showToast({ title: err.message || '生成失败', icon: 'none' })
    })
  },

  onCopyCode() {
    if (!this.data.invite) return
    wx.setClipboardData({
      data: this.data.invite.code,
      success() { wx.showToast({ title: '已复制', icon: 'success' }) },
    })
  },
})
