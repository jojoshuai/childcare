// miniprogram/app.js
App({
  globalData: {
    token: null,
    refreshToken: null,
    user: null,     // { id, nickname, family_id, role }
    children: [],   // 孩子列表缓存
  },

  onLaunch() {
    // 每次启动重新静默登录，后端根据 openid 识别用户
    this.silentLogin()
  },

  silentLogin() {
    wx.login({
      success: (res) => {
        if (!res.code) {
          wx.showToast({ title: '微信登录失败', icon: 'none' })
          return
        }
        const { request } = require('./utils/request')
        request({
          url: '/api/auth/wx-login',
          method: 'POST',
          data: { code: res.code },
        }).then((data) => {
          this.globalData.token = data.token
          this.globalData.refreshToken = data.refresh_token
          this.globalData.user = data.user
          wx.setStorageSync('token', data.token)
          wx.setStorageSync('refresh_token', data.refresh_token)
          wx.setStorageSync('user', data.user)

          if (!data.user.family_id) {
            // 未加入家庭，强制跳转到邀请码页
            wx.reLaunch({ url: '/pages/join/join' })
          }
          // 有 family_id 则正常留在 tabBar（首页）
        }).catch((err) => {
          wx.showToast({ title: err.message || '登录失败', icon: 'none' })
        })
      },
      fail() {
        wx.showToast({ title: '微信登录失败', icon: 'none' })
      },
    })
  },
})
