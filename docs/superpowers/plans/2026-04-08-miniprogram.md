# 微信小程序实现计划

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现完整的微信小程序 MVP，包含静默登录、邀请码加入家庭、快速录入、生长曲线图、家庭管理。

**Architecture:** 微信原生 WXML/WXSS/JS，globalData 管理全局状态，utils/request.js 封装所有 HTTP 请求（含 401 自动刷新），wx-charts 渲染生长曲线图。

**Tech Stack:** 微信原生小程序、wx-charts（npm）、Go 后端 API（已完成）

---

## 文件结构

| 文件 | 职责 |
|------|------|
| `miniprogram/app.js` | 全局状态 globalData + onLaunch 静默登录 |
| `miniprogram/app.json` | 页面注册 + tabBar 配置 |
| `miniprogram/app.wxss` | 全局样式（绿色主题） |
| `miniprogram/utils/request.js` | wx.request 封装，自动带 token，401 自动刷新 |
| `miniprogram/utils/util.js` | calcAgeMonths / ageLabel 工具函数 |
| `miniprogram/pages/join/join.{js,wxml,wxss}` | 输入邀请码加入家庭 |
| `miniprogram/pages/index/index.{js,wxml,wxss}` | 首页孩子列表 |
| `miniprogram/pages/add/add.{js,wxml,wxss}` | 快速录入（Tab 核心页） |
| `miniprogram/pages/family/family.{js,wxml,wxss}` | 家庭管理 + 邀请码生成 |
| `miniprogram/pages/chart/chart.{js,wxml,wxss}` | 生长曲线图 + 记录列表 |
| `miniprogram/icons/*.png` | tabBar 图标（6 个占位 PNG） |

---

## Chunk 1: 脚手架 + utils + app

### Task 1: 创建项目结构

**Files:**
- Create: `miniprogram/app.json`
- Create: `miniprogram/package.json`
- Create: `miniprogram/icons/` (6 个 PNG 占位文件)

- [ ] **Step 1: 创建目录**

```bash
cd /Users/zyb/Own/code/jojoshuai/childcare
mkdir -p miniprogram/utils
mkdir -p miniprogram/pages/join
mkdir -p miniprogram/pages/index
mkdir -p miniprogram/pages/add
mkdir -p miniprogram/pages/family
mkdir -p miniprogram/pages/chart
mkdir -p miniprogram/icons
```

- [ ] **Step 2: 创建 tabBar 图标占位文件（最小合法 PNG）**

```bash
cd miniprogram
# 最小透明 PNG（1x1 像素），tabBar 可用；上线前替换为真实图标
python3 -c "
import base64, os
data = base64.b64decode('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==')
for name in ['home','home-active','add','add-active','family','family-active']:
    open(f'icons/{name}.png','wb').write(data)
print('done')
"
```

预期输出：`done`，`icons/` 下出现 6 个 PNG 文件。

- [ ] **Step 3: 创建 app.json**

```json
{
  "pages": [
    "pages/index/index",
    "pages/add/add",
    "pages/family/family",
    "pages/chart/chart",
    "pages/join/join"
  ],
  "tabBar": {
    "color": "#6b7280",
    "selectedColor": "#16a34a",
    "backgroundColor": "#ffffff",
    "borderStyle": "white",
    "list": [
      {
        "pagePath": "pages/index/index",
        "text": "首页",
        "iconPath": "icons/home.png",
        "selectedIconPath": "icons/home-active.png"
      },
      {
        "pagePath": "pages/add/add",
        "text": "录入",
        "iconPath": "icons/add.png",
        "selectedIconPath": "icons/add-active.png"
      },
      {
        "pagePath": "pages/family/family",
        "text": "家庭",
        "iconPath": "icons/family.png",
        "selectedIconPath": "icons/family-active.png"
      }
    ]
  },
  "window": {
    "backgroundTextStyle": "light",
    "navigationBarBackgroundColor": "#16a34a",
    "navigationBarTitleText": "儿童成长",
    "navigationBarTextStyle": "white",
    "backgroundColor": "#f0fdf4",
    "enablePullDownRefresh": false
  }
}
```

