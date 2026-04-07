# Childcare Growth Tracker — Design Spec

**Date:** 2026-04-07
**Scope:** MVP — 身高/体重/头围记录 + 生长曲线图 + Web + 微信小程序 + 家庭共享

---

## 1. 产品目标

为家长提供一个私有化部署的孩子成长数据记录工具，支持：
- 记录孩子的身高、体重、头围数据
- 生成叠加 WHO 标准百分位线的生长曲线图
- Web 端查看图表，微信小程序快速录入
- 家庭成员（包括老人）通过邀请码加入共享

---

## 2. 整体架构

```
┌─────────────────┐         ┌─────────────────┐
│   Web (React)   │         │   微信小程序     │
│  IP地址本地访问  │         │  手机快速录入    │
└────────┬────────┘         └────────┬────────┘
         │ HTTP（IP直连）             │ 微信内网专线
         │                           │（无需域名/备案）
         └──────────────┬────────────┘
                  ┌─────▼──────┐
                  │  Go + Gin  │
                  │   后端服务  │
                  └─────┬──────┘
                  ┌─────▼──────┐
                  │ PostgreSQL │
                  │   数据库    │
                  └────────────┘
             （部署在微信云托管 CloudRun）
```

**技术栈：**
- 后端：Go + Gin + PostgreSQL，Docker 镜像部署到微信云托管
- Web：React + TypeScript + Recharts + Ant Design，通过 IP 地址访问（局域网/本地）
- 小程序：微信原生（WXML/JS）+ wx-charts，通过微信内网专线调用后端，无需域名备案
- 认证：JWT（Web 账号密码 / 小程序微信 openid）

**开发阶段：** 使用小程序测试号，跳过域名校验，专注功能开发。
**上线阶段：** 注册正式小程序号，后端部署到微信云托管，小程序无需备案即可上线；Web 端通过 IP 访问，暂不备案。

---

## 3. 数据模型

### families
| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| name | string | 家庭名称 |
| created_at | timestamp | 创建时间 |

### users
| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| family_id | UUID | 关联家庭（可为空：小程序用户注册后未加入家庭时为空） |
| username | string | Web 登录用（可为空） |
| password_hash | string | Web 密码哈希（可为空） |
| wx_openid | string | 微信 openid（可为空） |
| nickname | string | 显示名称 |
| role | enum | owner / member（无家庭时为空） |
| created_at | timestamp | 创建时间 |

### children
| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| family_id | UUID | 关联家庭 |
| name | string | 孩子姓名 |
| gender | enum | male / female |
| birth_date | date | 出生日期 |
| created_at | timestamp | 创建时间 |

### measurements
| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| child_id | UUID | 关联孩子 |
| type | enum | weight / height / head_circumference |
| value | float | 数值（kg / cm） |
| measured_at | date | 测量日期 |
| note | string | 备注（可为空） |
| created_by | UUID | 录入用户 |
| created_at | timestamp | 创建时间 |

### invite_codes
| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| family_id | UUID | 关联家庭 |
| code | string | 6位邀请码 |
| expires_at | timestamp | 过期时间（24小时） |
| used | boolean | 是否已使用（单次有效，使用后不可复用） |
| created_by | UUID | 生成邀请码的用户 ID |
| created_at | timestamp | 创建时间 |

---

## 4. API 设计

### 错误响应格式

所有接口错误统一返回如下结构：
```json
{
  "code": "INVALID_INVITE_CODE",
  "message": "邀请码无效或已过期"
}
```
HTTP 状态码遵循语义：400 参数错误、401 未登录、403 无权限、404 资源不存在、500 服务器错误。

### 认证

`POST /api/auth/register` — Web 注册。原子性地创建用户 + 家庭，用户角色为 `owner`，`family_id` 在创建时赋值。请求体：`{username, password, family_name, nickname}`，返回 `{token, refresh_token, user}`。**Web 注册始终创建新家庭**，Web 用户不支持加入已有家庭（仅小程序通过邀请码加入）。

`POST /api/auth/login` — Web 登录。返回同上。

`POST /api/auth/wx-login` — 小程序登录。传微信 `code`，后端换取 `openid`。若 openid 已存在则登录，否则创建新用户（此时 `family_id` 为空，用户需通过邀请码加入家庭后才能操作数据）。返回 `{token, refresh_token, user}`。

`POST /api/auth/refresh` — 刷新 token。Access token 有效期 7 天，refresh token 有效期 30 天。refresh token 过期后用户需重新登录。小程序每次启动检查 token 是否即将过期，自动刷新。

**所有三个认证接口返回的 `user` 对象结构：**
```json
{
  "id": "uuid",
  "nickname": "张三",
  "family_id": "uuid 或 null",
  "role": "owner 或 member 或 null"
}
```
小程序可通过 `family_id === null` 判断用户尚未加入家庭，引导进入邀请码页面。

**未加入家庭的用户（`family_id` 为空）调用任何数据接口时，返回 403：**
```json
{ "code": "NO_FAMILY_JOINED", "message": "请先通过邀请码加入家庭" }
```

### 家庭 & 邀请

权限规则：仅 `owner` 可生成邀请码；`owner` 和 `member` 均可查看家庭信息和录入数据。

