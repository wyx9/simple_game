# Simple Game 项目总览

> 本文档由代码梳理自动生成，供后续 ZCode 会话快速理解项目使用。

## 一、项目定位

基于 Go 语言的高性能游戏服务器框架，核心特征：

- **Actor 模型**：协程隔离 + 消息通信的并发架构
- **多协议接入**：TCP（Protobuf 业务流）、HTTP（管理/调试）、gRPC（远程调用示例）
- **冷热数据分离**：MySQL 存冷数据，Redis 存热数据（玩家在线状态）
- **雪花算法**：自研分布式 ID 生成（`pkg/uuid.go`）

模块名：`simple_game`，Go 版本 `1.23.3`。

## 二、入口与启动流程

入口：`main.go`，启动顺序固定如下：

```
main()
 ├─ pkg.StartLog()              // 启动日志消费协程（chan 消费）
 ├─ config.Start()              // 多路径查找并解析 config/config.yaml
 ├─ routes.Init()               // 注册 Protobuf 工厂（LoginReq/HeartReq）
 ├─ register.RegisteredRoute()  // 注册业务路由 → controller
 ├─ pkg.MysqlStart()            // 初始化 MySQL 连接
 ├─ pkg.RedisStart()            // 初始化 Redis 连接
 ├─ go server.StartGRPC()       // gRPC 服务 :9901（Hello.Say）
 ├─ go http.StartHttp()         // HTTP 服务 :9902（/、/hello）
 ├─ go server.Start()           // TCP 游戏服务 :8888（核心）
 └─ <-sigChan → Stop()          // 收到 SIGINT/SIGTERM 优雅退出
```

`Stop()` 当前实现为空壳（保存逻辑被注释 TODO）。

## 三、目录结构与职责

| 目录 | 职责 | 关键文件 |
|------|------|----------|
| `main.go` | 程序入口、启动编排 | — |
| `config/` | YAML 配置加载 | `gosconf.go` `config.yaml` |
| `pkg/` | 基础设施包 | logger / mysql / redis / msg(TCP分包) / uuid(雪花) |
| `routes/` | 路由分发与 Protobuf 工厂 | `routes.go` `protobuf_factory.go` |
| `register/` | 业务路由注册（协议名→handler） | `register_route.go` |
| `controller/` | 业务处理函数 | `all_controller.go` `base_controller.go` |
| `server/` | 核心服务：TCP/Actor/玩家 | server.go / actor_base.go / actor_manager.go / actor_player.go / grpc.go |
| `http/` | HTTP 服务 | `http.go` |
| `libs/` | 公共库：消息包编解码、定时器 | `packet_msg.go` `timer_task.go` |
| `api/protos/` | Protobuf 协议定义与生成代码 | `hello.proto` `user.proto` |
| `client/` | 测试客户端（TCP/gRPC） | `client1.go` `rpc_client.go` |
| `configs/` | Excel 配置表导入示例 | `excel01.go` `excel02.go` |
| `test/` | 算法/数据结构练习测试（非业务） | 各 `*_test.go` |
| `Makefile` | `make pt` 生成协议代码 | — |

## 四、核心架构：Actor 模型

### 4.1 接口与基类（`server/actor_base.go`）

```go
type IActor interface {
    Start()
    Stop()
    Handler(msg interface{}) interface{}
}
```

- `ActorBase`：通用基座，内含 `request chan interface{}`（默认 1024 缓冲）、`stop chan`、`ctx/cancel`、原子关闭标志 `closed`。
- `loop()`：先调 `Start()`，再起协程消费 `request`，每条消息用 `HandlerTimeout`（默认 5s）超时保护。
- `NewActorBase(name, callBack, config)`：创建即注册到 `ActorManner` 并启动循环。

### 4.2 管理器（`server/actor_manager.go`）

- `AManager`：`actorMap`（name→Actor）+ `playerMap`（ip→Actor），均用 `sync.Map`。
- 通信方式：
  - `CastMsg(name, msg)` — 异步投递，带 `SendTimeout`（默认 5s）。
  - `CallMsg(name, msg)` — 同步调用，封装 `CallRequest{Msg, Response}` 等待回包。

### 4.3 玩家 Actor（`server/actor_player.go`）

