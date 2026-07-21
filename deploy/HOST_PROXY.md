# 本机代理与 OpenAI OAuth 出网（部署备忘）

> 适用：本机 Docker 部署 sub2api（`docker-compose.local.yml` + `docker-compose.fork.yml`），上游账号为 **OpenAI OAuth**，依赖访问 `chatgpt.com`。  
> 目的：升级 / 重新部署后避免再次出现「测连通 EOF」却误以为是代码回归。

---

## 1. 典型症状

管理后台 **测试账号连接** 失败，日志类似：

```text
Account test error: Request failed: Post "https://chatgpt.com/backend-api/codex/responses": EOF
```

前端弹窗：

```text
Request failed: Post "https://chatgpt.com/backend-api/codex/responses": EOF
```

同时：

- `GET /health` 正常，`sub2api` 容器 healthy
- 账号状态仍是 `active`
- 接口 `POST /api/v1/admin/accounts/:id/test` 约 **10s** 后返回错误（不是业务 4xx 体）

**结论：多半是出网 / DNS / 未绑代理，不是镜像版本本身坏了。**

---

## 2. 根因（本机环境实测）

### 2.1 OAuth 测连通走哪条链路

OpenAI **OAuth** 账号测试不会打 `api.openai.com` 平台 API，而是 ChatGPT 内部 Codex：

- URL：`https://chatgpt.com/backend-api/codex/responses`
- 代码：`backend/internal/service/account_test_service.go`（`testOpenAIAccountConnection`）
- 若账号绑定了代理，会经 `account.Proxy.URL()` 出站（`DoWithTLS`）

### 2.2 容器直连失败的原因

1. **本机 Clash / Mihomo 开了 fake-ip DNS**  
   - Docker 内 `resolv.conf` 上游常见：`0.250.250.200`（Clash DNS）  
   - `chatgpt.com` 被解析成 `198.18.0.x` 等假 IP，TLS 握手被掐断 → `EOF` / `unexpected eof while reading`
2. **账号 `proxy_id` 为空**  
   - 测试与真实转发都 **直连**，撞上坏 DNS / 被墙线路
3. **Clash 默认 `allow-lan: false`**  
   - 代理只听本机进程场景；容器经 `host.docker.internal` 访问有时不稳，建议打开「允许局域网连接」

### 2.3 如何区分「代码问题」vs「网络问题」

| 检查 | 网络问题 | 代码/鉴权问题 |
|------|----------|----------------|
| 延迟 | ~8–15s 后 EOF | 通常更快返回 401/403/429 等 |
| 容器内直连 `chatgpt.com` | 超时 / SSL EOF | 能建连 |
| 经本机代理访问同 URL | 得到 HTTP 状态码（如 401） | 同左 |
| `accounts.proxy_id` | 多为 NULL | 已绑定仍失败 → 再查 token / 上游策略 |

快速验证（与 sub2api 同一 compose 网络）：

```bash
NET=$(docker inspect sub2api --format '{{range $k,$v := .NetworkSettings.Networks}}{{$k}}{{end}}')

# 经本机 Clash mixed-port（默认 7890）
docker run --rm --network "$NET" curlimages/curl:8.5.0 -sS \
  -x http://host.docker.internal:7890 --connect-timeout 10 --max-time 20 \
  -o /dev/null -w 'HTTP %{http_code} time %{time_total}\n' \
  -X POST 'https://chatgpt.com/backend-api/codex/responses' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer invalid' \
  -d '{"model":"gpt-5.1","input":"hi"}'
```

- **期望**：`HTTP 401`（假 token）且耗时 < 2s → 代理路径正常  
- **异常**：超时 / EOF → Clash 未开、端口不对、或未允许局域网

---

## 3. 本仓库当前推荐配置

### 3.1 代理记录（数据库 `proxies`）

| 字段 | 推荐值 |
|------|--------|
| name | `本机 Clash (host.docker.internal:7890)` |
| protocol | `http`（mixed-port 也可用 `socks5`/`socks5h`，优先 http） |
| host | `host.docker.internal` |
| port | `7890`（与 Clash Party / 系统代理一致） |
| status | `active` |

> OrbStack / Docker Desktop 下容器可通过 `host.docker.internal` 访问宿主机。  
> 不要写 `127.0.0.1`：在容器里那是容器自己，不是宿主机。

### 3.2 账号绑定

所有需要访问 ChatGPT/OpenAI 的账号应设置 `proxy_id` 指向上述代理。

SQL 示例（**仅维护用**，按需改端口/名称）：

