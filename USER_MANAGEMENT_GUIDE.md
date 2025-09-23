# GPT-Load 用户管理系统集成指南

## 🚀 功能概览

本次更新为 GPT-Load 项目添加了完整的用户管理系统，实现了企业级的用户认证、权限控制和管理功能。

### ✨ 主要特性

- **多角色权限系统**: 管理员、用户、访客三级权限
- **安全认证**: bcrypt密码加密、JWT令牌、会话管理
- **用户管理**: 完整的用户CRUD操作
- **权限控制**: 基于角色的路由和功能访问控制
- **安全机制**: 账户锁定、失败尝试限制
- **活动日志**: 用户操作记录和审计

## 📋 部署步骤

### 1. 数据库迁移

首先执行数据库迁移以创建用户相关表：

```bash
# 方法一：使用迁移工具
go run cmd/migrate/main.go

# 方法二：在项目启动时自动执行
# 系统会自动检测并执行迁移
```

### 2. 依赖包安装

确保前端依赖已安装：

```bash
cd web
npm install
# 或
yarn install
```

### 3. 配置检查

确保以下Go依赖已添加到 `go.mod`：

```go
require (
    github.com/google/uuid v1.3.0
    golang.org/x/crypto v0.14.0
    // 其他现有依赖...
)
```

## 🔧 系统配置

### 默认管理员账户

系统会自动创建默认管理员账户：
- **用户名**: `admin`
- **密码**: `password`
- **角色**: 管理员

⚠️ **安全提醒**: 首次登录后请立即修改默认密码！

### 用户角色说明

| 角色 | 权限说明 |
|------|----------|
| **管理员 (admin)** | 完全权限：用户管理、系统设置、所有功能 |
| **用户 (user)** | 标准权限：查看和使用主要功能，无管理权限 |
| **访客 (viewer)** | 只读权限：仅能查看信息，不能修改 |

## 🎯 功能使用

### 用户登录

1. 访问 `/user-login` 进入新用户登录页面
2. 使用用户名/邮箱和密码登录
3. 系统支持"记住登录状态"功能

### 用户管理（管理员）

1. 登录后在导航栏找到"用户管理"
2. 功能包括：
   - 查看用户列表
   - 创建新用户
   - 编辑用户信息
   - 删除用户
   - 角色和状态管理

### 个人资料管理

1. 点击右上角用户头像菜单
2. 选择"个人资料"
3. 可修改：
   - 显示名称
   - 邮箱地址
   - 头像URL
   - 密码

## 🔐 安全特性

### 密码安全

- 使用 bcrypt 加密存储
- 最少6位密码要求
- 密码修改需验证旧密码

### 账户安全

- 连续5次登录失败自动锁定30分钟
- 会话令牌24小时过期
- IP地址记录和监控

### 权限控制

- 路由级别的权限检查
- API接口权限验证
- 前端菜单动态显示

## 🌐 API接口

### 公开接口

```bash
POST /api/users/login          # 用户登录
```

### 认证接口（需要登录）

```bash
GET    /api/users/profile      # 获取当前用户信息
PUT    /api/users/profile      # 更新当前用户信息
POST   /api/users/logout       # 退出登录
POST   /api/users/change-password # 修改密码
```

### 管理员接口（需要管理员权限）

```bash
POST   /api/users             # 创建用户
GET    /api/users             # 获取用户列表
GET    /api/users/:id         # 获取指定用户
PUT    /api/users/:id         # 更新用户信息
DELETE /api/users/:id         # 删除用户
```

## 🎨 前端组件

### 新增页面

- **用户登录页面**: `/user-login`
- **用户管理页面**: `/users`
- **个人资料页面**: `/profile`

### 新增组件

- **UserMenu**: 用户菜单和头像下拉
- **UserManagement**: 用户管理表格
- **UserProfile**: 个人资料编辑

## 🔄 系统兼容性

### 双系统并存

- 新用户系统与原有认证系统完全兼容
- 优先使用新用户系统，回退到原有系统
- 无缝迁移，不影响现有用户

### 权限映射

```typescript
// 新系统权限检查示例
const hasPermission = (permission: string) => {
  const userRole = userStore.currentUser?.role

  if (userRole === 'admin') return true

  const permissions = {
    'user:read': ['admin', 'user', 'viewer'],
    'user:write': ['admin', 'user'],
    'user:delete': ['admin'],
    // 更多权限...
  }

  return permissions[permission]?.includes(userRole) || false
}
```

## 🛠 开发说明

### 数据库表结构

```sql
-- 用户表
users (
  id, username, email, password_hash, display_name,
  role, status, avatar, last_login_at, last_login_ip,
  login_count, failed_attempts, locked_until,
  created_at, updated_at
)

-- 用户会话表
user_sessions (
  id, user_id, token, user_agent, ip_address,
  expires_at, created_at, updated_at
)

-- 用户活动日志
user_activities (
  id, user_id, action, resource, resource_id,
  ip_address, user_agent, details, timestamp
)

-- 用户分组关联表
user_groups (
  id, user_id, group_id, created_at
)
```

### 扩展开发

1. **添加新权限**:
   ```typescript
   // 在 stores/user.ts 中添加新权限
   const permissions = {
     'new:permission': ['admin', 'user'],
     // ...
   }
   ```

2. **添加新用户字段**:
   ```go
   // 在 models/types.go User 结构体中添加
   type User struct {
     // 现有字段...
     NewField string `gorm:"type:varchar(255)" json:"new_field"`
   }
   ```

3. **自定义验证规则**:
   ```go
   // 在 services/user_service.go 中添加验证逻辑
   func validateCustomRules(user *User) error {
     // 自定义验证逻辑
   }
   ```

## 🐛 故障排除

### 常见问题

1. **迁移失败**:
   ```bash
   # 检查数据库连接
   # 确保数据库权限足够
   # 查看错误日志
   ```

2. **登录失败**:
   ```bash
   # 检查密码是否正确
   # 确认账户未被锁定
   # 查看用户状态是否为 'active'
   ```

3. **权限问题**:
   ```bash
   # 检查用户角色设置
   # 确认路由权限配置
   # 查看前端权限检查逻辑
   ```

### 调试技巧

```typescript
// 前端调试
console.log('User Store:', userStore.currentUser)
console.log('Token:', userStore.token)
console.log('Permissions:', userStore.hasRole('admin'))

// 后端调试
// 查看日志文件或控制台输出
// 检查数据库记录
// 验证JWT令牌
```

## 📞 技术支持

如遇到问题，请检查：

1. **数据库连接**是否正常
2. **依赖包**是否完整安装
3. **迁移**是否成功执行
4. **配置文件**是否正确

系统已完成全面集成，现在您可以使用完整的用户管理功能了！

---

**版本**: v1.0.0
**更新时间**: 2024年
**兼容性**: 完全向后兼容