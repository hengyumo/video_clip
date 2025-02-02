# 定义项目名称
BINARY_NAME=clipper

# 定义版本号
VERSION=1.0.0

# 定义支持的操作系统和架构
PLATFORMS=darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64

# 定义默认目标
all: fmt clean build

# 格式化代码
fmt:
	go fmt ./...

# 清理构建目录
clean:
	rm -rf build/

# 构建所有平台的可执行文件
build:
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		output_dir=build/$${os}_$${arch}; \
		output_name=$${output_dir}/$(BINARY_NAME); \
		if [ "$${os}" = "windows" ]; then \
			output_name=$${output_name}.exe; \
		fi; \
		echo "Building for $${os}/$${arch}..."; \
		GOOS=$${os} GOARCH=$${arch} go build -o $${output_name} ./cmd/clipper/main.go; \
	done

# 打包所有平台的可执行文件
package: build
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		output_dir=build/$${os}_$${arch}; \
		output_name=$${output_dir}/$(BINARY_NAME); \
		if [ "$${os}" = "windows" ]; then \
			output_name=$${output_name}.exe; \
		fi; \
		zip -j build/$(BINARY_NAME)_$(VERSION)_$${os}_$${arch}.zip $${output_name}; \
	done

# 运行默认目标
.DEFAULT_GOAL := all