- [ ] **Step 4: 创建 package.json（用于 npm 安装 wx-charts）**

```json
{
  "name": "childcare-miniprogram",
  "version": "1.0.0",
  "dependencies": {
    "wx-charts": "^2.0.0"
  }
}
```

- [ ] **Step 5: 安装 wx-charts**

```bash
cd /Users/zyb/Own/code/jojoshuai/childcare/miniprogram
npm install
```

预期：生成 `node_modules/wx-charts/` 目录。

**注意：** 安装后需在微信开发者工具中点击 **工具 → 构建 npm**，生成 `miniprogram_npm/` 目录，之后代码中才能 `require('wx-charts')`。

---

### Task 2: utils/request.js

**Files:**
- Create: `miniprogram/utils/request.js`

- [ ] **Step 1: 创建 request.js**

```javascript
// miniprogram/utils/request.js
//
// 修改 BASE_URL 为你的后端地址：
//   开发环境：局域网 IP，如 'http://192.168.1.100:8080'
//   生产环境：微信云托管内网地址，如 'http://childcare-xxxx.ap-shanghai.service.tcloudbase.com'
const BASE_URL = 'http://192.168.1.100:8080'

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
      pendingQueue.forEach(({ reject }) => reject({ message: '网络错误' }))
      pendingQueue = []
      reLogin()
      reject({ message: '网络错误' })
    },
    complete() {
      refreshing = false
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
```

---

### Task 3: utils/util.js

**Files:**
- Create: `miniprogram/utils/util.js`

- [ ] **Step 1: 创建 util.js**

```javascript
// miniprogram/utils/util.js

/**
 * 计算从 birthDate 到 endDate（默认今天）的月龄（整数，向下取整）
 * @param {string} birthDate - ISO 日期字符串，如 '2023-01-15'
 * @param {string} [endDate]  - ISO 日期字符串，不传则用今天
 * @returns {number}
 */
function calcAgeMonths(birthDate, endDate) {
  const end = endDate ? new Date(endDate) : new Date()
  const diffMs = end.getTime() - new Date(birthDate).getTime()
  const diffDays = diffMs / (1000 * 60 * 60 * 24)
  return Math.floor(diffDays / 30.4375)
}

/**
 * 将 birthDate 转为友好月龄字符串，如 "8个月" 或 "1岁3个月"
 * @param {string} birthDate
 * @returns {string}
 */
function ageLabel(birthDate) {
  const months = calcAgeMonths(birthDate)
  if (months < 12) return `${months}个月`
  const years = Math.floor(months / 12)
  const rem = months % 12
  return rem > 0 ? `${years}岁${rem}个月` : `${years}岁`
}

/**
 * 今天的 YYYY-MM-DD 字符串
 * @returns {string}
 */
function today() {
  return new Date().toISOString().slice(0, 10)
}

module.exports = { calcAgeMonths, ageLabel, today }
```

---

### Task 4: app.js + app.wxss

**Files:**
- Create: `miniprogram/app.js`
- Create: `miniprogram/app.wxss`

- [ ] **Step 1: 创建 app.js**

```javascript
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
```

- [ ] **Step 2: 创建 app.wxss**

```css
/* miniprogram/app.wxss */
page {
  background-color: #f0fdf4;
  font-family: -apple-system, BlinkMacSystemFont, 'PingFang SC', sans-serif;
  color: #111827;
}

/* 通用按钮 */
.btn-primary {
  background-color: #16a34a;
  color: #fff;
  border-radius: 8rpx;
  font-size: 32rpx;
  padding: 0;
  line-height: 96rpx;
  height: 96rpx;
  border: none;
}
.btn-primary::after { border: none; }

/* 通用卡片 */
.card {
  background: #fff;
  border-radius: 12rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
  box-shadow: 0 1px 4px rgba(0,0,0,0.06);
}

/* 分隔线 */
.divider {
  height: 1rpx;
  background: #d1fae5;
  margin: 16rpx 0;
}

/* 空状态提示 */
.empty-tip {
  text-align: center;
  color: #9ca3af;
  font-size: 28rpx;
  padding: 80rpx 0;
}
```

