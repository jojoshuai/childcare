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

```
app.js onLaunch
  → 检查 storage 中是否有 token
  → wx.login() 拿 code
  → POST /api/auth/wx-login { code }
  → 存 token / refresh_token / user 到 globalData + wx.setStorageSync
  → user.family_id === null → wx.reLaunch 跳 /pages/join/join
  → user.family_id 有值 → 正常进入 tabBar（首页）
```

Token 策略：
- Access token 7 天有效，refresh token 30 天有效
- 每次启动重新静默登录（wx-login code 每次都会变），后端根据 openid 识别用户

---

## 6. 页面设计

### 6.1 join 页（加入家庭）

触发条件：`user.family_id === null`

- 说明文字："请输入家长提供的 6 位邀请码"
- 大号输入框，`maxlength="6"`，自动大写
- 提交按钮 → `POST /api/family/join { code }`
- 成功：更新 globalData.user，跳首页
- 失败：显示错误提示（邀请码无效/过期/已使用）

### 6.2 index 页（首页）

- 孩子列表，每行：头像（👦/👧）+ 姓名 + 月龄
- 点击 → 跳转 `/pages/chart/chart?id=<childId>`
- 空列表：提示"请在 Web 端添加孩子"
- 下拉刷新支持

### 6.3 add 页（快速录入，Tab 核心页）

1. **选孩子**：单个孩子时默认选中不显示选择器；多个孩子时显示 picker
2. **选类型**：三个大色块按钮（体重 / 身高 / 头围），选中绿色高亮
3. **输数值**：`input type="digit"`，弹数字键盘；placeholder 显示参考范围
4. **选日期**：`picker mode="date"`，默认今天
5. **备注**：选填，`input` 单行文本
6. **提交**：`POST /api/children/:id/measurements`
   - 成功：`wx.showToast({ title: '记录成功' })`，表单重置（保留孩子和类型选择）
   - 失败：`wx.showToast({ title: err.message, icon: 'none' })`

### 6.4 chart 页（生长曲线图）

入口：从 index 页点孩子进入，URL 携带 `?id=<childId>`

- **顶部**：孩子姓名 + 月龄
- **类型 Tab**：体重 / 身高 / 头围（切换重新拉数据）
- **图表区**：wx-charts 折线图
  - 孩子数据：绿色实线 + 圆点
  - WHO P3：灰色虚线
  - WHO P50：灰色虚线（较深）
  - WHO P97：灰色虚线
  - 月龄 ≥ 61 时：隐藏 WHO 线，图表下方显示提示文字
  - X 轴：月龄（整数），Y 轴：数值（kg 或 cm）
- **记录列表**：按 measured_at 倒序，每行显示日期 + 数值 + 删除按钮
  - 删除：二次确认 `wx.showModal`，确认后 `DELETE /api/children/:id/measurements/:mid`

### 6.5 family 页（家庭管理，Tab）

- **家庭名称**
- **成员列表**：昵称 + 角色（创建者/成员）
- **Owner 专属**：生成邀请码按钮
  - 显示 6 位码 + 有效期（`有效期至 HH:mm`）
  - 长按可复制邀请码

---

## 7. API 层（utils/request.js）

```javascript
// 自动附加 token，401 时刷新后重试，刷新失败重新静默登录
request({ url, method, data }) → Promise
```

- Request：自动附加 `Authorization: Bearer <token>`
- Response 401：调 `POST /api/auth/refresh`，成功则重试，失败则重新 `wx.login()` 走静默登录流程
- 所有接口 baseURL：开发环境用本地 IP，生产环境用微信云托管内网地址

---

## 8. 月龄计算

与 Web 端一致：
```
ageMonths = Math.floor(diffDays / 30.4375)
```
其中 `diffDays = (today - birthDate) 的天数差`。月龄 ≥ 61 时不显示 WHO 参考线。

---

## 9. 全局状态（app.js globalData）

```javascript
globalData: {
  token: null,          // access token
  refreshToken: null,   // refresh token
  user: null,           // { id, nickname, family_id, role }
  children: [],         // 孩子列表缓存（首页加载后存入）
}
```

---

## 10. MVP 范围外

- 小程序端添加/删除孩子（仅 Web 端支持）
- 图表内直接添加记录（添加走 add Tab）
- 图表编辑记录（chart 页只支持查看和删除）
- 响应式适配平板
- 深色模式
- 数据导出