- `Player` 实现 `IActor`：持有 `conn`、`PlayerData`、`timer`。
- `Start()`：从 Redis 加载数据 + 启动 10s 定时 tick。
- `Stop()`：停定时器 + 持久化到 Redis。
- `Handler(msg)`：`"tick"` 走定时持久化；其余走 `HandlerByClient` → `routes.Route` 分发。
- `StartNewPlayerActor(name, conn)`：登录时由 TCP 服务调用创建。

## 五、TCP 服务与消息流（`server/server.go`）

### 连接管理
- `Server` 持 `connMap map[string]*Connection`（key=RemoteAddr）+ `sync.RWMutex`。
- 常量：`maxConnections=1000`、`heartbeatTimeout=30s`、`readTimeout=10s`、`shutdownTimeout=10s`。
- `handleConnection`：每连接一协程，循环 `pkg.RecvData` 读包。

### 消息处理流程

```
客户端 TCP 包
  ↓ pkg.RecvData (4字节大端长度 + payload)
  ↓ libs.DeCodePack (JSON 解析 PacketMsg{Name, Data})
  ↓ 首包强制 LoginReq：
      - proto.Unmarshal → pt.LoginReq
      - 标记 isAuth=true，actorId=UUid
      - StartNewPlayerActor(actorId, conn)
  ↓ 后续包：ActorManner.CastMsg(actorId, buf) → Player.Handler
      - routes.Route(player, Name, Data)
      - ProtoBufFactory[name]() 创建 message 并 Unmarshal
      - 调用注册的 handler 返回 proto.Message
      - protojson.Marshal 后 conn.Write 回包
```

### 协议包格式（`libs/packet_msg.go`）

```go
type PacketMsg struct {
    Name string `json:"name"`  // 协议名，如 "LoginReq"
    Data []byte `json:"data"`  // proto.Marshal 后的字节
}
```
- `EnCodePack/DeCodePack`：JSON 序列化整个 PacketMsg。
- `Pack2Msg(data)`：根据 proto.Message 反射取类型名作为 Name。
- TCP 链路层：4 字节大端长度头 + JSON(PacketMsg)（注意：业务 Data 内部再用 protobuf 编码）。

## 六、路由与控制器

### 6.1 路由注册（`routes/routes.go` + `register/register_route.go`）

```go
type Handler func(ctx interface{}, req any) proto.Message
var routes = map[string]Handler{}
```

`register_route.go` 中注册：
- `"LoginReq"` → `controller.LoginController` 返回 `pt.LoginRes{UUid, Code:0}`
- `"HeartReq"` → `controller.TestController` 返回 `pt.HeartRes{Time: now}`

### 6.2 Protobuf 工厂（`routes/protobuf_factory.go`）

`ProtoBufFactory[name] = func() proto.Message`，通过反射 `reflect.TypeOf(p).Elem().Name()` 取结构名作 key。`Init()` 注册 `LoginReq`、`HeartReq`。

> ⚠️ 注意 `routes.Route` 中调用 `f()` 后未将 unmarshal 结果传给 handler，而是传 `f()`（每次新建实例），存在已知缺陷——Unmarshal 的对象未被复用，handler 拿到的是空对象。

### 6.3 控制器（`controller/`）

- `BaseController.Ctx(ctx)`：把 `interface{}` 强转 `*server.Player`。
- `AllController` 内嵌 `BaseController`，聚合业务方法。

## 七、配置（`config/`）

`config.yaml` 结构：

```yaml
OpenGm: bool
MySql: {Addr, Port, User, PassWord, DBName}
Redis: {Addr, PassWord, DB}
Tcp:   {Addr, Port}   # 8888
Http:  {Addr, Port}   # 9902
Rpc:   {Addr, Port}   # 9901 (注意 grpc.go 当前硬编码 :9901)
```

`Start()` 按多候选路径查找文件（相对路径 / 绝对路径 / game 子目录 / 上级目录）。

## 八、基础设施（`pkg/`）

