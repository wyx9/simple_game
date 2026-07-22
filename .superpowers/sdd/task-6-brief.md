### Task 6: 迁移 controller/ → game/controller/ + register/ → game/register/

**Files:**
- Move: `controller/base_controller.go` → `game/controller/base_controller.go`
- Move: `controller/all_controller.go` → `game/controller/all_controller.go`
- Move: `register/register_route.go` → `game/register/register_route.go`

**Interfaces:**
- Consumes: `server` 包 (Player) → 需改 import 为 `game`
- Produces: `game/controller` 包、`game/register` 包

- [ ] **Step 1: 移动 controller 文件**

```bash
mkdir -p game/controller game/register
git mv controller/base_controller.go game/controller/base_controller.go
git mv controller/all_controller.go game/controller/all_controller.go
git mv register/register_route.go game/register/register_route.go
```

- [ ] **Step 2: 修改 controller/base_controller.go 的 import**

`"simple_game/server"` → `"simple_game/game"`，类型引用 `*server.Player` → `*game.Player`：

```go
package controller

import (
    "simple_game/game"
)

type BaseController struct {
}

func (this BaseController) Ctx(ctx interface{}) *game.Player {
    return ctx.(*game.Player)
}
```

- [ ] **Step 3: 修改 controller/all_controller.go 的 import**

`"simple_game/server"` → `"simple_game/game"`：

```go
package controller

import (
    "google.golang.org/protobuf/proto"
    "simple_game/game/api/protos/pt"
    "simple_game/game"
    "time"
)

type AllController struct {
    BaseController
}

func LoginController(actor *game.Player, req *pt.LoginReq) *pt.LoginRes {
    return &pt.LoginRes{
        UUid: req.UUid,
        Code: 0,
    }
}

func TestController(actor *game.Player, req *pt.HeartReq) proto.Message {
    return &pt.HeartRes{Time: time.Now().Unix()}
}
```

- [ ] **Step 4: 修改 register/register_route.go 的 import**

```go
package register

import (
    "google.golang.org/protobuf/proto"
    "simple_game/game/api/protos/pt"
    "simple_game/game/controller"
    "simple_game/game/routes"
)

func RegisteredRoute() {
    allController := controller.AllController{}

    routes.AddRoute("LoginReq", func(ctx interface{}, req any) proto.Message {
        player := allController.Ctx(ctx)
        return controller.LoginController(player, req.(*pt.LoginReq))
    })

    routes.AddRoute("HeartReq", func(ctx interface{}, req any) proto.Message {
        player := allController.Ctx(ctx)
        return controller.TestController(player, req.(*pt.HeartReq))
    })
}
```

- [ ] **Step 5: 编译验证**

```bash
go build ./game/controller/... ./game/register/...
```

注意：此时 `game` 主包尚未创建，controller 引用 `*game.Player` 会编译失败。**这是预期的**——等待 Task 8 迁移 Actor 文件后一起验证。

- [ ] **Step 6: 提交**

```bash
git add game/controller/ game/register/ controller/ register/
git commit -m "refactor: 迁移 controller/ register/ → game/"
```

---

## Phase 2: Game 核心迁移

