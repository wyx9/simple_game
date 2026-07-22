### Task 11: 完善 game/main.go

**Files:**
- Modify: `game/main.go`

**Interfaces:**
- Produces: 可执行的 Game 服务入口

- [ ] **Step 1: 创建 game/config.go（配套配置加载）**

```go
// game/config.go
package game

import (
    "os"
    "gopkg.in/yaml.v3"
)

type GameConfig struct {
    Listen struct {
        Addr string `yaml:"Addr"`
        Port string `yaml:"Port"`
    } `yaml:"Listen"`
    Grpc struct {
        Addr string `yaml:"Addr"`
        Port string `yaml:"Port"`
    } `yaml:"Grpc"`
    SaveLog     bool   `yaml:"SaveLog"`
    TokenSecret string `yaml:"TokenSecret"`
    MySql       struct {
        Addr     string `yaml:"Addr"`
        Port     int    `yaml:"Port"`
        User     string `yaml:"User"`
        PassWord string `yaml:"PassWord"`
        DBName   string `yaml:"DBName"`
    } `yaml:"MySql"`
    Redis struct {
        Addr     string `yaml:"Addr"`
        PassWord string `yaml:"PassWord"`
        DB       int    `yaml:"DB"`
    } `yaml:"Redis"`
}

func LoadGameConfig(path string) (*GameConfig, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    cfg := &GameConfig{}
    if err := yaml.Unmarshal(data, cfg); err != nil {
        return nil, err
    }
    return cfg, nil
}
```

- [ ] **Step 2: 重写 game/main.go**

```go
// game/main.go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "simple_game/game"
    "simple_game/game/pkg"
    "simple_game/game/register"
    "simple_game/game/routes"
)

func main() {
    // 加载配置
    cfg, err := game.LoadGameConfig("game/config/config.yaml")
    if err != nil {
        fmt.Println("load config failed:", err)
        os.Exit(1)
    }

    // 日志
    pkg.StartLog()
    if cfg.SaveLog {
        _ = pkg.InitLogFile()
    }

    // 路由注册
    routes.Init()
    register.RegisteredRoute()

    // 基础设施
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
        cfg.MySql.User, cfg.MySql.PassWord, cfg.MySql.Addr, cfg.MySql.Port, cfg.MySql.DBName)
    pkg.MysqlStart(dsn)
    pkg.RedisStart(cfg.Redis.Addr, cfg.Redis.PassWord, cfg.Redis.DB)

    // 信号
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // 启动服务
    go game.StartGRPC()
    go game.Start(
        fmt.Sprintf("%s:%s", cfg.Listen.Addr, cfg.Listen.Port),
        cfg.TokenSecret,
    )

    fmt.Println("Game server started")
    <-sigChan
    fmt.Println("Game server shutting down...")
}
```

- [ ] **Step 3: 编译验证**

```bash
go build ./game/...
```

预期 `game/main.go` 编译为可执行文件。验证 `go build -o /dev/null ./game` 成功。

- [ ] **Step 4: 提交**

```bash
git add game/config.go game/main.go
git commit -m "feat: 完善 game/main.go 入口"
```

---

## Phase 4: Agent 服务

