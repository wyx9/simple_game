### Task 4: 迁移 api/protos/pt/ → game/api/protos/pt/

**Files:**
- Move: `api/protos/pt/` 下所有 .go 文件 → `game/api/protos/pt/`

- [ ] **Step 1: 移动文件**

```bash
mkdir -p game/api/protos/pt
git mv api/protos/pt/*.go game/api/protos/pt/
```

- [ ] **Step 2: 验证编译**

```bash
go build ./game/api/protos/pt/...
```

纯 protobuf 生成代码，无内部 import，无需修改。

- [ ] **Step 3: 提交**

```bash
git add game/api/ api/
git commit -m "refactor: 迁移 api/protos/pt/ → game/api/protos/pt/"
```

---

