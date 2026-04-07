# Web 前端设计 Spec

**Date:** 2026-04-08
**Scope:** MVP Web 端 — React SPA，供家庭内部通过 IP 访问

---

## 1. 目标

为家长提供桌面端 Web 界面，实现：
- 账号登录 / 注册
- 查看孩子列表和成长数据
- 录入/编辑/删除测量记录
- 查看叠加 WHO 百分位线的生长曲线图
- 家庭管理（邀请码生成）

---

## 2. 技术栈

| 项目 | 选型 |
|------|------|
| 框架 | React 18 + TypeScript + Vite |
| UI | Ant Design 5 |
| 图表 | Recharts |
| 路由 | React Router v6 |
| HTTP | Axios |
| 状态 | React Context（AuthContext） |
| 目录 | `childcare/web/` |

---

## 3. 视觉风格

- **主色**：绿色 `#16a34a`
- **背景**：浅绿 `#f0fdf4`
- **布局**：左侧固定导航栏 + 右侧内容区
- **语感**：清新自然，贴近"成长"主题

---

## 4. API 响应字段名

后端 model 结构体已添加 json tags，所有接口响应均使用 **snake_case**：

| Go 字段 | JSON 字段 |
|---------|-----------|
| `ID` | `id` |
| `FamilyID` | `family_id` |
| `ChildID` | `child_id` |
| `BirthDate` | `birth_date` |
| `MeasuredAt` | `measured_at` |
| `CreatedAt` | `created_at` |
| `CreatedBy` | `created_by` |

**注意：**
- `POST /api/auth/refresh` 只返回 `{ token, refresh_token }`，**不**返回 `user` 对象。刷新后 `AuthContext` 中的 `user` 保持不变（仍来自 localStorage）。
- 测量记录的 `note` 字段可为 null。Drawer 中备注为空时，前端发送 `note: null`（不发送空字符串），与后端 `*string` 类型对齐。

---

## 5. 页面结构

### 5.1 `/login` — 登录 / 注册

- 居中卡片，Tab 切换登录 / 注册
- 登录：用户名 + 密码
- 注册：用户名 + 密码 + 家庭名称 + 昵称
- 成功后写入 `localStorage`（token + refresh_token + user），跳转 `/dashboard`

### 5.2 `/dashboard` — 首页

- 左侧导航栏：Logo、首页、孩子们、家庭
- 主区：孩子列表（列表行）
  - 每行：头像（👦 male / 👧 female）+ 姓名 + 月龄 + 最近一次测量摘要 + 箭头
  - 底部：添加孩子行（打开 Modal，填写姓名/性别/生日）
  - 删除孩子：每行 hover 显示删除图标（仅 owner），确认后删除
- 加载中：Ant Design Skeleton 骨架屏
- 请求失败：`message.error('加载失败，请刷新重试')`

### 5.3 `/children/:id` — 孩子详情页（核心）

- 顶部：孩子姓名 + 月龄
- 顶部 Tab：体重 / 身高 / 头围
- **左右分栏**（左约 60%，右约 40%）：

**左侧 — Recharts 折线图：**
- X 轴：月龄（整数，`floor((measured_at - birth_date) / 30.4375)`，单位：月）
- Y 轴：数值（体重 kg，身高/头围 cm）
- 孩子数据：绿色实线 + 圆形数据点，Tooltip hover 显示 `日期 + 数值`
- WHO P3/P50/P97：灰色虚线，图例标注
- 月龄 ≥ 61 时：隐藏 WHO 线，图表底部显示提示文字"WHO 参考数据覆盖范围为 0-60 个月"
- WHO 数据来源：`GET /api/who-standards?gender=<gender>&type=<type>`（需已登录且有 family_id）

**右侧 — 测量记录列表：**
- 按 `measured_at` 倒序排列
- 每行：日期 + 数值 + 编辑图标 + 删除图标
- 顶部"+ 添加"按钮，点击打开右侧 Drawer

**右侧 Drawer（添加/编辑）：**
- 字段：类型（只读，来自当前 Tab）+ 日期（默认今天）+ 数值 + 备注（选填）
- 保存成功后：关闭 Drawer，**重新从 API 拉取完整列表**（`GET /api/children/:id/measurements?type=<type>`）
- 后端校验失败（400）：在 Drawer 内显示 `message.error(err.message)`
- 加载/提交中：按钮显示 loading 状态

### 5.4 `/family` — 家庭管理

- 家庭名称
- 成员列表（昵称 + 角色 owner/member）
- Owner 专属：生成邀请码按钮
  - 点击后显示 6 位码 + 到期时间（静态显示，格式：`有效期至 HH:mm`，使用浏览器本地时间 `new Date(expires_at).toLocaleTimeString()`）
  - 无倒计时动画（MVP 不需要）

---

## 6. 路由守卫

- 未登录访问受保护页面 → 重定向 `/login`
- Web 用户注册时必然已有家庭，无需处理 `family_id === null` 的情况

---

## 7. API 层

`src/api/` 下按模块拆分：

```
axios.ts         — Axios 实例 + 拦截器（含 token 刷新逻辑）
auth.ts          — register / login（refresh 由 axios.ts 内部调用，不对外暴露）
children.ts      — list / create / update / delete
measurements.ts  — list(type?) / create / update / delete
family.ts        — getFamily / generateInvite
who.ts           — getWHOStandards(gender, type)
```

**Axios 拦截器（axios.ts）：**
- Request：自动附加 `Authorization: Bearer <token>`
- Response 401：直接调用 `POST /api/auth/refresh`（不通过 auth.ts），成功则重试原请求，失败则清除 localStorage 并跳转 `/login`

**注意：** `family.ts` 不暴露 `joinFamily`（Web 端注册即建家庭，不支持加入已有家庭）。

---

## 8. 状态管理

`AuthContext` 提供：
- `user`、`token`（从 localStorage 初始化）
- `login(token, refreshToken, user)`：写 localStorage + 更新 Context
- `logout()`：清 localStorage + 跳转 `/login`

其余数据（孩子列表、测量记录）在各页面组件内用 `useState` + `useEffect` 管理，不引入全局状态库。

---

## 9. 月龄计算

图表 X 轴和 WHO 参考线截断均使用以下公式：

```
ageMonths = floor((measuredAt - birthDate) / 30.4375)
```

其中时间差单位为天（`Math.floor(diffDays / 30.4375)`）。月龄 ≥ 61 时不显示 WHO 参考线。

---

## 10. 文件结构

```
web/
├── index.html
├── vite.config.ts
├── package.json
└── src/
    ├── main.tsx
    ├── App.tsx              # 路由配置 + AuthContext Provider
    ├── api/
    │   ├── axios.ts         # Axios 实例 + 拦截器（含 refresh 逻辑）
    │   ├── auth.ts          # register / login
    │   ├── children.ts
    │   ├── measurements.ts
    │   ├── family.ts        # getFamily / generateInvite
    │   └── who.ts
    ├── context/
    │   └── AuthContext.tsx
    ├── components/
    │   ├── Sidebar.tsx
    │   ├── GrowthChart.tsx
    │   └── MeasurementDrawer.tsx
    └── pages/
        ├── Login.tsx
        ├── Dashboard.tsx
        ├── ChildDetail.tsx
        └── Family.tsx
```

---

## 11. MVP 范围外

- 响应式/移动端适配（Web 仅桌面）
- 国际化
- 深色模式
- 数据导出
- 邀请码倒计时动画
- 编辑孩子信息（仅支持删除）
