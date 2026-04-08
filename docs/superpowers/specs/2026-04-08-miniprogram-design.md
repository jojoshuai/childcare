# 微信小程序设计 Spec

**Date:** 2026-04-08
**Scope:** MVP 微信小程序 — 快速录入成长数据 + 生长曲线查看 + 家庭共享

---

## 1. 目标

为家庭成员（尤其老人）提供手机端微信小程序，实现：
- 微信一键登录（无账号密码）
- 通过邀请码加入家庭
- 快速录入体重/身高/头围
- 查看叠加 WHO P3/P50/P97 的生长曲线图
- 家庭管理（查看成员、生成邀请码）

---

## 2. 技术栈

| 项目 | 选型 |
|------|------|
| 框架 | 微信原生（WXML/WXSS/JS） |
| 图表 | wx-charts |
| 状态 | app.js globalData |
| HTTP | wx.request 封装（utils/request.js） |
| 目录 | `childcare/miniprogram/` |

---

## 3. 视觉风格

- **主色**：绿色 `#16a34a`
- **背景**：浅绿 `#f0fdf4`
- **按钮/输入框**：尺寸放大，便于老人操作
- **语感**：清新自然，贴近"成长"主题

---

## 4. 目录结构

```
miniprogram/
├── app.js              # 全局：globalData（token/user）+ 启动静默登录
├── app.json            # tabBar 配置 + 页面注册
├── app.wxss            # 全局样式
├── utils/
│   ├── request.js      # wx.request 封装：自动附 token，401 自动刷新
│   └── util.js         # 月龄计算等工具函数
├── pages/
│   ├── index/          # Tab1：首页，孩子列表
│   │   ├── index.js
│   │   ├── index.wxml
│   │   └── index.wxss
│   ├── add/            # Tab2：快速录入（核心）
│   │   ├── add.js
│   │   ├── add.wxml
│   │   └── add.wxss
│   ├── family/         # Tab3：家庭管理
│   │   ├── family.js
│   │   ├── family.wxml
│   │   └── family.wxss
│   ├── chart/          # 生长曲线图（非 Tab，从首页跳转）
│   │   ├── chart.js
│   │   ├── chart.wxml
│   │   └── chart.wxss
│   └── join/           # 输入邀请码（非 Tab，无 family_id 时强制跳转）
│       ├── join.js
│       ├── join.wxml
│       └── join.wxss
└── miniprogram_npm/    # wx-charts（npm 构建后生成）
```

---

## 5. 启动流程

每次启动都重新执行静默登录，后端根据 openid 识别用户，因此无需在本地判断 token 是否过期：

```
app.js onLaunch
  → wx.login() 拿 code
  → POST /api/auth/wx-login { "code": "<code>" }
  → 后端返回 { token, refresh_token, user }
  → 存入 globalData + wx.setStorageSync
  → user.family_id === null → wx.reLaunch 跳 /pages/join/join
  → user.family_id 有值 → 正常进入 tabBar（首页）
```

**Token 策略：**
- 每次启动执行 wx-login，发行新 token，所以 token 过期几乎不会成为问题
- `request.js` 仍保留 401 拦截：先尝试 `POST /api/auth/refresh`，失败则重新走 wx-login 流程
- `globalData.user` 仅在 wx-login 成功时更新，`/api/auth/refresh` 不返回 user 对象，刷新后 user 保持不变

---

## 6. API 响应字段

后端所有接口均返回 snake_case：

| Go 字段 | JSON 字段 |
|---------|-----------|
| `FamilyID` | `family_id` |
| `BirthDate` | `birth_date` |
| `MeasuredAt` | `measured_at` |
| `CreatedAt` | `created_at` |

**`GET /api/children` 返回字段：**
```json
[
  {
    "id": "uuid",
    "family_id": "uuid",
    "name": "小明",
    "gender": "male",
    "birth_date": "2023-01-15",
    "created_at": "2026-04-08T10:00:00Z"
  }
]
```

chart 页需要 `birth_date`（月龄计算）和 `gender`（WHO 查询参数），均在此响应中。

---

## 7. 页面设计

### 7.1 join 页（加入家庭）

触发条件：`user.family_id === null`

- 说明文字："请输入家长提供的 6 位邀请码"
- 大号输入框，`maxlength="6"`，提交前调用 `.toUpperCase()`（后端邀请码全为大写字母和数字）
- 请求：`POST /api/family/join { "code": "AB1234" }`
- 成功：后端返回 `{ "family_id": "uuid" }`，手动更新 `globalData.user.family_id`，然后 `wx.reLaunch` 跳首页
- 失败：`wx.showToast` 显示错误（邀请码无效/过期/已使用）

### 7.2 index 页（首页）

- onShow 时调用 `GET /api/children`，响应字段见第 6 节
- 孩子列表，每行：头像（👦 male / 👧 female）+ 姓名 + 月龄
- 点击 → `wx.navigateTo('/pages/chart/chart?id=<childId>')`
- 空列表：提示"请在 Web 端添加孩子"
- 下拉刷新：`enablePullDownRefresh: true`，onPullDownRefresh 重新拉接口后 `wx.stopPullDownRefresh()`

### 7.3 add 页（快速录入，Tab 核心页）

