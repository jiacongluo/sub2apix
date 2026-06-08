# sub2apix 迁移进度

## 2026-06-08

- 已审计当前 Docker 状态：官方 `weishaw/sub2api:latest` 正在运行。
- 已确认旧部署使用本地目录模式，数据在 `/Users/calopteryx/code/sub2api/deploy` 下。
- 已确认新 fork 目录还未部署。
- 已创建迁移计划文件、发现记录和进度记录。
- 已停止旧官方 compose 部署，`sub2api`、`sub2api-postgres`、`sub2api-redis` 均已从 Docker 容器列表移除。
- 已创建备份包：`/Users/calopteryx/code/sub2apix/deploy/migration-backups/sub2api-old-deploy-20260608-193907.tar.gz`。
- 备份包 sha256：`22b6e94e7e4e26905d05a06278e975859a637577c7d359c84a12cc2b87764c82`。
- 已复制旧 `.env`、`data`、`postgres_data`、`redis_data` 到 `/Users/calopteryx/code/sub2apix/deploy`。
- 新目录数据大小：`.env` 24K、`data` 38M、`postgres_data` 194M、`redis_data` 3.6M。
- 已创建 `deploy/docker-compose.fork.yml`，将 `sub2api` 服务镜像覆盖为 `sub2apix:local`。
- 已更新 `.dockerignore`，排除 `deploy/migration-backups/` 和 `.codegraph/`，避免构建上下文携带本地迁移/索引数据。
- 已成功构建 fork 镜像 `sub2apix:local`，镜像 digest：`sha256:b642ad13b677cd6189a569ad83b5af4810c48e4506dee79036ae4e80e364ec5c`。
- 已启动 fork 部署，当前运行容器为 `sub2api (sub2apix:local)`、`sub2api-postgres (postgres:18-alpine)`、`sub2api-redis (redis:8-alpine)`，均为 healthy。
- 健康接口 `http://127.0.0.1:8080/health` 返回 `{"status":"ok"}`。
- 已确认应用、Postgres、Redis 容器 compose working dir 均为 `/Users/calopteryx/code/sub2apix/deploy`。
- 已确认应用、Postgres、Redis 容器挂载均指向 `/Users/calopteryx/code/sub2apix/deploy` 下迁移后的数据目录。
- 已确认数据库核心表计数：`users=15`、`accounts=5149`、`api_keys=13`、`schema_migrations=183`。
- 已确认容器内二进制版本：`Sub2API 0.1.135 (commit: docker, built: 2026-06-08T11:43:40Z)`。
- 已删除旧官方部署目录中的 `.env`、`data`、`postgres_data`、`redis_data`。
- 已删除旧官方镜像 `weishaw/sub2api:latest`。
- 已删除旧官方匿名卷 `4d0489f71738fffb96bb409baebdeb9a8e0e31d9d335d38120d0fe8347f4f5cb`。
- 终审确认旧目录目标项、旧官方镜像、旧匿名卷均无残留，新 fork 部署仍 healthy。
