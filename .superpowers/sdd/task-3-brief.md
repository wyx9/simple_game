### Task 3: 迁移 libs/ → game/libs/

**Files:**
- Move: `libs/packet_msg.go` → `game/libs/packet_msg.go`
- Delete: `libs/packet_msg_test.go`（测试文件暂不迁移，后续补）

**Interfaces:**
- Produces: `game/libs` 包，导出 `PacketMsg`、`Pack2Msg`、`DeCodePack`、`EnCodePack`

- [ ] **Step 1: 移动文件**

```bash
mkdir -p game/libs
git mv libs/packet_msg.go game/libs/packet_msg.go
```

- [ ] **Step 2: 编译验证**

```bash
go build ./game/libs/...
```

`packet_msg.go` 只依赖标准库 + `google.golang.org/protobuf`，无内部 import，无需修改。

- [ ] **Step 3: 提交**

```bash
git add game/libs/ libs/
git commit -m "refactor: 迁移 libs/ → game/libs/"
```

---