1. **选孩子**：onShow 时拉 `GET /api/children`；单个孩子默认选中不显示选择器；多个孩子显示 `picker`
2. **选类型**：三个大色块按钮（体重 / 身高 / 头围），选中绿色高亮
3. **输数值**：`input type="digit"`，弹数字键盘；placeholder 显示参考范围（体重 0.5–200 kg，身高 20–250 cm，头围 20–80 cm）
4. **选日期**：`picker mode="date"`，默认今天（`new Date().toISOString().slice(0,10)`）
5. **备注**：选填，`input` 单行文本
6. **提交**：`POST /api/children/:id/measurements { type, value, measured_at, note }`
   - `note` 为空字符串时发送 `null`，与后端 `*string` 类型对齐
   - 成功：`wx.showToast({ title: '记录成功', icon: 'success' })`，重置数值/日期/备注（保留孩子和类型选择）
   - 失败：`wx.showToast({ title: err.message, icon: 'none' })`

### 7.4 chart 页（生长曲线图）

入口：`wx.navigateTo('/pages/chart/chart?id=<childId>')`，页面从 `options.id` 取 childId。

**数据加载：**
- onLoad 时从 `globalData.children`（或重新调 `GET /api/children`）取当前孩子的 `birth_date` 和 `gender`
- 切换类型时并行调用：
  - `GET /api/children/:id/measurements?type=<type>` → 孩子测量数据
  - `GET /api/who-standards?gender=<gender>&type=<type>` → WHO 参考数据（返回 `{ data: [{month, p3, p50, p97}] }`）

**图表（wx-charts）：**
- 孩子数据：绿色 `#16a34a` 实线 + 圆点
- WHO P3：灰色 `#d1d5db` 虚线
- WHO P50：灰色 `#9ca3af` 虚线（稍深）
- WHO P97：灰色 `#d1d5db` 虚线
- X 轴：月龄（整数，`Math.floor(diffDays / 30.4375)`）
- Y 轴：数值（体重 kg，身高/头围 cm）
- 月龄 ≥ 61 时：不传 WHO 数据给图表，图表下方显示"WHO 参考数据覆盖范围为 0-60 个月"

**记录列表：**
- 按 `measured_at` 倒序排列
- 每行：日期 + 数值 + 删除图标
- 删除：`wx.showModal` 二次确认，确认后 `DELETE /api/children/:id/measurements/:mid`，成功后重新加载数据

### 7.5 family 页（家庭管理，Tab）

- onShow 时调用 `GET /api/family` → `{ id, name, members: [{id, nickname, role}] }`
- 显示家庭名称 + 成员列表（昵称 + 角色：创建者/成员）
- **Owner 专属**：生成邀请码按钮
  - 调 `POST /api/family/invite` → `{ code, expires_at }`
  - `expires_at` 为 RFC3339 格式（如 `"2026-04-09T10:30:00Z"`），显示时转本地时间：`new Date(expires_at).toLocaleTimeString([], {hour:'2-digit', minute:'2-digit'})`
  - 显示 6 位码 + "有效期至 HH:mm"
  - 长按邀请码调 `wx.setClipboardData` 复制

---

## 8. API 层（utils/request.js）

```javascript
// baseURL 配置
const BASE_URL = 'http://192.168.x.x:8080'  // 开发环境（局域网 IP）
// 生产环境改为微信云托管内网地址

function request({ url, method = 'GET', data } = {}) {
  return new Promise((resolve, reject) => {
    const app = getApp()
    wx.request({
      url: BASE_URL + url,
      method,
      data,
      header: { Authorization: `Bearer ${app.globalData.token}` },
      success(res) {
        if (res.statusCode === 401) {
          // 尝试刷新 token
          refreshAndRetry({ url, method, data }).then(resolve).catch(reject)
        } else if (res.statusCode >= 400) {
          reject(res.data)
        } else {
          resolve(res.data)
        }
      },
      fail(err) { reject(err) }
    })
  })
}
```

401 处理：先调 `POST /api/auth/refresh { refresh_token }`，成功则更新 token 并重试原请求；失败则重新执行 wx-login 静默登录流程。

---

## 9. 月龄计算（utils/util.js）

```javascript
function calcAgeMonths(birthDate) {
  const diffMs = Date.now() - new Date(birthDate).getTime()
  const diffDays = diffMs / (1000 * 60 * 60 * 24)
  return Math.floor(diffDays / 30.4375)
}

function ageLabel(birthDate) {
  const months = calcAgeMonths(birthDate)
  if (months < 12) return `${months}个月`
  const years = Math.floor(months / 12)
  const rem = months % 12
  return rem > 0 ? `${years}岁${rem}个月` : `${years}岁`
}
```

月龄 ≥ 61 时不显示 WHO 参考线。

---

## 10. 全局状态（app.js globalData）

```javascript
globalData: {
  token: null,          // access token（每次启动 wx-login 后更新）
  refreshToken: null,   // refresh token
  user: null,           // { id, nickname, family_id, role }（仅 wx-login 时更新）
  children: [],         // 孩子列表缓存（首页/录入页加载后存入）
}
```

---

## 11. MVP 范围外

- 小程序端添加/删除孩子（仅 Web 端支持）
- 图表内直接添加记录（添加走 add Tab）
- 图表编辑记录（chart 页只支持查看和删除）
- 响应式适配平板
- 深色模式
- 数据导出
