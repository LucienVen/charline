# CharLine 开发实施步骤总结

## 1. 总体规划
- 确定项目架构：客户端（CLI）+ 服务端（消息转发）+ Kafka
- 服务端无存储，仅处理认证 / 邀请码 / 消息转发
- 客户端本地持久化：SQLite + Token 本地文件

## 2. 服务端开发步骤
1. 初始化 Golang 项目，选择框架（gin/fiber/net/http）
2. 定义 REST API：
   - `POST /register?invite=xxx` 注册并返回 JWT
   - `POST /login` Token 检查接口（用于客户端校验）
   - `POST /message` 收到消息并转发 Kafka
3. 编写 JWT 认证中间件
4. 实现邀请码机制（内存/Redis/SQLite）
5. Kafka producer 实现消息投递
6. 群组（默认群）管理机制
7. 测试端到端流程

## 3. 客户端开发步骤
1. Golang CLI 项目初始化
2. 启动逻辑：
   - 检查本地 Token 文件
   - 如果不存在，进入邀请码注册流程
3. 注册命令：`charline url username invite`
4. 登录命令：自动 token 校验
5. 聊天流程：
   - WebSocket/长轮询接收消息
   - 输入监听 + 消息发送
6. Token本地保存：`~/.charline/token`
7. SQLite 本地记录聊天内容

## 4. SQLite 集成
- 指定本地路径 `~/.charline/charline.db`
- 建表：
  - `messages(id, sender, content, group, ts)`
- 发送/接收消息均写入数据库
- 使用 go-sqlite3 或 modernc.org/sqlite 库
- 编译带 SQLite 静态库或使用纯 Go 版本

## 5. 编译和打包
- Mac/Linux/Windows 跨平台编译
- CLI 发布为单二进制
- SQLite 内置纯 Go 驱动可减少依赖问题

## 6. Kubernetes / 部署（可选后期）
- Dockerfile 编写
- 部署云服务器
- Kafka 使用独立服务/云托管版

## 7. Token 与安全策略
- 使用 JWT（无过期）
- Payload 携带：
  ```json
  { "uid": "xxx", "grp": "default", "v": 1 }
  ```
- 本地保存 Token，不丢失可一直登录
- 版本号升级时，旧 Token 仍可登录，但服务端可选择版本拒绝

## 8. 日志与监控
- 服务端输出访问和错误日志
- 客户端可选择开启 debug 模式
- Kafka 失败重试机制

## 9. 后续扩展（规划）
- 多群组系统
- 群聊/私聊路由优化
- 邀请创建设备限制
- TLS 加密通信
- 客户端 UI（TUI 模式）

