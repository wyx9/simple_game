### Task 2: 迁移 pkg/ → game/pkg/（改 import + 解除 config 循环依赖）

**Files:**
- Move: `pkg/logger.go` → `game/pkg/logger.go`
- Move: `pkg/msg.go` → `game/pkg/msg.go`
- Move: `pkg/netconn.go` → `game/pkg/netconn.go`
- Move: `pkg/wsconn.go` → `game/pkg/wsconn.go`
- Move: `pkg/conn_dial.go` → `game/pkg/conn_dial.go`
- Move: `pkg/uuid.go` → `game/pkg/uuid.go`
- Move: `pkg/redis.go` → `game/pkg/redis.go` (关键修改：函数签名)
- Move: `pkg/mysql.go` → `game/pkg/mysql.go` (关键修改：函数签名)

**Interfaces:**
- Consumes: 原 `pkg/` 下所有 .go 文件
- Produces: `game/pkg/` 包，`RedisStart(addr, password string, db int)`、`MysqlStart(dsn string)` 新签名

- [ ] **Step 1: 移动不依赖 config 的文件（批量 git mv）**

```bash
mkdir -p game/pkg
git mv pkg/logger.go game/pkg/logger.go
git mv pkg/msg.go game/pkg/msg.go
git mv pkg/netconn.go game/pkg/netconn.go
git mv pkg/wsconn.go game/pkg/wsconn.go
git mv pkg/conn_dial.go game/pkg/conn_dial.go
git mv pkg/uuid.go game/pkg/uuid.go
```

- [ ] **Step 2: 修改 redis.go — RedisStart 接受参数替代 config 导入**

```bash
git mv pkg/redis.go game/pkg/redis.go
```

然后编辑 `game/pkg/redis.go`，将：

```go
import (
    "context"
    "github.com/go-redis/redis/v8"
    "simple_game/config"
)

var RedisClient *redis.Client
var RCtx context.Context

func RedisStart() {
    addr, password, db := config.Conf.Redis.Addr, config.Conf.Redis.PassWord, config.Conf.Redis.DB
```

改为：

```go
import (
    "context"
    "github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client
var RCtx context.Context

func RedisStart(addr, password string, db int) {
```

- [ ] **Step 3: 修改 mysql.go — MysqlStart 接受 dsn 参数**

```bash
git mv pkg/mysql.go game/pkg/mysql.go
```

然后编辑 `game/pkg/mysql.go`，将：

```go
import (
    "fmt"
    _ "github.com/go-sql-driver/mysql"
    "github.com/jmoiron/sqlx"
    "simple_game/config"
)

var DB *sqlx.DB

func MysqlStart() {
    dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.Conf.MySql.User, config.Conf.MySql.PassWord, config.Conf.MySql.Addr, config.Conf.MySql.Port, config.Conf.MySql.DBName)
```

改为：

```go
import (
    _ "github.com/go-sql-driver/mysql"
    "github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

func MysqlStart(dsn string) {
```

并把 `database, err := sqlx.Open("mysql", dataSourceName)` 改为 `database, err := sqlx.Open("mysql", dsn)`。

- [ ] **Step 4: 修正已迁移文件中的内部 import 路径**

`game/pkg/` 内部互相引用从 `"simple_game/pkg"` → 不需要 import（同包），直接调用。

检查 `netconn.go:38` 的 `NewListener` 是否引用了自身包 — 确认它在同包调用 `newTCPListener`/`newWSListener`，无需修改。

检查 `wsconn.go` 中的 `ERROR(...)` 调用（同包函数），无需修改。

检查 `conn_dial.go` 中的 `tcpConn`/`wsConn` 引用（同包），无需修改。

- [ ] **Step 5: 编译验证**

```bash
go build ./game/pkg/...
```

- [ ] **Step 6: 提交**

```bash
git add game/pkg/ pkg/
git commit -m "refactor: 迁移 pkg/ → game/pkg/，RedisStart/MysqlStart 解除 config 依赖"
```

---