- [ ] **Step 3: Commit**

```bash
cd /Users/zyb/Own/code/jojoshuai/childcare
git add miniprogram/
git commit -m "feat(mp): scaffold miniprogram project + utils + app"
```

---

## Chunk 2: join 页 + index 页

### Task 5: join 页（加入家庭）

**Files:**
- Create: `miniprogram/pages/join/join.js`
- Create: `miniprogram/pages/join/join.wxml`
- Create: `miniprogram/pages/join/join.wxss`
- Create: `miniprogram/pages/join/join.json`

- [ ] **Step 1: 创建 join.json**

```json
{
  "navigationBarTitleText": "加入家庭"
}
```

- [ ] **Step 2: 创建 join.wxml**

```xml
<!-- miniprogram/pages/join/join.wxml -->
<view class="container">
  <view class="logo">🌱</view>
  <view class="title">欢迎使用儿童成长记录</view>
  <view class="desc">请输入家长提供的 6 位邀请码加入家庭</view>

  <input
    class="code-input"
    type="text"
    maxlength="6"
    placeholder="输入邀请码，如 AB1234"
    value="{{code}}"
    bindinput="onCodeInput"
  />

  <view wx:if="{{errorMsg}}" class="error-msg">{{errorMsg}}</view>

  <button
    class="btn-primary submit-btn"
    loading="{{loading}}"
    disabled="{{loading || code.length < 6}}"
    bindtap="onSubmit"
  >加入家庭</button>
</view>
```

- [ ] **Step 3: 创建 join.wxss**

```css
/* miniprogram/pages/join/join.wxss */
.container {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 120rpx 48rpx 48rpx;
  min-height: 100vh;
  background: #f0fdf4;
}
.logo {
  font-size: 100rpx;
  margin-bottom: 24rpx;
}
.title {
  font-size: 36rpx;
  font-weight: 600;
  color: #15803d;
  margin-bottom: 16rpx;
}
.desc {
  font-size: 28rpx;
  color: #6b7280;
  text-align: center;
  margin-bottom: 64rpx;
  line-height: 1.6;
}
.code-input {
  width: 100%;
  height: 96rpx;
  border: 2rpx solid #d1fae5;
  border-radius: 12rpx;
  background: #fff;
  font-size: 48rpx;
  letter-spacing: 16rpx;
  text-align: center;
  padding: 0 24rpx;
  box-sizing: border-box;
  margin-bottom: 24rpx;
}
.error-msg {
  color: #ef4444;
  font-size: 26rpx;
  margin-bottom: 24rpx;
}
.submit-btn {
  width: 100%;
  margin-top: 16rpx;
}
```

- [ ] **Step 4: 创建 join.js**

```javascript
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
```

---

### Task 6: index 页（首页）

**Files:**
- Create: `miniprogram/pages/index/index.js`
- Create: `miniprogram/pages/index/index.wxml`
- Create: `miniprogram/pages/index/index.wxss`
- Create: `miniprogram/pages/index/index.json`

- [ ] **Step 1: 创建 index.json**

```json
{
  "navigationBarTitleText": "我的孩子",
  "enablePullDownRefresh": true
}
```

- [ ] **Step 2: 创建 index.wxml**

