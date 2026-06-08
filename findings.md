# sub2apix 迁移发现

## 旧部署状态

- 当前运行的 `sub2api` 容器使用镜像 `weishaw/sub2api:latest`。
- 旧部署 compose working dir 是 `/Users/calopteryx/code/sub2api/deploy`。
- 旧部署 compose 文件是 `/Users/calopteryx/code/sub2api/deploy/docker-compose.local.yml`。
- 旧 `sub2api` 容器挂载 `/Users/calopteryx/code/sub2api/deploy/data` 到 `/app/data`。
- 旧 Postgres 容器挂载 `/Users/calopteryx/code/sub2api/deploy/postgres_data` 到 `/var/lib/postgresql/data`。
- 旧 Redis 数据目录是 `/Users/calopteryx/code/sub2api/deploy/redis_data`。

## 新 fork 状态

- fork 仓库路径：`/Users/calopteryx/code/sub2apix`。
- git remote：`https://github.com/jiacongluo/sub2apix.git`。
- 已迁移 `.env`、`data`、`postgres_data`、`redis_data` 到 fork 的 `deploy` 目录。
- 新部署 compose working dir 是 `/Users/calopteryx/code/sub2apix/deploy`。
- 新部署 compose 文件是 `/Users/calopteryx/code/sub2apix/deploy/docker-compose.local.yml,/Users/calopteryx/code/sub2apix/deploy/docker-compose.fork.yml`。
- 当前 `sub2api` 容器使用镜像 `sub2apix:local`。
- 当前应用容器挂载 `/Users/calopteryx/code/sub2apix/deploy/data` 到 `/app/data`。
- 当前 Postgres 容器挂载 `/Users/calopteryx/code/sub2apix/deploy/postgres_data` 到 `/var/lib/postgresql/data`。
- 当前 Redis 容器挂载 `/Users/calopteryx/code/sub2apix/deploy/redis_data` 到 `/data`。
- CodeGraph 初始化生成了 `.codegraph/` 未跟踪目录；这不是部署数据。

## 清理结果

- 旧官方部署目录中的 `.env`、`data`、`postgres_data`、`redis_data` 已删除。
- 旧官方 Docker 镜像 `weishaw/sub2api:latest` 已删除。
- 旧官方部署遗留匿名卷 `4d0489f71738fffb96bb409baebdeb9a8e0e31d9d335d38120d0fe8347f4f5cb` 已删除。
- 新部署仍保留并使用匿名卷 `2a012f3c296d934c9b07e16498062c3de38b0afca836eeac2965e1342f07e7d0`。

## 重要风险

- 项目的网页在线更新功能固定查 `Wei-Shaw/sub2api` release。fork 部署不能通过该按钮更新，否则可能被官方二进制覆盖。
- `docker-compose.local.yml` 固定容器名为 `sub2api`、`sub2api-postgres`、`sub2api-redis`，新旧部署不能同时运行。
