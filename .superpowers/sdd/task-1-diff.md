# Review package: 67ff75b..HEAD

## Commits
7ab4460 feat: 创建三服目录骨架和配置文件

## Files changed
 agent/config/config.yaml | 15 +++++++++++++++
 game/config/config.yaml  | 24 ++++++++++++++++++++++++
 game/main.go             |  6 ++++++
 world/config/config.yaml | 22 ++++++++++++++++++++++
 4 files changed, 67 insertions(+)

## Diff
diff --git a/agent/config/config.yaml b/agent/config/config.yaml
new file mode 100644
index 0000000..af24504
--- /dev/null
+++ b/agent/config/config.yaml
@@ -0,0 +1,15 @@
+# agent/config/config.yaml
+Listeners:
+  - Network: tcp
+    Addr: "0.0.0.0"
+    Port: "8888"
+  - Network: ws
+    Addr: "0.0.0.0"
+    Port: "8889"
+
+GameAddr: "127.0.0.1:9900"
+
+Redis:
+  Addr: "127.0.0.1:6379"
+  PassWord: ""
+  DB: 4
diff --git a/game/config/config.yaml b/game/config/config.yaml
new file mode 100644
index 0000000..7348710
--- /dev/null
+++ b/game/config/config.yaml
@@ -0,0 +1,24 @@
+# game/config/config.yaml
+Listen:
+  Addr: "127.0.0.1"
+  Port: "9900"
+
+Grpc:
+  Addr: "127.0.0.1"
+  Port: "9901"
+
+SaveLog: false
+
+TokenSecret: "shared-secret-key"
+
+MySql:
+  Addr: "127.0.0.1"
+  Port: 3306
+  User: "root"
+  PassWord: "123456"
+  DBName: "simple_game"
+
+Redis:
+  Addr: "127.0.0.1:6379"
+  PassWord: ""
+  DB: 4
diff --git a/game/main.go b/game/main.go
new file mode 100644
index 0000000..5cf236f
--- /dev/null
+++ b/game/main.go
@@ -0,0 +1,6 @@
+// game/main.go
+package main
+
+func main() {
+	// 骨架 — 后续任务填充
+}
diff --git a/world/config/config.yaml b/world/config/config.yaml
new file mode 100644
index 0000000..ecac5d6
--- /dev/null
+++ b/world/config/config.yaml
@@ -0,0 +1,22 @@
+# world/config/config.yaml
+Http:
+  Addr: "127.0.0.1"
+  Port: "9902"
+
+AgentAddr: "127.0.0.1:8888"
+TokenSecret: "shared-secret-key"
+TokenExpire: 24h
+
+SaveLog: false
+
+MySql:
+  Addr: "127.0.0.1"
+  Port: 3306
+  User: "root"
+  PassWord: "123456"
+  DBName: "simple_game"
+
+Redis:
+  Addr: "127.0.0.1:6379"
+  PassWord: ""
+  DB: 4
