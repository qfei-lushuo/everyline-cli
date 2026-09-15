# 安装与更新细则

本文件承接主 Skill 的安装与更新入口。执行前遵守[执行前检查](../SKILL.md#preflight)、[Profile 与身份](../SKILL.md#identity)和[通用边界](../SKILL.md#boundaries)；完成后返回原任务。

<a id="setup"></a>
## 安装与更新

<a id="skill-only-setup"></a>
### 仅安装 Skill 后的 CLI 依赖检查

仅安装、导入、复制或更新 EveryLine Skill，未同时安装 CLI 时，安装 Agent 必须在技能文件就位后、当前任务结束前主动执行[执行前检查](../SKILL.md#preflight)，检测当前任务执行环境中的 CLI。此流程由 Skill 安装完成触发，不等待用户另行提出 CLI 安装请求。

CLI 已存在且验证可用时，复用检查结果继续原请求，无需重复安装。CLI 缺失时，按下方“安装来源与执行”第 7 条读取安装文档并自动安装最新正式版；只有宿主权限策略或用户设置要求额外确认时，才明确提示“当前缺少 EveryLine CLI，是否安装最新正式版？”并等待确认。

不得仅报告 Skill 文件安装成功后结束。无法完成检查或安装时，说明实际状态、具体原因及待完成步骤；只有验证通过后才报告 CLI 已就绪。

### 安装来源与执行

执行 npm 安装、迁移、重建或延迟更新时，默认使用下列正常全局安装命令，不设置 EVERYLINE_SKIP_SKILL_INSTALL=1，以便安装器同步两个 Skill 并保留首次安装授权门禁。只有用户明确要求单独安装 CLI 时才为该次进程设置 EVERYLINE_SKIP_SKILL_INSTALL=1；不永久修改环境。CLI 返回的 updateCommand 同样遵守此规则。安装后核对两个 Skill 的实际文件、依赖及宿主加载状态，不依据版本号猜测同步成功。

1. 正式 npm 包名为 @qfeius/everyline-cli，可执行命令为 everyline-cli，本 Skill 名为 everyline-review。按用户本次指定的安装包、版本、发布下载地址或私有源选择目标，不把[主 Skill](../SKILL.md) 的 metadata.version 当作固定安装版本。
2. 已有宿主可读取的 .tgz 时，直接使用该文件执行下列命令；用户本机路径不等于云端沙箱路径。仅有文件名、不可读取引用或缺少来源时，先取得可用附件、下载地址或正确源，不猜测个人目录。
   ```bash
   npm install -g --foreground-scripts --allow-scripts=@qfeius/everyline-cli <实际安装包路径>
   ```
3. 用户未指定包、版本或源时，使用官方 npm：
   ```bash
   npm install -g --foreground-scripts --allow-scripts=@qfeius/everyline-cli @qfeius/everyline-cli@latest --registry https://registry.npmjs.org
   ```
   不因指定包与本地同版本、isLatest=true 或公共源 E404 跳过用户指定的安装。公开源返回 E404 或版本不存在时，报告该源未找到目标，不反复换源、重试或猜下载地址。
4. 保留 npm prefix、用户指定安装目录、Profile、身份和凭据；使用 --foreground-scripts 展示安装器输出，不添加 --silent，不单独替换二进制。
5. 只有 npm ls -g --depth=0 确认旧无 scope 包 everyline-cli 占用同一全局命令时，才执行一次带 --force 的新包安装以迁移。新包验证成功后卸载旧 npm 包，再运行下列命令恢复入口并重新验证；任一步失败即停止后续迁移，普通安装不使用 --force。
   ```bash
   npm rebuild -g --foreground-scripts --allow-scripts=@qfeius/everyline-cli @qfeius/everyline-cli
   ```
6. 按目标包的实际内容分别核对 CLI 和两个 Skill。只有目标包包含 everyline-review/SKILL.md 和 everyline-review-config/SKILL.md，且安装器实际同步到宿主时，才报告两个 Skill 同步成功。来源包结构不匹配、缺少任一 Skill 或安装器报告同名目录冲突时，保留当前技能，报告技能更新缺口；不删改用户目录或自动覆盖现有技能文件。
7. 本条先复用当前轮的 CLI 检查结果；尚未检查时，仅安装 Skill 的场景执行[仅安装 Skill 后的 CLI 依赖检查](#skill-only-setup)，其他场景执行[执行前检查](../SKILL.md#preflight)。检查确认当前执行环境缺少 `everyline-cli` 时，立即说明“检测到当前缺少 EveryLine CLI，将安装最新正式版。”并读取[CLI 安装文档](https://ysi13ckdb9.feishu.cn/wiki/Pyy3wLSOziN5c1ka12scOejhnag)，获取安装方式并下载安装最新正式版 CLI。用户已要求安装 Skill、CLI 或使用依赖 CLI 的功能时，将补齐 CLI 作为当前任务的必要步骤直接执行；只有宿主权限策略或用户设置要求额外确认时，才明确询问“检测到当前缺少 EveryLine CLI，是否安装最新正式版以完成配置？”，等待确认后继续，已有安装确认直接复用。安装版本以官方 npm 的 `@qfeius/everyline-cli@latest` 为准，不固定文档中的示例版本。文档无法读取时，直接使用本节的官方 npm 安装命令。安装完成后按现有规则验证，通过后继续原请求；失败时说明具体原因及未完成步骤，不把安装检查留给用户自行发起，也不声称 CLI 已就绪。

### 验证安装与宿主加载

- 安装或更新成功需有 npm 成功退出、目标 CLI 可执行、版本核对结果，以及两个 Skill 的 SKILL.md 可读取的证据。仅核实 CLI 时只报告 CLI 的实际状态，技能状态单独说明，不笼统报告全部完成。
- 指定 .tgz 时以包内版本和实际内容为验收目标；latest 查询失败不否定已验证的安装，不触发第二次安装。同版本内容差异只能说明构建不同，不据此判断新旧或损坏；可报告“已按指定包重新安装”，不称为发现新版。缺少打包时的版本同步脚本本身不代表运行故障。
- Codex、WorkBuddy 的 npm 目录链接，需核对实际指向与两个 Skill 的文件；界面导入副本单独核验。CLI 更新不证明手动导入的技能副本已更新。
- 对支持豆包同步的安装器，macOS 可识别已存在的 ~/Library/Application Support/DoubaoWork/Default/.doubaowork/agent_mode/workspace/.user_skills；其他平台或自定义工作区按宿主实际提供的 EVERYLINE_DOUBAO_SKILLS_DIR 指定绝对目录。未发现时不创建猜测路径。EVERYLINE_SKIP_DOUBAO_SKILL_INSTALL=1 仅跳过豆包，EVERYLINE_SKIP_SKILL_INSTALL=1 跳过全部宿主。
- 同版本也核对内容；需保留的旧副本应放在技能扫描目录外。安装器返回 event=skills_updated、host=doubao、nextAction=reload_skills 时，核对事件目标并重新读取两个 Skill 的 SKILL.md。实际文件同步不证明当前会话已加载；没有即时加载入口时提示新建任务。
- 豆包云端 ZIP 副本通过技能管理重新导入两个独立 Skill ZIP（每个 ZIP 包含对应技能的完整目录）。CLI .tgz 和含额外发布材料的外层包不作为技能导入包；尚待导入或重载的步骤明确列为未完成。
- npm 包装版通过 version --output json 查询官方 npm latest，无需额外 manifest；独立二进制安装使用 HTTPS manifest。只采用真实返回的更新信息。

### 安装故障处理

- 区分进程创建失败与 npm/postinstall 失败：未启动时说明“安装命令尚未启动”；中断且退出结果未确认时标记“待验证”，不假定未改动或已成功。
- 非权限类进程创建故障最多做一次同环境最小只读探测，例如 pwd。仍失败就停止自动安装尝试，说明阶段、原始错误、是否启动与恢复步骤；不反复等待、变换工具或重跑安装。文件可读取不代表执行环境正常。
- 明确权限拒绝或拦截时遵循宿主审批，不通过改换执行通道或提权规避。持续环境异常可建议重启当前任务或执行环境，不声称等待几秒必然恢复；用户反馈恢复后先做一次只读探测再继续。

### 首次使用引导

Codex、WorkBuddy 和豆包都按“Skill 文件就位 → 当前环境 CLI 检查 → 缺失时安装或取得必要确认 → 验证 → 回复正文展示结果”执行。安装任务不能停在 Skill 文件复制成功。终端日志或工具 JSON 不代替正文说明。同会话只展示一次；仅安装时放在最终回复，安装后继续授权或业务时在身份选择前展示。

- 已确认本会话首次安装、宿主首次导入且首次运行、用户明确首次使用，或 CLI 明确要求首次配置时，使用下方统一文案。
- 缺少 firstInstall/authorizationRequired 或字段为 false，不否定已确认的首次安装事实；字段缺失、未登录、无历史任务或无默认身份本身也不构成首次安装信号。
- 已确认升级时使用[更新完成引导](#update-guidance)；同版本重装不重新触发首次介绍，未完成授权门禁继续遵守。
- 豆包 ZIP 导入不执行 npm postinstall。Agent 负责安装或导入技能时，应在当前轮读取主 Skill 并立即检查该任务环境中的 CLI；缺失时按第 7 条安装或请求必要确认。纯平台静态导入且没有执行中的 Agent 时，不声称已经完成 CLI 检查；宿主下一次实际读取本 Skill 时立即补做。WorkBuddy 在当前宿主内验证，不要求用户去其他宿主或终端查看完成提示。

只有 CLI 可执行、版本检查及技能文件验证通过后，才展示以下安装完成文案；仅技能文件复制成功时不得展示。CLI 缺失且等待必要确认时，明确展示第 7 条的安装确认提示；未能检查或安装失败时说明真实状态。

原样展示：

EveryLine CLI 已安装完成。目前支持合同审查，以及审查清单、规则和规则分组配置。使用前需要先完成账号授权，我现在可以为你打开授权页面或生成授权链接。

仅要求安装时，展示后等待用户决定是否授权；同次请求已经要求继续授权或业务时，复用该目标，先完成身份选择，授权成功后恢复原步骤。已明确选择身份时不重复询问。

<a id="update-guidance"></a>
### 更新完成引导

安装验证完成后，原样展示：

EveryLine CLI 已更新完成。目前支持合同审查，以及审查清单、规则和规则分组配置。

复用更新前已确认的 Profile 与身份，按[身份规则](../SKILL.md#identity)补齐缺失选择后，执行一次 auth status --profile <profile> --as <identity> --output json。不得从安装退出码、token 文件存在或历史有效期推断状态。

- authenticated=false：原样展示“使用前需要先完成账号授权，我现在可以为你打开授权页面或生成授权链接。”已有登录意愿时继续，否则等待用户决定。
- authenticated=true：原样展示“当前已存在生效授权，可直接调用cli能力；”。
- 状态调用失败：报告真实错误，状态保持未知，不展示有效或失效分支。每次更新只展示一次完成文案和一个状态分支。
- 若更新发生在成功审查同一轮，相关说明在过程消息中展示，最终审查回复仍按[成功结果输出](../SKILL.md#review-output)保持统一结构。
