.PHONY: build run clean test docs

# 构建应用
build:
	go build -o bin/transfer cmd/main.go

# 运行应用
run:
	go run cmd/main.go

# 清理构建文件
clean:
	rm -rf bin/
	rm -rf logs/

# 运行测试
test:
	go test -v ./...

# 生成API文档
docs:
	swag init -g cmd/main.go -o docs/

# 安装依赖
deps:
	go mod tidy
	go mod download

# 格式化代码
fmt:
	go fmt ./...

# 检查代码
lint:
	golangci-lint run

# 开发模式运行
dev: docs run