```xml
<!-- miniprogram/pages/index/index.wxml -->
<view class="container">
  <!-- 加载中 -->
  <view wx:if="{{loading}}" class="loading-wrap">
    <view class="loading-item" wx:for="{{[1,2,3]}}" wx:key="*this">
      <view class="skeleton avatar"></view>
      <view class="skeleton-content">
        <view class="skeleton line-long"></view>
        <view class="skeleton line-short"></view>
      </view>
    </view>
  </view>

  <!-- 空状态 -->
  <view wx:elif="{{children.length === 0}}" class="empty-tip">
    <view style="font-size:80rpx;margin-bottom:16rpx;">👶</view>
    <view>暂无孩子信息</view>
    <view style="margin-top:8rpx;font-size:24rpx;">请在 Web 端添加孩子后刷新</view>
  </view>

  <!-- 孩子列表 -->
  <view wx:else>
    <view
      class="child-item card"
      wx:for="{{children}}"
      wx:key="id"
      bindtap="onChildTap"
      data-id="{{item.id}}"
    >
      <view class="avatar">{{item.gender === 'male' ? '👦' : '👧'}}</view>
      <view class="info">
        <view class="name">{{item.name}}</view>
        <view class="age">{{item._ageLabel}}</view>
      </view>
      <view class="arrow">›</view>
    </view>
  </view>
</view>
```

- [ ] **Step 3: 创建 index.wxss**

```css
/* miniprogram/pages/index/index.wxss */
.container {
  padding: 24rpx;
  min-height: 100vh;
}
.child-item {
  display: flex;
  align-items: center;
  padding: 28rpx 24rpx;
  margin-bottom: 16rpx;
}
.avatar {
  font-size: 64rpx;
  margin-right: 24rpx;
  flex-shrink: 0;
}
.info { flex: 1; }
.name {
  font-size: 34rpx;
  font-weight: 600;
  color: #15803d;
  margin-bottom: 6rpx;
}
.age { font-size: 26rpx; color: #6b7280; }
.arrow {
  font-size: 48rpx;
  color: #16a34a;
  font-weight: 300;
}

/* 骨架屏 */
.loading-wrap { padding: 0; }
.loading-item {
  display: flex;
  align-items: center;
  background: #fff;
  border-radius: 12rpx;
  padding: 28rpx 24rpx;
  margin-bottom: 16rpx;
}
.skeleton {
  background: linear-gradient(90deg, #e5e7eb 25%, #f3f4f6 50%, #e5e7eb 75%);
  background-size: 200% 100%;
  border-radius: 6rpx;
  animation: shimmer 1.2s infinite;
}
@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
.skeleton.avatar { width: 80rpx; height: 80rpx; border-radius: 50%; margin-right: 24rpx; flex-shrink: 0; }
.skeleton-content { flex: 1; }
.skeleton.line-long { height: 32rpx; width: 60%; margin-bottom: 12rpx; }
.skeleton.line-short { height: 24rpx; width: 40%; }
```

- [ ] **Step 4: 创建 index.js**

```javascript
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
```

- [ ] **Step 5: Commit**

```bash
cd /Users/zyb/Own/code/jojoshuai/childcare
git add miniprogram/pages/join/ miniprogram/pages/index/
git commit -m "feat(mp): add join page and index page"
```

---

## Chunk 3: add 页 + family 页

### Task 7: add 页（快速录入）

**Files:**
- Create: `miniprogram/pages/add/add.js`
- Create: `miniprogram/pages/add/add.wxml`
- Create: `miniprogram/pages/add/add.wxss`
- Create: `miniprogram/pages/add/add.json`

- [ ] **Step 1: 创建 add.json**

```json
{
  "navigationBarTitleText": "快速录入"
}
```

- [ ] **Step 2: 创建 add.wxml**