```sql
-- 1) 确保本机代理存在
INSERT INTO proxies (name, protocol, host, port, status, fallback_mode, expiry_warn_days, created_at, updated_at)
SELECT
  '本机 Clash (host.docker.internal:7890)',
  'http',
  'host.docker.internal',
  7890,
  'active',
  'none',
  7,
  NOW(),
  NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM proxies
  WHERE deleted_at IS NULL
    AND host = 'host.docker.internal'
    AND port = 7890
);

-- 2) 未绑代理的账号全部挂上
UPDATE accounts a
SET proxy_id = p.id,
    updated_at = NOW()
FROM (
  SELECT id FROM proxies
  WHERE deleted_at IS NULL
    AND host = 'host.docker.internal'
    AND port = 7890
  ORDER BY id DESC
  LIMIT 1
) p
WHERE a.deleted_at IS NULL
  AND a.proxy_id IS NULL;
```

管理后台也可：**代理管理** 新建 → **账号** 批量编辑绑定代理。

### 3.3 Clash Party（Mihomo）建议长期设置

- **允许局域网连接**（`allow-lan: true`）  
  - 仅运行时 PATCH 会在 Clash 重启后丢失，请在 GUI 里持久打开
- mixed-port 保持 **7890**（或改完后同步改 `proxies.port`）
- 系统代理 / TUN 可开，但 **fake-ip 不能替代「账号级 HTTP 代理」**  
  - 容器 DNS 仍可能被劫持；**账号绑 `host.docker.internal:7890` 才是可靠方案**

运行时临时打开（重启 Clash 后失效）：

```bash
curl --unix-socket /tmp/mihomo-party-*.sock -sS -X PATCH http://localhost/configs \
  -H 'Content-Type: application/json' \
  -d '{"allow-lan":true}'
```

---

## 4. 升级 / 重新部署检查清单

每次 `git pull` + 重建 `sub2apix:local` + `compose up` 之后：

- [ ] 三个容器 healthy：`sub2api` / `postgres` / `redis`
- [ ] `curl -sS http://127.0.0.1:8080/health` → `{"status":"ok"}`
- [ ] **本机 Clash 已启动**，7890 可连
- [ ] Clash **允许局域网** 已开
- [ ] 数据库仍有本机代理：

  ```bash
  docker exec -i sub2api-postgres psql -U sub2api -d sub2api -c \
    "SELECT id,name,protocol,host,port,status FROM proxies WHERE deleted_at IS NULL;"
  ```

- [ ] OpenAI 账号均已绑代理：

  ```bash
  docker exec -i sub2api-postgres psql -U sub2api -d sub2api -c \
    "SELECT COUNT(*) FILTER (WHERE proxy_id IS NULL) AS no_proxy, COUNT(*) AS total
     FROM accounts WHERE deleted_at IS NULL AND platform = 'openai';"
  ```

  `no_proxy` 应为 `0`（或确认例外账号故意直连）

- [ ] 后台对 **一个 OAuth 账号** 点测连通：不应再 EOF  
  - 若 401：token 过期，刷新 OAuth  
  - 若 429：额度/限流，与代理无关  
  - 若 EOF：回到第 2 节网络排查

### 不会被升级冲掉的

| 项 | 是否持久 |
|----|----------|
| `deploy/data`、`postgres_data`、`redis_data` | ✅ 本地卷，compose 重建通常保留 |
| `proxies` / `accounts.proxy_id` | ✅ 在 Postgres 里，**镜像升级不丢** |
| 本机 Clash 配置 | ✅ 在 Clash Party，与 sub2api 无关 |
| 运行时 `allow-lan` PATCH | ❌ Clash 重启可能丢 |
| 新 **导入** 的账号 `proxy_id` | ❌ 默认空，需再绑或导入时指定 |

### 容易再踩的坑

1. **只更新代码/镜像，不检查代理** → OAuth 测试又 EOF  
2. **新导入一批 OAuth 账号未选代理**  
3. **Clash 关了或端口改了**，DB 里还是旧 `7890`  
4. **把 proxy host 写成 `127.0.0.1`**  
5. **以为 TUN/系统代理会自动作用于容器** — 不会，必须账号级代理或容器网络特殊方案  
6. **把 EOF 当成「升级引入的 bug」** — 先按第 2.3 节验证代理路径

---

## 5. 相关路径

| 路径 | 说明 |
|------|------|
| `deploy/docker-compose.local.yml` | 本地目录卷部署 |
| `deploy/docker-compose.fork.yml` | 覆盖镜像为 `sub2apix:local` |
| `deploy/.env` | 环境变量（**不含**账号级出站代理） |
| `backend/internal/service/account_test_service.go` | 测连通实现 |
| `backend/internal/service/proxy.go` | `Proxy.URL()` |

---

## 6. 变更记录

| 日期 | 说明 |
|------|------|
| 2026-07-21 | 升级至 upstream v0.1.162 后 OAuth 测连通 EOF；确认为 fake-ip DNS + 账号无代理。新增 `proxies` 本机 Clash，15 个 openai 账号绑定 `host.docker.internal:7890`；经代理 POST codex 返回 401 验证通路。本文档建立。 |