| 文件 | 功能 |
|------|------|
| `logger.go` | 4 级日志（DEBUG/INFO/WARNING/ERROR），写文件 + chan 异步打印控制台（带 ANSI 颜色） |
| `mysql.go` | `sqlx.Open` 全局 `DB` |
| `redis.go` | `redis.NewClient` 全局 `RedisClient` + `RCtx` |
| `msg.go` | `SendData/RecvData`：4 字节大端长度前缀的 TCP 分包 |
| `uuid.go` | 雪花算法 `GetSnowflakeId` / `NewUid`；机器 ID 由本地 IP 推导 |

## 九、辅助模块

### 9.1 定时器（`libs/timer_task.go`）
- `TickerEventQueue`（`container.list`）存储事件。
- `AddTimer(triggerAt, cb)` 绝对时间触发；`AddScheduleTimer(delay, cb)` 相对时间。
- `TimerStart()` 在 `init()` 起协程轮询，到点回调；`ScheduleAt>0` 自动循环。
- ⚠️ `CheckEventTrigger` 用了 `Front().Prev()`，存在 nil 风险（似乎未实际使用）。

### 9.2 客户端（`client/`）
- `client1.go`：TCP 客户端，3s 心跳 + 30s 超时断连 + 自动重连，控制台输入触发 `LoginReq`。
- `rpc_client.go`：gRPC 客户端调 `Hello.Say`。

### 9.3 协议（`api/protos/`）
- `user.proto`：`Student`、`LoginReq/LoginRes`、`HeartReq/HeartRes`
- `hello.proto`：gRPC `Hello` 服务 + `SayRequest/SayResponse`
- 生成代码在 `api/protos/pt/`，`make pt` 重新生成。

## 十、已知问题与 TODO

1. **`routes.Route` Unmarshal 结果丢失**：`_ = proto.Unmarshal(msg.([]byte), f())` 创建的实例未传给 handler，handler 拿到的是 `f()` 新建空对象。
2. **`grpc.go` 端口硬编码**：`net.Listen("tcp", ":9901")` 未读 `config.Conf.Rpc`。
3. **`server.handleConnection` 重复 err 判断**：第 159/163 行 `if err != nil` 逻辑冗余。
4. **`Stop()` 未实现**：退出时未触发玩家数据持久化（注释掉的 TODO）。
5. **`libs/timer_task.go` 的 `CheckEventTrigger`**：`Front().Prev()` 必为 nil，调用会 panic（当前未被调用）。
6. **`game_test.go` 已过期**：引用了不存在的 `utils`/`actor`/`packet_msg` 包，与当前架构不符（属于重构前残留）。
7. **`test/` 目录**：纯算法练习（dp/astar/skip/quicksort 等），与游戏业务无关。

## 十一、关键调用关系图

```
main
 ├─ config.Start ──→ config.Conf (全局)
 ├─ routes.Init ──→ ProtoBufFactory
 ├─ register.RegisteredRoute ──→ routes.routes
 ├─ pkg.MysqlStart/RedisStart ──→ pkg.DB / pkg.RedisClient
 ├─ server.StartGRPC (gRPC :9901)
 ├─ http.StartHttp (HTTP :9902)
 └─ server.Start (TCP :8888)
      └─ Server.serverLoop
           └─ handleConnection
                ├─ pkg.RecvData  (4字节长度+payload)
                ├─ libs.DeCodePack (JSON→PacketMsg)
                ├─ 首包 LoginReq → StartNewPlayerActor
                │    └─ NewActorBase → ActorManner.Add → loop()
                └─ ActorManner.CastMsg(actorId, buf)
                     └─ Player.Handler → HandlerByClient
                          └─ routes.Route(player, Name, Data)
                               ├─ ProtoBufFactory[name]().Unmarshal
                               └─ routes.routes[name](player, req)
                                    └─ controller.LoginController/TestController
                               └─ protojson.Marshal → conn.Write
```

## 十二、快速运行

```bash
# 1. 安装依赖
go mod tidy

# 2. 修改 ./config/config.yaml（MySQL/Redis/端口）

# 3. 启动服务器
go run main.go

# 4. 启动 TCP 客户端测试
cd client && go run client1.go

# 5. 重新生成协议
make pt

# 6. 性能分析
go tool pprof --http=:1234 http://127.0.0.1:4890/debug/pprof/heap
```

---

*文档生成日期：2026-07-05*
