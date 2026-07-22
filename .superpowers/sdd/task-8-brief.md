### Task 8: 迁移 actor_player.go → game/actor_player.go

**Files:**
- Move: `server/actor_player.go` → `game/actor_player.go`

**Interfaces:**
- Produces: `game.Player` 结构体、`game.PlayerData`、`game.StartNewPlayerActor`
- 修正 import: `"simple_game/pkg"` → `"simple_game/game/pkg"`，`"simple_game/libs"` → `"simple_game/game/libs"`，`"simple_game/routes"` → `"simple_game/game/routes"`

- [ ] **Step 1: 移动文件**

```bash
git mv server/actor_player.go game/actor_player.go
```

- [ ] **Step 2: 修改 import**

```go
package game

import (
    "encoding/json"
    "fmt"
    "simple_game/game/libs"
    "simple_game/game/pkg"
    "simple_game/game/routes"
    "time"
)
```

删除 `"google.golang.org/protobuf/encoding/protojson"` 和 `"google.golang.org/protobuf/proto"` import（上一轮修复中已移除这些引用）。

- [ ] **Step 3: 删除 StartNewPlayerActor 中的 Login 逻辑（按 spec 7.1）**

原函数直接创建 Actor，不再包含登录认证（token 验证由 Game server 在 handleAgentConn 中提前完成）：

```go
func StartNewPlayerActor(name string, conn pkg.NetConn) (*Player, error) {
    playerActor := &Player{
        name: name,
        ip:   conn.RemoteAddr(),
        conn: conn,
        PlayerData: &PlayerData{
            Name:        name,
            LastLoginAt: time.Now().Unix(),
        },
    }
    _ = NewActorBase(name, playerActor, nil)
    return playerActor, nil
}
```

（无实质变更，只是确认函数与当前一致）

- [ ] **Step 4: 验证 game 包整体编译**

```bash
go build ./game/...
```

预期：game 包编译通过（server.go 尚需重写，故 game main 暂不编译）。

- [ ] **Step 5: 提交**

```bash
git add game/actor_player.go server/
git commit -m "refactor: 迁移 actor_player → game/"
```

---

