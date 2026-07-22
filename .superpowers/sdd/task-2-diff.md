# Review package: 7ab4460..HEAD

## Commits
16b7b46 refactor: 迁移 pkg/ → game/pkg/，RedisStart/MysqlStart 解除 config 依赖

## Files changed
 {pkg => game/pkg}/conn_dial.go |  0
 {pkg => game/pkg}/logger.go    |  0
 {pkg => game/pkg}/msg.go       |  0
 game/pkg/mysql.go              | 20 ++++++++++++++++++++
 {pkg => game/pkg}/netconn.go   |  0
 {pkg => game/pkg}/redis.go     |  4 +---
 {pkg => game/pkg}/uuid.go      |  0
 {pkg => game/pkg}/wsconn.go    |  0
 pkg/mysql.go                   | 22 ----------------------
 9 files changed, 21 insertions(+), 25 deletions(-)

## Diff
diff --git a/pkg/conn_dial.go b/game/pkg/conn_dial.go
similarity index 100%
rename from pkg/conn_dial.go
rename to game/pkg/conn_dial.go
diff --git a/pkg/logger.go b/game/pkg/logger.go
similarity index 100%
rename from pkg/logger.go
rename to game/pkg/logger.go
diff --git a/pkg/msg.go b/game/pkg/msg.go
similarity index 100%
rename from pkg/msg.go
rename to game/pkg/msg.go
diff --git a/game/pkg/mysql.go b/game/pkg/mysql.go
new file mode 100644
index 0000000..7cef372
--- /dev/null
+++ b/game/pkg/mysql.go
@@ -0,0 +1,20 @@
+package pkg
+
+import (
+	"fmt"
+	_ "github.com/go-sql-driver/mysql"
+	"github.com/jmoiron/sqlx"
+)
+
+var DB *sqlx.DB
+
+func MysqlStart(dsn string) {
+	database, err := sqlx.Open("mysql", dsn)
+	if err != nil || database == nil {
+		fmt.Println("mysql conn - fail-", err)
+		return
+	} else {
+		INFO("mysql server start suc")
+	}
+	DB = database
+}
diff --git a/pkg/netconn.go b/game/pkg/netconn.go
similarity index 100%
rename from pkg/netconn.go
rename to game/pkg/netconn.go
diff --git a/pkg/redis.go b/game/pkg/redis.go
similarity index 79%
rename from pkg/redis.go
rename to game/pkg/redis.go
index 7ad3646..55ad4d3 100644
--- a/pkg/redis.go
+++ b/game/pkg/redis.go
@@ -1,26 +1,24 @@
 package pkg
 
 import (
 	"context"
 	"github.com/go-redis/redis/v8"
-	"simple_game/config"
 )
 
 //https://redis.uptrace.dev/guide/go-redis.html#installation
 
 var RedisClient *redis.Client
 
 var RCtx context.Context
 
-func RedisStart() {
-	addr, password, db := config.Conf.Redis.Addr, config.Conf.Redis.PassWord, config.Conf.Redis.DB
+func RedisStart(addr, password string, db int) {
 	rdb := redis.NewClient(&redis.Options{
 		Addr:     addr,
 		Password: password, // no password set
 		DB:       db,       // use default DB
 	})
 
 	RedisClient = rdb
 	RCtx = context.Background()
 	result, _ := RedisClient.Ping(RCtx).Result()
 	if result == "PONG" {
diff --git a/pkg/uuid.go b/game/pkg/uuid.go
similarity index 100%
rename from pkg/uuid.go
rename to game/pkg/uuid.go
diff --git a/pkg/wsconn.go b/game/pkg/wsconn.go
similarity index 100%
rename from pkg/wsconn.go
rename to game/pkg/wsconn.go
diff --git a/pkg/mysql.go b/pkg/mysql.go
deleted file mode 100644
index da8d268..0000000
--- a/pkg/mysql.go
+++ /dev/null
@@ -1,22 +0,0 @@
-package pkg
-
-import (
-	"fmt"
-	_ "github.com/go-sql-driver/mysql"
-	"github.com/jmoiron/sqlx"
-	"simple_game/config"
-)
-
-var DB *sqlx.DB
-
-func MysqlStart() {
-	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.Conf.MySql.User, config.Conf.MySql.PassWord, config.Conf.MySql.Addr, config.Conf.MySql.Port, config.Conf.MySql.DBName)
-	database, err := sqlx.Open("mysql", dataSourceName)
-	if err != nil || database == nil {
-		fmt.Println("mysql conn - fail-", err)
-		return
-	} else {
-		INFO("mysql server start suc")
-	}
-	DB = database
-}
