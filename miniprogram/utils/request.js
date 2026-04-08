// miniprogram/utils/request.js
//
// 修改 BASE_URL 为你的后端地址：
//   开发环境：局域网 IP，如 'http://192.168.1.100:8080'
//   生产环境：微信云托管内网地址
const BASE_URL = 'http://caawkcij.childcare.ccwxy1gg.8utosasx.com'

let refreshing = false
let pendingQueue = []

function doRequest({ url, method = 'GET', data, header = {} }) {
  const app = getApp()
  return new Promise((resolve, reject) => {
    wx.request({
      url: BASE_URL + url,
      method,
      data,
      header: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${app.globalData.token || ''}`,
        ...header,
      },
      success(res) {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(res.data)
        } else if (res.statusCode === 401) {
          handleUnauthorized({ url, method, data, header }, resolve, reject)
        } else {
          reject(res.data || { message: '请求失败' })
        }
      },
      fail(err) {
        reject({ message: '网络错误，请检查连接' })
      },
    })
  })
}

function handleUnauthorized(origReq, resolve, reject) {
  if (refreshing) {
    pendingQueue.push({ origReq, resolve, reject })
    return
  }
  refreshing = true
  const refreshToken = getApp().globalData.refreshToken || wx.getStorageSync('refresh_token')

  if (!refreshToken) {
    refreshing = false
    reLogin()
    reject({ message: '登录已过期，请重新登录' })
    return
  }

  wx.request({
    url: BASE_URL + '/api/auth/refresh',
    method: 'POST',
    data: { refresh_token: refreshToken },
    header: { 'Content-Type': 'application/json' },
    success(res) {
      refreshing = false
      if (res.statusCode === 200) {
        getApp().globalData.token = res.data.token
        getApp().globalData.refreshToken = res.data.refresh_token
        wx.setStorageSync('token', res.data.token)
        wx.setStorageSync('refresh_token', res.data.refresh_token)
        // 重试所有等待中的请求
        pendingQueue.forEach(({ origReq, resolve, reject }) => {
          doRequest(origReq).then(resolve).catch(reject)
        })
        pendingQueue = []
        // 重试原请求
        doRequest(origReq).then(resolve).catch(reject)
      } else {
        pendingQueue.forEach(({ reject }) => reject({ message: '登录已过期' }))
        pendingQueue = []
        reLogin()
        reject({ message: '登录已过期，请重新登录' })
      }
    },
    fail() {
      refreshing = false
      pendingQueue.forEach(({ reject }) => reject({ message: '网络错误' }))
      pendingQueue = []
      reLogin()
      reject({ message: '网络错误' })
    },
  })
}

function reLogin() {
  wx.login({
    success(res) {
      if (!res.code) return
      wx.request({
        url: BASE_URL + '/api/auth/wx-login',
        method: 'POST',
        data: { code: res.code },
        header: { 'Content-Type': 'application/json' },
        success(r) {
          if (r.statusCode === 200) {
            getApp().globalData.token = r.data.token
            getApp().globalData.refreshToken = r.data.refresh_token
            getApp().globalData.user = r.data.user
            wx.setStorageSync('token', r.data.token)
            wx.setStorageSync('refresh_token', r.data.refresh_token)
            wx.setStorageSync('user', r.data.user)
          }
        },
      })
    },
  })
}

module.exports = { request: doRequest }