```xml
<!-- miniprogram/pages/add/add.wxml -->
<view class="container">

  <!-- 选孩子（多个孩子时显示） -->
  <view class="card" wx:if="{{children.length > 1}}">
    <view class="section-label">选择孩子</view>
    <picker mode="selector" range="{{children}}" range-key="name" value="{{childIndex}}" bindchange="onChildChange">
      <view class="picker-row">
        <text>{{children[childIndex].gender === 'male' ? '👦' : '👧'}} {{children[childIndex].name}}</text>
        <text class="picker-arrow">▾</text>
      </view>
    </picker>
  </view>

  <!-- 无孩子时提示 -->
  <view class="card empty-tip" wx:elif="{{children.length === 0}}">
    暂无孩子，请先在 Web 端添加
  </view>

  <!-- 有孩子（含只有一个孩子）时显示录入表单 -->
  <view wx:if="{{children.length > 0}}">

    <!-- 选类型 -->
    <view class="card">
      <view class="section-label">测量类型</view>
      <view class="type-row">
        <view
          class="type-btn {{type === 'weight' ? 'type-btn--active' : ''}}"
          bindtap="onTypeSelect"
          data-type="weight"
        >⚖️ 体重</view>
        <view
          class="type-btn {{type === 'height' ? 'type-btn--active' : ''}}"
          bindtap="onTypeSelect"
          data-type="height"
        >📏 身高</view>
        <view
          class="type-btn {{type === 'head_circumference' ? 'type-btn--active' : ''}}"
          bindtap="onTypeSelect"
          data-type="head_circumference"
        >🔵 头围</view>
      </view>
    </view>

    <!-- 输入数值 -->
    <view class="card">
      <view class="section-label">{{typeLabelMap[type]}} <text class="hint">{{typeHintMap[type]}}</text></view>
      <view class="value-row">
        <input
          class="value-input"
          type="digit"
          placeholder="请输入数值"
          value="{{value}}"
          bindinput="onValueInput"
        />
        <text class="unit">{{unitMap[type]}}</text>
      </view>
    </view>

    <!-- 选日期 -->
    <view class="card">
      <view class="section-label">测量日期</view>
      <picker mode="date" value="{{date}}" end="{{today}}" bindchange="onDateChange">
        <view class="picker-row">
          <text>{{date}}</text>
          <text class="picker-arrow">▾</text>
        </view>
      </picker>
    </view>

    <!-- 备注 -->
    <view class="card">
      <view class="section-label">备注（选填）</view>
      <input
        class="note-input"
        placeholder="可以记录一些备注"
        value="{{note}}"
        bindinput="onNoteInput"
      />
    </view>

    <!-- 提交 -->
    <button
      class="btn-primary"
      loading="{{loading}}"
      disabled="{{loading || !value}}"
      bindtap="onSubmit"
    >保存记录</button>

  </view>
</view>
```

- [ ] **Step 3: 创建 add.wxss**

```css
/* miniprogram/pages/add/add.wxss */
.container { padding: 24rpx; }
.section-label {
  font-size: 26rpx;
  color: #6b7280;
  margin-bottom: 16rpx;
}
.hint { font-size: 22rpx; color: #9ca3af; }

/* 类型按钮 */
.type-row {
  display: flex;
  gap: 16rpx;
}
.type-btn {
  flex: 1;
  height: 96rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2rpx solid #d1fae5;
  border-radius: 12rpx;
  font-size: 26rpx;
  color: #374151;
  background: #f9fafb;
}
.type-btn--active {
  background: #16a34a;
  color: #fff;
  border-color: #16a34a;
}

/* 数值输入 */
.value-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
}
.value-input {
  flex: 1;
  height: 80rpx;
  font-size: 44rpx;
  color: #111827;
  border-bottom: 2rpx solid #d1fae5;
  padding: 0 8rpx;
}
.unit { font-size: 28rpx; color: #6b7280; }

/* picker 行 */
.picker-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 72rpx;
  font-size: 30rpx;
  color: #111827;
}
.picker-arrow { color: #9ca3af; }

/* 备注 */
.note-input {
  width: 100%;
  height: 72rpx;
  font-size: 28rpx;
  color: #374151;
}
```

- [ ] **Step 4: 创建 add.js**

```javascript
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
```

---

### Task 8: family 页（家庭管理）

**Files:**
- Create: `miniprogram/pages/family/family.js`
- Create: `miniprogram/pages/family/family.wxml`
- Create: `miniprogram/pages/family/family.wxss`
- Create: `miniprogram/pages/family/family.json`

- [ ] **Step 1: 创建 family.json**

```json
{
  "navigationBarTitleText": "家庭"
}
```

