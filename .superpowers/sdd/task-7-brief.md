### Task 7: 迁移 game/ 核心 — actor_base.go, actor_manager.go, grpc.go

**Files:**
- Move: `server/actor_base.go` → `game/actor_base.go`
- Move: `server/actor_manager.go` → `game/actor_manager.go`
- Move: `server/grpc.go` → `game/grpc.go`

**Interfaces:**
- Produces: `game` 包中的 `IActor`、`ActorBase`、`ActorConfig`、`AManager`、`ActorManner`、`CallRequest`、`Player`（type alias 占位，Task 8 填补）

- [ ] **Step 1: 移动文件**

```bash
git mv server/actor_base.go game/actor_base.go
git mv server/actor_manager.go game/actor_manager.go
git mv server/grpc.go game/grpc.go
```

- [ ] **Step 2: 修改 actor_base.go 的 import**

`"simple_game/pkg"` → `"simple_game/game/pkg"`：

```go
import (
    "context"
    "errors"
    "simple_game/game/pkg"
    "sync/atomic"
    "time"
)
```

- [ ] **Step 3: 修改 actor_manager.go — 无需改 import（只依赖 sync/time 标准库）**

检查无误。

- [ ] **Step 4: 修改 grpc.go 的 import**

`"simple_game/pkg"` → `"simple_game/game/pkg"`，`pt2 "simple_game/api/protos/pt"` → `pt2 "simple_game/game/api/protos/pt"`：

```go
package game

import (
    "context"
    "fmt"
    "net"
    pt2 "simple_game/game/api/protos/pt"
    "simple_game/game/pkg"

    "google.golang.org/grpc"
)

type server struct {
    pt2.UnimplementedHelloServer
}

func (s *server) Say(ctx context.Context, req *pt2.SayRequest) (*pt2.SayResponse, error) {
    fmt.Println("request:", req.Msg)
    return &pt2.SayResponse{Msg: "Hello " + req.Msg}, nil
}

func StartGRPC() {
    listen, err := net.Listen("tcp", ":9901")
    if err != nil {
        fmt.Printf("failed to listen: %v", err)
        return
    }

    s := grpc.NewServer()
    pt2.RegisterHelloServer(s, &server{})
    defer func() {
        s.Stop()
        _ = listen.Close()
    }()

    pkg.INFO("grpc server start port: ", 9901)
    err = s.Serve(listen)
    if err != nil {
        pkg.ERROR("failed to serve", err)
        return
    }
}
```

- [ ] **Step 5: 验证编译（此时 game 包尚缺 Player 定义和 server.go）**

暂不编译 — 等待 Task 8 完成后一起验证。

- [ ] **Step 6: 提交**

```bash
git add game/actor_base.go game/actor_manager.go game/grpc.go server/
git commit -m "refactor: 迁移 actor_base, actor_manager, grpc → game/"
```

---

