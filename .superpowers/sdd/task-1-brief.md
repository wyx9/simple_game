### Task 1: 创建三服目录骨架 + game 配置

**Files:**
- Create: `game/config/config.yaml`
- Create: `game/main.go`（骨架，后续任务填充）
- Create: `agent/config/config.yaml`
- Create: `world/config/config.yaml`

**Interfaces:**
- Produces: 各服务的 config.yaml 文件，供后续任务读取

- [ ] **Step 1: 创建 game/config/config.yaml**

```bash
mkdir -p game/config agent/config world/config
```

- [ ] **Step 2: 写入 game/config/config.yaml**

```yaml
# game/config/config.yaml
Listen:
  Addr: "127.0.0.1"
  Port: "9900"

Grpc:
  Addr: "127.0.0.1"
  Port: "9901"

SaveLog: false

TokenSecret: "shared-secret-key"

MySql:
  Addr: "127.0.0.1"
  Port: 3306
  User: "root"
  PassWord: "123456"
  DBName: "simple_game"

Redis:
  Addr: "127.0.0.1:6379"
  PassWord: ""
  DB: 4
```

- [ ] **Step 3: 写入 agent/config/config.yaml**

```yaml
# agent/config/config.yaml
Listeners:
  - Network: tcp
    Addr: "0.0.0.0"
    Port: "8888"
  - Network: ws
    Addr: "0.0.0.0"
    Port: "8889"

GameAddr: "127.0.0.1:9900"

Redis:
  Addr: "127.0.0.1:6379"
  PassWord: ""
  DB: 4
```

- [ ] **Step 4: 写入 world/config/config.yaml**

```yaml
# world/config/config.yaml
Http:
  Addr: "127.0.0.1"
  Port: "9902"

AgentAddr: "127.0.0.1:8888"
TokenSecret: "shared-secret-key"
TokenExpire: 24h

SaveLog: false

MySql:
  Addr: "127.0.0.1"
  Port: 3306
  User: "root"
  PassWord: "123456"
  DBName: "simple_game"

Redis:
  Addr: "127.0.0.1:6379"
  PassWord: ""
  DB: 4
```

- [ ] **Step 5: 创建 game/main.go 骨架（后续任务填充）**

```go
// game/main.go
package main

func main() {
	// 骨架 — 后续任务填充
}
```

- [ ] **Step 6: 提交**

```bash
git add agent/ world/ game/
git commit -m "feat: 创建三服目录骨架和配置文件"
```

---

