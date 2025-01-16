# Simple Game - 基于 Actor 模型的 Go 游戏服务器框架

一个基于 Go 语言开发的高性能游戏服务器框架，采用 Actor 模型实现并发处理。

## 主要特性

- **网络支持**
  - TCP 连接（使用 Protocol Buffers 数据格式）
  - HTTP 服务支持
  - gRPC 远程调用支持

- **Actor 模型实现**
  - 协程间通信机制
  - 协程数据隔离
  - 基于消息的通信架构

- **数据持久化**
  - MySQL 处理冷数据存储
  - Redis 处理热数据缓存
  - 高效的数据管理机制

- **核心组件**
  - 定时器系统
  - 日志模块
  - 路由模块
  - 测试客户端
  - 通过 Makefile 生成协议文件

## 项目结构

```
simple_game/
├── api/        - API 定义
├── client/     - 客户端实现
├── config/     - 配置文件
├── configs/    - 附加配置
├── controller/ - 游戏控制器
├── http/       - HTTP 处理器
├── libs/       - 公共库
├── pkg/        - 核心包
├── register/   - 服务注册
├── routes/     - 路由定义
├── server/     - 服务器实现
├── test/       - 测试文件
└── utils/      - 工具函数
```

## 环境要求

- Go 1.23.3 或更高版本
- Protocol Buffers 编译器
- MySQL 数据库
- Redis 缓存服务

## 快速开始

### 环境配置

```shell
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
```

### 项目初始化

1. 安装依赖：
```shell
go mod tidy
```

2. 配置应用：
   - 修改 `./config/config.yaml` 配置文件

3. 运行服务器：
```shell
go run main.go
```

### 协议生成

#### 生成 Protocol Buffers 文件

```shell
cd protos
protoc -I=. --go_out=. ./*.proto
# 或使用 make：
make pt
```

#### 生成 gRPC 代码

1. 安装所需的 protoc 插件：
```shell
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
```

2. 生成 gRPC 代码：
```shell
cd protos
protoc --go_out=. --go-grpc_out=. *.proto
# 或使用 make：
make pt
```

### 构建客户端

```shell
cd client
go build -o client1 client1.go
go build -o client2 client2.go
```

## 性能分析

启动性能分析工具：
```shell
go tool pprof --http=:1234 http://127.0.0.1:4890/debug/pprof/heap
```

## 主要依赖

项目使用的主要依赖包括：
- github.com/go-redis/redis/v8 - Redis 客户端
- github.com/go-sql-driver/mysql - MySQL 驱动
- github.com/jmoiron/sqlx - SQL 扩展库
- google.golang.org/grpc - gRPC 框架
- google.golang.org/protobuf - Protocol Buffers 支持
- gopkg.in/yaml.v3 - YAML 配置支持

## 技术架构

### 网络层
- TCP 服务：支持使用 Protocol Buffers 的 TCP 连接
- HTTP 服务：提供 RESTful API 支持
- gRPC 服务：支持高效的远程过程调用

### Actor 模型特性
- 协程隔离：确保每个 Actor 的数据安全
- 消息通信：Actor 间通过消息传递进行通信
- 状态管理：每个 Actor 管理自己的状态

### 数据层
- 冷数据：使用 MySQL 存储需要持久化的数据
- 热数据：使用 Redis 处理高频访问的数据
- 数据同步：支持热数据和冷数据的同步机制