- [ ] **Step 2: 创建 family.wxml**

```xml
<!-- miniprogram/pages/family/family.wxml -->
<view class="container">
  <view wx:if="{{loading}}" class="empty-tip">加载中…</view>

  <view wx:elif="{{family}}">
    <!-- 家庭名称 -->
    <view class="family-name">{{family.name}}</view>

    <!-- 成员列表 -->
    <view class="card">
      <view class="section-label">家庭成员</view>
      <view class="member-item" wx:for="{{family.members}}" wx:key="id">
        <view class="member-avatar">{{item.nickname.slice(0,1)}}</view>
        <view class="member-name">{{item.nickname}}</view>
        <view class="member-role {{item.role === 'owner' ? 'role-owner' : 'role-member'}}">
          {{item.role === 'owner' ? '创建者' : '成员'}}
        </view>
      </view>
    </view>

    <!-- 邀请码（仅 owner） -->
    <view class="card" wx:if="{{isOwner}}">
      <view class="section-label">邀请家人加入</view>
      <view class="invite-desc">生成邀请码后，将 6 位码告诉家人，他们在小程序输入即可加入。</view>

      <button
        class="btn-primary"
        loading="{{generating}}"
        disabled="{{generating}}"
        bindtap="onGenerate"
      >生成邀请码</button>

      <view wx:if="{{invite}}" class="invite-result">
        <view class="invite-code" bindlongpress="onCopyCode">{{invite.code}}</view>
        <view class="invite-expire">有效期至 {{invite.expireTime}}</view>
        <view class="invite-copy-hint">长按邀请码可复制</view>
      </view>
    </view>
  </view>

  <!-- 退出提示（小程序不支持退出账号，说明一下） -->
  <view class="logout-note">账号绑定微信，打开小程序自动登录</view>
</view>
```

- [ ] **Step 3: 创建 family.wxss**

```css
/* miniprogram/pages/family/family.wxss */
.container { padding: 24rpx; }

.family-name {
  font-size: 40rpx;
  font-weight: 700;
  color: #15803d;
  margin-bottom: 24rpx;
}
.section-label {
  font-size: 26rpx;
  color: #6b7280;
  margin-bottom: 16rpx;
}

/* 成员列表 */
.member-item {
  display: flex;
  align-items: center;
  padding: 16rpx 0;
  border-bottom: 1rpx solid #f0fdf4;
}
.member-item:last-child { border-bottom: none; }
.member-avatar {
  width: 64rpx;
  height: 64rpx;
  border-radius: 50%;
  background: #d1fae5;
  color: #16a34a;
  font-size: 28rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16rpx;
  flex-shrink: 0;
}
.member-name { flex: 1; font-size: 30rpx; }
.member-role {
  font-size: 22rpx;
  padding: 4rpx 12rpx;
  border-radius: 20rpx;
}
.role-owner { background: #d1fae5; color: #15803d; }
.role-member { background: #f3f4f6; color: #6b7280; }

/* 邀请码区域 */
.invite-desc {
  font-size: 26rpx;
  color: #6b7280;
  margin-bottom: 24rpx;
  line-height: 1.6;
}
.invite-result {
  margin-top: 32rpx;
  background: #f0fdf4;
  border: 1rpx solid #d1fae5;
  border-radius: 12rpx;
  padding: 32rpx;
  text-align: center;
}
.invite-code {
  font-size: 72rpx;
  font-weight: 700;
  letter-spacing: 16rpx;
  color: #16a34a;
  font-family: monospace;
  margin-bottom: 12rpx;
}
.invite-expire { font-size: 26rpx; color: #6b7280; margin-bottom: 8rpx; }
.invite-copy-hint { font-size: 22rpx; color: #9ca3af; }

.logout-note {
  text-align: center;
  font-size: 24rpx;
  color: #9ca3af;
  margin-top: 48rpx;
}
```

- [ ] **Step 4: 创建 family.js**

```javascript
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
```

- [ ] **Step 5: Commit**