```
GET  /api/family                 # 获取家庭信息及成员列表（需已加入家庭）
POST /api/family/invite          # 生成邀请码（仅 owner，6位，24小时有效）
POST /api/family/join            # 用邀请码加入家庭（小程序用户初次加入）
                                 # 错误码：INVITE_CODE_NOT_FOUND（不存在）
                                 #         INVITE_CODE_EXPIRED（已过期）
                                 #         INVITE_CODE_ALREADY_USED（已被使用）
```

### 孩子

`owner` 和 `member` 均可添加、编辑孩子信息；仅 `owner` 可删除孩子（删除会级联删除该孩子所有测量记录）。

```
GET    /api/children             # 获取家庭下所有孩子
POST   /api/children             # 添加孩子
PUT    /api/children/:id         # 编辑孩子信息
DELETE /api/children/:id         # 删除孩子（仅 owner，级联删除测量记录）
```

### 测量记录

MVP 返回全量记录，不分页（预计单孩子 5 年内数据量 < 200 条，可接受）。支持 `?type=` 筛选；`type` 枚举值统一使用英文：`weight` / `height` / `head_circumference`。

当孩子月龄超过 60 个月（5岁）时，图表不显示 WHO 参考线，并在前端提示"WHO 参考数据覆盖范围为 0-60 个月"。

```
GET    /api/children/:id/measurements        # 获取测量历史（支持 ?type=weight|height|head_circumference）
POST   /api/children/:id/measurements        # 添加记录
PUT    /api/children/:id/measurements/:mid   # 编辑记录
DELETE /api/children/:id/measurements/:mid   # 删除记录
```

测量值合理范围（后端校验）：体重 0.5–200 kg；身高 20–250 cm；头围 20–80 cm。

### WHO 参考数据
```
GET /api/who-standards?gender=female&type=weight  # 获取 WHO 百分位数据
```
WHO 数据（0-60月龄，P3/P50/P97）内嵌在后端代码中，不存数据库。`type` 参数与测量记录使用相同枚举值。

---

## 5. Web 端设计（React）

### 页面结构
```
/login              # 登录 / 注册
/dashboard          # 首页：孩子卡片列表
/children/:id       # 孩子详情页
/family             # 家庭管理
```

### 孩子详情页（核心）
- 顶部 Tab：体重 / 身高 / 头围
- 图表区：
  - 折线图（Recharts），X轴日期，Y轴数值
  - 孩子数据用醒目颜色折线
  - WHO P3、P50、P97 灰色虚线叠加
  - Hover 显示具体数值
- 记录列表：表格，支持编辑 / 删除
- 浮动按钮：添加记录（弹窗：日期 + 数值 + 可选备注）

---

## 6. 微信小程序设计

### 页面结构
```
/pages/login        # 微信授权登录
/pages/index        # 首页：孩子列表
/pages/add          # 快速录入（核心）
/pages/chart        # 简版生长曲线图
/pages/family       # 家庭管理 / 邀请码加入
```

### 快速录入页（核心）
1. 选择孩子（单个孩子时默认选中）
2. 选择类型：体重 / 身高 / 头围（大按钮）
3. 输入数值（数字键盘）
4. 日期默认今天（可修改）
5. 提交

### 老人加入流程

流程顺序：先微信授权登录，再输入邀请码加入家庭。

```
老人打开小程序
         ↓
微信授权（获取 openid，创建无家庭账号）
         ↓
进入"加入家庭"页 → 输入邀请码（如：AB1234）
         ↓
加入家庭组 → 可以录入数据
```

老人只需要会用微信，不需要记任何账号密码。

### 图表
使用 wx-charts，展示趋势折线 + P50 参考线（小屏适配，简化显示）。

---

## 7. 部署

### 开发阶段
- 小程序使用**测试号**（免注册，扫码即得），开发者工具内跳过域名校验
- 后端本地运行，小程序通过开发者工具代理或局域网 IP 访问

### 上线阶段
- **后端**：打包 Docker 镜像，部署到**微信云托管（CloudRun）**
  - 小程序调用走微信内网专线，无需域名、无需 ICP 备案
  - 按量计费，家庭级流量费用极低
- **数据库**：PostgreSQL，作为云托管的附属服务或同区腾讯云数据库
- **Web 端**：React 构建产物托管在任意静态服务器（或本机），通过云托管公网 IP + 端口访问（仅家庭内部使用，不对外公开，无需备案）
- **数据库迁移**：使用 `golang-migrate`，迁移文件纳入版本控制
- **配置管理**：敏感配置（数据库 DSN、JWT Secret、微信 AppID/Secret）通过环境变量注入；本地开发用 `.env` 文件，云托管通过控制台环境变量配置，禁止硬编码

### 域名备案
暂不备案。Web 端仅供自己家庭使用，IP 访问即可满足需求。未来如需对外开放 Web 端，再补办备案。

---

## 8. MVP 范围（不在此次实现）

以下功能明确排除在 MVP 之外：
- 疫苗接种管理
- 成长里程碑
- 成长相册
- 喂奶 / 睡眠 / 换尿布记录
- 数据导出（PDF/Excel）
- 智能提醒推送
