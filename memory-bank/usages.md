# 项目使用文档

## Makefile 使用指南

### 语法说明
```makefile
目标(TARGET):
	命令(COMMAND)
```

### 可用命令

| 命令 | 说明 | 示例 |
|------|------|------|
| `make` 或 `make all` | 构建服务端和客户端 | `make` |
| `make server` | 仅构建服务端 | `make server` |
| `make client` | 仅构建客户端 | `make client` |
| `make run-server` | 构建并运行服务端 | `make run-server` |
| `make run-client` | 构建并运行客户端 | `make run-client` |
| `make deps` | 拉取 Go 依赖 | `make deps` |
| `make test` | 运行测试 | `make test` |
| `make lint` | 代码检查 | `make lint` |
| `make clean` | 清理构建产物 | `make clean` |

### 命令详解

#### 1. 构建命令
```bash
# 构建所有（默认）
make

# 仅构建服务端，输出到 bin/server
make server

# 仅构建客户端，输出到 bin/client
make client
```

#### 2. 运行命令
```bash
# 构建并启动服务端（监听 :8080）
make run-server

# 构建并启动客户端
make run-client
```

#### 3. 依赖管理
```bash
# 拉取所有依赖
make deps
```

#### 4. 质量检查
```bash
# 运行所有测试
make test

# 代码规范检查（需要安装 golangci-lint）
make lint
```

#### 5. 清理
```bash
# 删除 bin/ 目录
make clean
```

### 输出目录结构
```
bin/
├── server    # 服务端可执行文件
└── client    # 客户端可执行文件
```

### 典型工作流程

**首次使用**:
```bash
make deps    # 拉取依赖
make build   # 构建
```

**日常开发**:
```bash
make run-server   # 启动服务端（新终端）
make run-client   # 启动客户端（新终端）
```

**提交前检查**:
```bash
make test   # 运行测试
make lint   # 代码检查
make clean  # 清理构建产物
```