```bash
cd /Users/zyb/Own/code/jojoshuai/childcare
git add miniprogram/pages/add/ miniprogram/pages/family/
git commit -m "feat(mp): add add page and family page"
```

---

## Chunk 4: chart 页 + 最终验证

### Task 9: chart 页（生长曲线图）

**Files:**
- Create: `miniprogram/pages/chart/chart.js`
- Create: `miniprogram/pages/chart/chart.wxml`
- Create: `miniprogram/pages/chart/chart.wxss`
- Create: `miniprogram/pages/chart/chart.json`

**图表策略：**
- X 轴 categories = 所有有测量记录的月龄 + WHO 月龄（0,6,12,...,60）的并集，排序去重
- 孩子数据：稀疏数组，无记录的月龄为 null
- WHO 数据：在 categories 月龄对应位置取值，超出 60 月龄不传
- wx-charts 对 null 值不绘制数据点，连接两侧已有点

- [ ] **Step 1: 创建 chart.json**

```json
{
  "navigationBarTitleText": "生长曲线"
}
```

- [ ] **Step 2: 创建 chart.wxml**

```xml
<!-- miniprogram/pages/chart/chart.wxml -->
<view class="container">
  <!-- 顶部信息 -->
  <view class="child-header">
    <text class="child-name">{{childName}}</text>
    <text class="child-age">{{childAge}}</text>
  </view>

  <!-- 类型 Tab -->
  <view class="tab-row">
    <view
      class="tab-item {{type === 'weight' ? 'tab-item--active' : ''}}"
      bindtap="onTypeChange"
      data-type="weight"
    >体重</view>
    <view
      class="tab-item {{type === 'height' ? 'tab-item--active' : ''}}"
      bindtap="onTypeChange"
      data-type="height"
    >身高</view>
    <view
      class="tab-item {{type === 'head_circumference' ? 'tab-item--active' : ''}}"
      bindtap="onTypeChange"
      data-type="head_circumference"
    >头围</view>
  </view>

  <!-- 图表区 -->
  <view class="chart-card card">
    <view wx:if="{{chartLoading}}" class="chart-loading">加载中…</view>
    <view wx:elif="{{measurements.length === 0}}" class="chart-empty">
      <view>暂无{{typeLabelMap[type]}}记录</view>
      <view style="font-size:24rpx;margin-top:8rpx;color:#9ca3af">去"录入"Tab 添加记录</view>
    </view>
    <view wx:else>
      <canvas canvas-id="growthChart" class="chart-canvas" style="width:{{chartWidth}}px;height:250px;"></canvas>
      <view wx:if="{{!showWHO}}" class="who-tip">WHO 参考数据覆盖范围为 0-60 个月</view>
    </view>
  </view>

  <!-- 记录列表 -->
  <view class="card" wx:if="{{measurements.length > 0}}">
    <view class="section-label">历史记录</view>
    <view class="record-item" wx:for="{{measurements}}" wx:key="id">
      <view class="record-date">{{item.measured_at}}</view>
      <view class="record-value">{{item.value}} {{unitMap[type]}}</view>
      <view class="record-del" bindtap="onDelete" data-id="{{item.id}}">🗑</view>
    </view>
  </view>
</view>
```

- [ ] **Step 3: 创建 chart.wxss**

