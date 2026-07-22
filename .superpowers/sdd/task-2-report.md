# Task 2 Report: 迁移 pkg/ → game/pkg/（改 import + 解除 config 循环依赖）

## Status: DONE

## Summary
Successfully migrated all 8 files from `pkg/` to `game/pkg/` with the required signature changes to break the circular dependency on `simple_game/config`.

## Changes Made

### Files Moved (pure moves, no modifications)
- `pkg/logger.go` → `game/pkg/logger.go`
- `pkg/msg.go` → `game/pkg/msg.go`
- `pkg/netconn.go` → `game/pkg/netconn.go`
- `pkg/wsconn.go` → `game/pkg/wsconn.go`
- `pkg/conn_dial.go` → `game/pkg/conn_dial.go`
- `pkg/uuid.go` → `game/pkg/uuid.go`

### Files Moved with Signature Changes

**redis.go** (`pkg/redis.go` → `game/pkg/redis.go`):
- Removed `"simple_game/config"` import
- Changed `RedisStart()` → `RedisStart(addr, password string, db int)`
- Removed `config.Conf.Redis.Addr/PassWord/DB` variable reads

**mysql.go** (`pkg/mysql.go` → `game/pkg/mysql.go`):
- Removed `"simple_game/config"` import
- Kept `"fmt"` import (still used by `fmt.Println` in error path)
- Changed `MysqlStart()` → `MysqlStart(dsn string)`
- Removed `fmt.Sprintf(...)` with config fields, uses `dsn` parameter directly

### Notes
- `mysql.go` detected as create+delete by git rather than rename (content changed enough to fall below rename similarity threshold), but the move is functionally correct.
- No internal import path changes needed — all files share `package pkg` and reference each other directly.

## Commit
- **Commit hash**: `16b7b46`
- **Message**: `refactor: 迁移 pkg/ → game/pkg/，RedisStart/MysqlStart 解除 config 依赖`

## Build Verification
- `go build ./game/pkg/...` passed with no errors.

## Concerns
- None for this task. Downstream files in `server/`, `client/`, `routes/`, `http/`, `main.go` still import `"simple_game/pkg"` — these will need to be updated to `"simple_game/game/pkg"` in subsequent tasks. They will not compile until updated.
