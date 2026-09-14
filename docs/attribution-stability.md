# 客户端来源识别稳定性（process-ancestry-v6）

EveryLine CLI 与 Contract CLI 使用相同采集和判定规则，仅产品编码不同。来源字段用于统计与诊断，不参与登录鉴权或业务权限判断。

- 先采集祖先进程路径、进程名和已有运行时组合证据，发布完整的初步报告，再执行签名检查。
- 已登记签名／包身份命中返回 high；签名变化、未登记或验证失败，不再否决明确产品路径（medium）或进程名（low）。此策略适用于 macOS 和 Windows 的所有已登记客户端。
- 多个不同产品的祖先进程路径同时命中时，返回 unknown / conflicting_process_evidence。国内与海外 WorkBuddy 仍属于同一产品。
- WorkBuddy macOS 同时登记 com.workbuddy.workbuddy 与 com.tencent.workbuddy.mac，Team ID 均为 FN2V63AD2J。资源校验仍通过 --ignore-resources 跳过。
- 探测仍限时 5 秒，辅助进程在签名检查前后各输出一份 JSON。若后续检查卡住或失败，父进程保留已收到的完整报告，并记录诊断 warning；没有收到有效报告才返回 unknown。业务 Context 取消规则不变，辅助进程仍被终止和回收。
- 云端 Linux 产品标记和分类不变；没有证据时不根据操作系统或本机安装的软件猜测平台。

验证覆盖：新旧 Bundle ID、身份变化与验签失败、Windows 证书变化、签名阶段超时后保留路径结果、成功后提升置信度、冲突路径保持 unknown，以及无报告超时、非法输出和进程回收。

单元测试与交叉编译不能代替 macOS / Windows 实机验收。需在 WorkBuddy 中分别使用两个 CLI 发请求，检查 detector_version=v6、来源 workbuddy 及相应置信度。
