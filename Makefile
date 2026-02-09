.PHONY: build run clean test

# 变量
APP_NAME := server
BUILD_DIR := bin
MAIN_FILE := cmd/server/main.go

# 构建
build:
	@echo "🔨 构建中..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "✅ 构建完成: $(BUILD_DIR)/$(APP_NAME)"

# 运行
run:
	@echo "🚀 启动服务..."
	go run $(MAIN_FILE)

# 清理
clean:
	@echo "🧹 清理中..."
	@rm -rf $(BUILD_DIR)
	@echo "✅ 清理完成"

# 测试
test:
	@echo "🧪 运行测试..."
	go test -v ./...

# 安装依赖
deps:
	@echo "📦 安装依赖..."
	go mod tidy
	@echo "✅ 依赖安装完成"

# 格式化代码
fmt:
	@echo "🎨 格式化代码..."
	go fmt ./...

# 代码检查
lint:
	@echo "🔍 代码检查..."
	go vet ./...
