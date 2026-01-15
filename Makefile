.PHONY: all build server client run-server run-client clean deps test lint

# 默认目标
all: build

# 拉取依赖
deps:
	@echo "==> 拉取依赖..."
	cd server && go mod download
	cd client && go mod download

# 构建所有
build: server client

# 构建服务端
server:
	@echo "==> 构建服务端..."
	cd server && go build -o ../bin/server ./cmd

# 构建客户端
client:
	@echo "==> 构建客户端..."
	cd client && go build -o ../bin/client ./cmd

# 运行服务端
run-server: server
	@echo "==> 启动服务端..."
	./bin/server

# 运行客户端
run-client: client
	@echo "==> 启动客户端..."
	./bin/client

# 清理
clean:
	@echo "==> 清理构建产物..."
	rm -rf bin/

# 测试
test:
	@echo "==> 运行测试..."
	cd server && go test -v ./...
	cd client && go test -v ./...

# 代码检查
lint:
	@echo "==> 代码检查..."
	golangci-lint run ./...
