# 授权与恢复细则

执行前遵守主 Skill 的[首次安装授权门禁](../SKILL.md#first-install-auth)、[Profile 与身份](../SKILL.md#identity)、[点击授权入口](../SKILL.md#点击授权入口)和[通用边界](../SKILL.md#boundaries)。按已确认的身份与当前宿主读取适用分支；Device 会话准备必须早于首次身份相关配置和状态命令。授权恢复后继续原任务，不重新收集已确认输入。

<a id="user-auth"></a>
## user 授权

豆包与 WorkBuddy 先按“固定 Device 会话”准备运行时，再执行；这里及后文的每条 CLI 命令都必须携带已固定的会话变量：

```bash
everyline-cli auth status --profile <profile> --as user --output json
```

只有 `authenticated=true` 表示授权有效。未授权时按宿主选择流程：

| 宿主 | user 授权方式 | 运行时要求 |
| --- | --- | --- |
| Codex 本地任务 | `auth login` OAuth/PKCE | CLI 与浏览器共享本机 loopback |
| 豆包 AgentKit / Skills Sandbox | `auth init` + `auth complete` Device Grant | `SKILL_SESSION_WORKSPACE` 和平台注入的 `EVERYLINE_CLI_CREDENTIAL_KEY_V1` |
| 豆包普通工作任务（含“本地电脑”模式） | `auth init` + `auth complete` Device Grant | 首次 `auth status` 前固定 `SESSION_ID`，宿主缺失时只生成一次；每条命令显式传入并从同一初始工作目录执行 |
| WorkBuddy | `auth init` + `auth complete` Device Grant | 同一授权事务内固定的 `CODEBUDDY_SESSION_ID` 和系统凭证库 |

### Codex 本地 OAuth

读取 `auth login --help`，执行：

```bash
everyline-cli auth login --profile <profile> --as user --no-open-browser --timeout 3m
```

该命令会先读取当前 authorization server metadata 的 `registration_endpoint`，通过该端点动态注册浏览器 public client，再使用返回的 `client_id` 发起 OAuth/PKCE；即使 Profile 中留有旧 `oauth_client_id`，本次显式登录也会用新返回值替换它。Agent 不单独调用注册接口，不猜测注册路径，也不复用或改写输出中的 client ID。浏览器 client 与 `oauth_device_client_id` 分开保存，Codex 登录不覆盖 Device client。

Agent 保持该命令在同一个运行会话中等待 loopback callback，从 CLI 输出取得完整授权 URL，并按“点击授权入口”展示 `[点击授权](<FULL_AUTHORIZATION_URL>)`。用户主动点击并在浏览器完成授权；Agent 不调用系统浏览器打开命令。工具执行应保留后台会话，或提供至少三分钟的进程存活时间；取得链接后及时展示，工具短暂返回不代表 CLI 已退出。保持同一次登录会话，不为切换展示方式重启登录。

### 豆包与 WorkBuddy Device Grant

豆包沙箱与用户本机浏览器不共享网络命名空间；用户浏览器中的 `127.0.0.1:8000` 指向用户本机。豆包和 WorkBuddy Device 运行时不得执行 `auth login --profile <profile> --as user`，也不得等待 loopback callback。豆包“本地电脑”模式同样遵循此规则，不因 CLI 与浏览器都在本机而切换到 OAuth/PKCE。

固定流程为：CLI 执行 `auth init` → 将完整 HTTPS 授权链接生成为“点击授权”入口 → 用户主动点击并在任意浏览器批准 → 用户在新消息中确认“已授权” → CLI 执行一次 `auth complete` 查询账号服务并保存凭证。

<a id="device-session"></a>
### 固定 Device 会话

豆包普通工作任务（含“本地电脑”模式）在首次 `auth status` 前固定 `SESSION_ID` 和任务初始工作目录：优先复用宿主已有的稳定 `SESSION_ID`；宿主未提供时只生成一次 UUID，保存为当前任务上下文。不要每次命令都重新生成，也不要仅在一次 shell 中 `export` 后假定后续工具调用会继承。此后 `config`、`auth status`、`auth init`、`auth complete`、退出及所有 user 业务命令都显式传入同一 `SESSION_ID`，工具的工作目录始终设置为同一初始目录。审查或配置管理恢复执行时也必须携带这两个值。

```bash
# 各次工具调用的工作目录均为已记录的任务初始目录；占位符复用同一个已生成值。
SESSION_ID=<same-session-id> everyline-cli auth status --profile <profile> --as user --output json
SESSION_ID=<same-session-id> everyline-cli auth init --profile <profile> --as user --output json
# 用户在新消息中确认已授权后：
SESSION_ID=<same-session-id> everyline-cli auth complete --profile <profile> --as user --output json
SESSION_ID=<same-session-id> everyline-cli auth status --profile <profile> --as user --output json
```

豆包 AgentKit / Skills Sandbox 已提供 `SKILL_SESSION_WORKSPACE` 时，保留平台工作区与注入密钥，不用生成的 `SESSION_ID` 替换该模式。运行时缺失提示应通过补齐并复用会话变量解决；豆包不得照抄缺少会话变量时返回的本地 `auth login` 建议。`source=cache` 的历史本机 OAuth 状态不作为豆包 Device 授权成功依据；准备好运行时后以 `source=device` 查询当前会话状态。

WorkBuddy 在第一次 `auth init` 前取得并冻结一个非敏感的 `CODEBUDDY_SESSION_ID`：优先记录宿主已有的稳定值；宿主未提供时只生成一次，并把该值保留为当前授权事务状态。首次 `auth status` 之前也必须完成此准备，避免查询到本地 OAuth 缓存。`auth init`、`auth complete` 和随后的 `auth status` 都显式复用完全相同的值，不使用每次命令都会变化的通用 `SESSION_ID`。等价调用形式如下，其中两处 `<same-session-id>` 必须逐字相同：

```bash
CODEBUDDY_SESSION_ID=<same-session-id> everyline-cli auth init --profile <profile> --as user --output json
CODEBUDDY_SESSION_ID=<same-session-id> everyline-cli auth complete --profile <profile> --as user --output json
```

如果 `auth complete` 返回“没有待完成的 Device 授权”，先恢复 `auth init` 使用的原 `CODEBUDDY_SESSION_ID` 并重试一次 `auth complete`；这次本地存储未命中的失败没有请求 token endpoint，不计作重复兑换。不得因此直接执行 `auth init --restart`。只有 CLI 明确返回 `denied`、`expired` 或 `invalid_grant`，并且用户同意重新授权时，才开始新事务；原标识已经丢失时先如实说明事务状态丢失并等待用户决定。

豆包遇到同样的本地事务未命中时，先恢复原 `SESSION_ID` 和初始工作目录，再重试一次 `auth complete`，遵循相同的新事务规则。

### 发起与完成 Device 授权

prod 预设固定使用 `business_type=contract-review`、`scope=contract-review:full` 和对应开放平台 resource。Codex `auth login` 按上节规则动态注册浏览器 client，并始终走 OAuth Authorization Code + PKCE。豆包和 WorkBuddy 始终执行 `auth init`/`auth complete` Device Grant；prod 预设提供独立 EveryLine Device client `zscli_bc60fee4de9913ae`，`auth init` 不动态注册 client，也不复用 Codex 浏览器 client。自定义环境使用 Device Grant 时，Profile 必须配置平台确认的 `oauth_device_client_id`。两类 client 分开使用，Agent 不在两种授权方式之间复制 client ID。

读取 `auth init --help` 和 `auth complete --help`。首次安装门禁期间执行：

```bash
everyline-cli auth init --restart --profile <profile> --as user --output json
```

非首次安装的普通未授权流程执行一次：

```bash
everyline-cli auth init --profile <profile> --as user --output json
```

- 将 `verification_uri_complete` 作为不可拆分的完整 HTTPS URL，按“点击授权入口”生成 `[点击授权](<verification_uri_complete>)`；展示文字可以隐藏 URL，但链接目标必须逐字一致，不省略、拆分、解码或重拼 query。
- `auth init` 返回 `reused=true` 表示复用了已经存在的 Device 事务；同一会话已经展示过该链接时不再次生成按钮、链接或浏览器页面。首次安装事件重放、状态复查或工具重试都不得触发第二次 `auth init --restart`。
- 展示链接后结束当前轮次。只有用户在新消息中明确表示已完成浏览器授权，才执行一次 `auth complete`。
- `pending` 表示仍待用户完成；结束本轮，不持续轮询。
- `succeeded` 后重新执行 `auth status`，再继续被中断的业务步骤一次。
- `denied`、`expired`、`invalid_grant` 先说明真实状态；用户明确同意重新授权后使用 `auth init --restart`。
- `uncertain` 不重复兑换同一个 device code；保留状态并停止当前业务操作。
- metadata 缺少 `device_authorization_endpoint` 时原样报告认证服务能力缺口，不在远端沙箱回退到 loopback 登录，也不猜测 endpoint 或 client ID。

Device code、access token、refresh token 和加密密钥不进入对话、日志或普通配置。`EVERYLINE_CLI_CREDENTIAL_KEY_V1` 由豆包运行平台稳定注入，不在会话中临时生成或展示。

<a id="app-auth"></a>
## app 授权

用户已选择 app 后，先按下方顺序固定 app ID 和 Profile，再执行 `auth status --profile <profile> --as app --output json`。未授权时读取 `auth login --help`，只采用帮助中真实存在的安全入口：

- app 授权先取得并固定非敏感的 app ID，再进入 app secret 输入；两个连续阶段的顺序不得颠倒。当前消息和目标 Profile 都没有 app ID 时，只询问一次 app ID；该轮不同时请求 app secret。已有唯一 app ID 时直接复用，不重复询问。
- 取得 app ID 后先创建或校验目标 Profile，再查询 app 授权状态。只有 `authenticated=false` 时才发起一次 `auth login --as app`；同一次登录事务只输入一次 app secret。登录命令结束后只执行一次 `auth status` 验证，不再次启动登录或要求第二次输入 secret。
- 登录失败时保留已固定的 Profile 和 app ID，报告原始错误后结束本次尝试；不自动重跑 `auth login`。只有用户随后明确要求重试时才开始一笔新的登录事务，并在该新事务中输入一次 app secret。
- Codex 本地在 app ID 固定后，让用户在自己的终端通过 `--app-secret-stdin` 隐藏输入一次；豆包在 app ID 固定后使用平台密钥入口完成一次安全输入或注入，再由 Agent 发起一次登录；WorkBuddy 按下方专用终端命令执行。三个宿主都不在对话里收集 app secret。
- WorkBuddy 不使用对话文字输入、`AskUserQuestion`、选项卡或 Agent 捕获的 stdin 收集 app secret。取得非敏感的 Profile、app ID 和当前 WorkBuddy Node `bin` 目录后，把下面的一行命令替换成真实值并完整展示，让用户在自己的 WorkBuddy 终端亲自执行，然后结束当前轮等待用户确认：

```bash
export PATH=<WORKBUDDY_NODE_BIN>:$PATH && everyline-cli auth login --profile <profile> --as app --app-id <app-id> --app-secret-stdin
```

- 命令启动后 CLI 显示 `App secret:`，用户直接输入并按回车；终端不回显字符。Agent 不代为执行这条登录命令，不读取、转发或复述输入内容。
- 用户回复已完成后，只执行 `auth status --profile <profile> --as app --output json` 验证；以 `authenticated=true` 为成功依据，不要求用户提供登录输出。
- `<WORKBUDDY_NODE_BIN>` 使用当前 WorkBuddy 实际 Node 可执行文件所在目录，例如 `/Users/<user>/.workbuddy/binaries/node/versions/<version>/bin`，不固定用户名或 Node 版本。
- Codex 本地或 CI 仍可使用由用户直接操作的隐藏输入或管道形式的 `--app-secret-stdin`；Agent 捕获 stdin 时让用户在自己的终端完成输入。
- 需要重新输入 secret 时，只让用户在自己的终端或平台密钥入口操作；不得要求用户在对话中提供、粘贴或转述 app secret，也不得以“告诉我如何获取”为由索取其内容。
- 平台托管网页或剪贴板入口仅在实时帮助明确注册且用户选择后使用。
- app secret 不放入命令参数、普通环境变量、输入 JSON、日志或对话。
- 持久化只交给 CLI 支持的安全存储；登录后再次以 `authenticated=true` 判定成功。
- user 与 app 凭据按身份独立保存；发起 app 授权不得先调用 user 的 `auth logout`。发起或重试 app 授权只显式使用 `--as app`，也不得把退出登录当作身份切换步骤。
- app token 过期只表示当前 token 不可继续使用；不推断凭据已变更或轮换，也不证明已保存的 app ID 或 app secret 无效。
- 服务端返回 `http=200 code=10003 msg=invalid param` 时原样报告通用参数错误及已有 request ID；除非 CLI 结构化结果明确指出具体凭据字段，不得将其归因为 app ID 或 app secret 错误。
- app 授权失败后保持原 Profile 和 app 身份，不自动建议改用其他 Profile 或 user 身份；下一步仅提示用户在自己的终端通过实时帮助确认的安全入口重试，或等待用户主动指定新的 Profile/身份。

<a id="auth-recovery"></a>
## 状态、退出与恢复

- 状态：只展示身份、授权状态、必要到期状态和下一步，不展示凭据来源、存储路径或敏感错误上下文。
- 退出：只有用户明确要求时执行对应身份的 `auth logout`；先读取帮助，执行后回读状态。
- 未授权或凭据过期：原身份已经由用户在本次流程中明确选择时继续该身份；否则先完成 user/app 单选再重新授权，成功后只重试原业务操作一次。
- user Token 进入五分钟刷新窗口且刷新失败、或明确过期时，保留当前 Profile、user 身份和宿主会话，重新生成一次授权链接供用户手动登录。Codex 执行一次 `auth login --profile <profile> --as user --no-open-browser --timeout 3m`，按上节方式保留进程并及时展示链接；豆包/WorkBuddy 按错误提示执行一次 `auth init --restart --profile <profile> --as user --output json`，按本 Skill 的授权入口规则展示新返回的完整 URL，待用户在新消息中确认完成后执行一次 `auth complete`。
- 用户已明确要求“过期后重新生成授权链接手动登录”时直接按该策略恢复，不重复确认重新授权意愿。恢复操作尚未开始时按错误提示执行一次 --restart；已生成待完成事务后复用原链接，后续只执行 auth complete，只有该事务明确为 `expired`、`denied` 或 `invalid_grant` 时才执行一次 `auth init --restart`，不因状态复查反复生成链接。新链接使用 CLI 的本次输出，不复用历史 `user_code` 或 OAuth URL。
- 预先约定的“过期后重新授权”覆盖 Token 到期、五分钟窗口刷新失败及事务 expired；事务 denied 或 invalid_grant 时，先说明状态并取得针对该情况的重新授权意愿。用户本轮已经明确覆盖该情况时直接复用，不重复询问；任何尚待完成的事务均不因状态复查而重建。
- 手动重新授权后执行 `auth status --profile <profile> --as user --output json`，确认 `authenticated=true` 后只恢复原业务步骤一次；授权期间暂停业务操作。进入到期前五分钟窗口后，缺少 refresh token 或刷新失败时，按 CLI 提示手动重新授权。
- 身份不匹配：展示当前身份，由用户决定是否切换。
- 权限不足：保留 CLI 返回的缺失范围和 request ID，不改换身份或绕过检查。
- 网络或服务错误：保留真实错误码和 request ID，不包装为授权成功。
- 服务端可信 `code=110004` 由 CLI 内部触发至多一次刷新和原请求重放；Agent 不额外重复写请求。
