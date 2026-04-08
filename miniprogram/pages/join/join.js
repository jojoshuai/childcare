// miniprogram/pages/join/join.js
const { request } = require('../../utils/request')

Page({
  data: {
    code: '',
    loading: false,
    errorMsg: '',
  },

  onCodeInput(e) {
    this.setData({ code: e.detail.value, errorMsg: '' })
  },

  onSubmit() {
    const code = this.data.code.trim().toUpperCase()
    if (code.length < 6) return

    this.setData({ loading: true, errorMsg: '' })
    request({
      url: '/api/family/join',
      method: 'POST',
      data: { code },
    }).then((data) => {
      // 更新 globalData.user.family_id（后端只返回 { family_id }）
      const app = getApp()
      app.globalData.user = { ...app.globalData.user, family_id: data.family_id }
      wx.setStorageSync('user', app.globalData.user)
      wx.reLaunch({ url: '/pages/index/index' })
    }).catch((err) => {
      const msgMap = {
        INVITE_CODE_NOT_FOUND: '邀请码不存在',
        INVITE_CODE_EXPIRED: '邀请码已过期',
        INVITE_CODE_ALREADY_USED: '邀请码已被使用',
      }
      this.setData({
        errorMsg: msgMap[err.code] || err.message || '加入失败，请重试',
        loading: false,
      })
    })
  },
})
