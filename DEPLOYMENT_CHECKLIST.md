# 🚀 用户管理系统部署检查清单

## ✅ 部署前检查

### 1. 依赖包检查
```bash
# Go依赖检查
go mod tidy
go mod download

# 前端依赖检查
cd web && npm install
```

### 2. 数据库迁移
```bash
# 执行用户管理系统迁移
go run cmd/migrate/main.go
```
✅ 应显示：用户管理系统数据库迁移完成!

### 3. 编译测试
```bash
# 后端编译
go build -o gpt-load main.go

# 前端构建
cd web && npm run build
```

## ✅ 部署后验证

### 1. 默认管理员登录测试
- 访问: `http://your-domain/user-login`
- 用户名: `admin`
- 密码: `password`
- ✅ 应能成功登录并看到用户菜单

### 2. 功能验证
- [ ] 个人资料页面正常显示
- [ ] 用户管理页面（管理员）可访问
- [ ] 创建新用户功能正常
- [ ] 权限控制生效（非管理员看不到用户管理）
- [ ] 退出登录功能正常

### 3. 原有功能兼容性
- [ ] Dashboard页面正常
- [ ] Keys管理正常
- [ ] Logs查看正常
- [ ] Settings页面正常（管理员）
- [ ] Claude Tokens功能正常

### 4. 安全测试
- [ ] 错误密码5次后账户锁定
- [ ] 会话过期自动跳转登录页
- [ ] 无权限用户不能访问管理页面

## ⚠️ 安全提醒

### 必须执行的安全操作：
1. **立即修改默认管理员密码**
2. **检查数据库连接安全性**
3. **配置HTTPS**（生产环境）
4. **设置防火墙规则**

### 可选的安全加强：
- 启用双因素认证
- 设置更严格的密码策略
- 配置访问日志监控
- 定期备份用户数据

## 🔄 回滚方案

如果遇到问题需要回滚：

1. **数据库回滚**:
   ```sql
   -- 删除新增的用户相关表（谨慎操作！）
   DROP TABLE IF EXISTS user_activities;
   DROP TABLE IF EXISTS user_sessions;
   DROP TABLE IF EXISTS user_groups;
   DROP TABLE IF EXISTS users;
   ```

2. **代码回滚**:
   ```bash
   git checkout HEAD~1  # 回滚到上一个版本
   ```

3. **重新构建**:
   ```bash
   go build && cd web && npm run build
   ```

## 📞 问题排查

### 常见错误及解决方案：

1. **"数据库连接失败"**
   - 检查 `DATABASE_DSN` 配置
   - 确认数据库服务运行状态

2. **"迁移失败"**
   - 检查数据库写入权限
   - 查看具体错误信息

3. **"登录页面404"**
   - 确认前端构建成功
   - 检查路由配置

4. **"权限验证失败"**
   - 清除浏览器缓存和localStorage
   - 检查JWT令牌是否有效

---

✅ **部署完成标准**: 所有检查项通过，系统功能正常，安全措施到位