```css
/* miniprogram/pages/chart/chart.wxss */
.container { padding: 24rpx; }

/* 顶部 */
.child-header {
  display: flex;
  align-items: baseline;
  gap: 16rpx;
  margin-bottom: 24rpx;
}
.child-name { font-size: 40rpx; font-weight: 700; color: #15803d; }
.child-age { font-size: 28rpx; color: #6b7280; }

/* Tab */
.tab-row {
  display: flex;
  background: #fff;
  border-radius: 12rpx;
  margin-bottom: 16rpx;
  overflow: hidden;
}
.tab-item {
  flex: 1;
  text-align: center;
  height: 80rpx;
  line-height: 80rpx;
  font-size: 28rpx;
  color: #6b7280;
}
.tab-item--active {
  background: #16a34a;
  color: #fff;
  font-weight: 600;
}

/* 图表 */
.chart-card { padding: 16rpx; }
.chart-canvas { display: block; }
.chart-loading, .chart-empty {
  height: 250px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #9ca3af;
  font-size: 28rpx;
}
.who-tip {
  text-align: center;
  font-size: 22rpx;
  color: #9ca3af;
  margin-top: 12rpx;
}

/* 记录列表 */
.section-label { font-size: 26rpx; color: #6b7280; margin-bottom: 16rpx; }
.record-item {
  display: flex;
  align-items: center;
  padding: 16rpx 0;
  border-bottom: 1rpx solid #f0fdf4;
}
.record-item:last-child { border-bottom: none; }
.record-date { flex: 1; font-size: 28rpx; color: #374151; }
.record-value { font-size: 30rpx; font-weight: 600; color: #15803d; margin-right: 24rpx; }
.record-del { font-size: 32rpx; padding: 8rpx; }
```

- [ ] **Step 4: 创建 chart.js**

```javascript
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

    // 计算每条测量记录的月龄
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
      { name: '孩子',   data: childValues, color: '#16a34a', format: v => v !== null ? String(v) : '' },
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
```

- [ ] **Step 5: 在微信开发者工具中构建 npm**

打开微信开发者工具，项目根目录选 `miniprogram/`，然后：

1. 菜单栏 → **工具** → **构建 npm**
2. 控制台出现 `npm packages built` 即成功
3. 确认 `miniprogram/miniprogram_npm/wx-charts/` 目录已生成

- [ ] **Step 6: Commit**

```bash
cd /Users/zyb/Own/code/jojoshuai/childcare
git add miniprogram/pages/chart/
git commit -m "feat(mp): add chart page with wx-charts growth curve"
```

---

### Task 10: 最终验证 + commit

- [ ] **Step 1: 在微信开发者工具中导入项目**

1. 打开微信开发者工具
2. 新建项目 → AppID 填写测试号 AppID（或点"测试号"按钮获取）
3. 项目目录选 `childcare/miniprogram/`
4. 后端启动：`cd backend && env $(cat .env | grep -v '^#' | xargs) go run .`

**注意：** 开发者工具中需勾选"不校验合法域名、web-view（业务域名）、TLS 版本以及 HTTPS 证书"，否则 http 请求会被拦截。

- [ ] **Step 2: 修改 utils/request.js 中的 BASE_URL**

将 `http://192.168.1.100:8080` 改为你本机在局域网中的实际 IP：

```bash
# 查看本机 IP
ifconfig | grep 'inet ' | grep -v '127.0.0.1'
```

然后编辑 `miniprogram/utils/request.js` 第 4 行的 `BASE_URL`。

- [ ] **Step 3: 手动验证核心流程**

- [ ] 首次打开小程序 → 静默登录 → 跳转 join 页（如果 family_id 为空）
- [ ] join 页输入邀请码 → 成功跳首页
- [ ] 首页显示孩子列表，月龄正确
- [ ] 下拉刷新 → 数据更新
- [ ] 点孩子 → 进入 chart 页，图表正确渲染，WHO 参考线显示
- [ ] 切换体重/身高/头围 Tab → 图表刷新
- [ ] 删除记录 → 二次确认，图表更新
- [ ] 录入 Tab：选孩子、选类型、输数值、选日期、提交 → Toast 显示"记录成功"
- [ ] 家庭 Tab：显示成员列表
- [ ] owner 身份 → 生成邀请码 → 长按复制

- [ ] **Step 4: 最终 commit**

```bash
cd /Users/zyb/Own/code/jojoshuai/childcare
git add miniprogram/
git commit -m "feat(mp): complete miniprogram MVP"
```

- [ ] **Step 5: 更新 dialog.md**

在 `docs/dialog.md` 追加本次对话记录：小程序 brainstorm → spec → 实现计划完成。
