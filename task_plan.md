# sub2apix 本地部署迁移计划

## 目标

将当前本地运行的官方 `sub2api` 部署迁移为 fork 仓库 `/Users/calopteryx/code/sub2apix` 构建出的本地部署；旧官方部署需要移除，不保留运行容器或旧数据目录残留；业务数据必须完好迁移并验证。

## 当前阶段

- [x] 阶段 1：审计旧部署和新目录状态
- [x] 阶段 2：停止旧部署并创建一致性备份
- [x] 阶段 3：迁移 `.env`、`data`、`postgres_data`、`redis_data` 到 fork 目录
- [x] 阶段 4：构建 fork 镜像并用新目录启动
- [x] 阶段 5：验证健康状态、挂载路径、镜像来源、数据库可读
- [x] 阶段 6：清理旧部署残留并最终审计

## 决策

- 使用 `deploy/docker-compose.local.yml`，因为旧部署已经使用本地目录挂载，最利于数据迁移。
- 不使用网页在线更新功能，因为它固定拉官方 `Wei-Shaw/sub2api` release，可能覆盖 fork 改动。
- fork 镜像使用本地 tag `sub2apix:local`，通过 `deploy/docker-compose.fork.yml` 覆盖官方镜像。
- 迁移前先停旧容器，避免 Postgres/Redis 数据目录复制时不一致。

## 验证清单

- `sub2api` 容器镜像为 `sub2apix:local`。
- `sub2api`、`sub2api-postgres`、`sub2api-redis` 的 compose working dir 为 `/Users/calopteryx/code/sub2apix/deploy`。
- `sub2api` 挂载 `/Users/calopteryx/code/sub2apix/deploy/data`。
- Postgres 挂载 `/Users/calopteryx/code/sub2apix/deploy/postgres_data`。
- Redis 挂载 `/Users/calopteryx/code/sub2apix/deploy/redis_data`。
- `http://127.0.0.1:8080/health` 返回成功。
- 数据库可连接，至少能读出核心表计数。
- `/Users/calopteryx/code/sub2api/deploy/{.env,data,postgres_data,redis_data}` 已移除或归档，不作为部署残留存在。
- 旧官方镜像 `weishaw/sub2api:latest` 已移除。
- 旧匿名 Docker 卷 `4d0489f71738fffb96bb409baebdeb9a8e0e31d9d335d38120d0fe8347f4f5cb` 已移除。

## 错误记录

| 时间 | 错误 | 处理 |
| --- | --- | --- |
| 2026-06-08 | `rtk test -f ...` 输出 shell 帮助且退出 2 | `test` 是 shell 内建，改用 `du/ls` 结果确认 `.env` 已存在 |
| 2026-06-08 | `docker exec sub2api /app/sub2api version` 被当作启动参数，尝试监听 8080 并退出 | 源码确认版本参数是 `-version`，改用 `/app/sub2api -version` 成功输出版本 |
