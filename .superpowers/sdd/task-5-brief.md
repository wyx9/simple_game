### Task 5: 迁移 routes/ → game/routes/

**Files:**
- Move: `routes/routes.go` → `game/routes/routes.go`
- Move: `routes/protobuf_factory.go` → `game/routes/protobuf_factory.go`

**Interfaces:**
- Consumes: `pkg` 包 (logger) → 需改 import
- Produces: `game/routes` 包，导出 `Handler`、`AddRoute`、`Route`、`ProtoBufFactory`、`RegisterProtoBufFactory`、`Init`

- [ ] **Step 1: 移动文件**

```bash
mkdir -p game/routes
git mv routes/routes.go game/routes/routes.go
git mv routes/protobuf_factory.go game/routes/protobuf_factory.go
```

- [ ] **Step 2: 修改 import 路径**

`game/routes/routes.go` 中 `"simple_game/pkg"` → `"simple_game/game/pkg"`：

```go
import (
    "errors"
    "google.golang.org/protobuf/proto"
    "simple_game/game/pkg"
)
```

`game/routes/protobuf_factory.go` 中 `"simple_game/api/protos/pt"` → `"simple_game/game/api/protos/pt"`：

```go
import (
    "google.golang.org/protobuf/proto"
    "reflect"
    "simple_game/game/api/protos/pt"
)
```

- [ ] **Step 3: 编译验证**

```bash
go build ./game/routes/...
```

- [ ] **Step 4: 提交**

```bash
git add game/routes/ routes/
git commit -m "refactor: 迁移 routes/ → game/routes/"
```

---

