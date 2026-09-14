"use strict";

const assert = require("node:assert/strict");
const { execFileSync, spawnSync } = require("node:child_process");
const { copyFileSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } = require("node:fs");
const { tmpdir } = require("node:os");
const { join } = require("node:path");
const test = require("node:test");
const { runInNewContext } = require("node:vm");
const versions = require("../../scripts/package-version");
const pkg = require("../../package.json");

/**
 * 验证 release 发布的真实前置脚本以正式版本标签为准，并要求 latest 渠道。
 * 入参：无；读取实际工作流并隔离输出、包数据和进程环境。
 * 返回值：void，预发布、构建元数据或其他渠道通过检查时断言失败。
 */
test("release 发布前置检查以正式版本标签为准并要求 latest 渠道", () => {
  const workflow = readFileSync(join(__dirname, "../../.github/workflows/release.yml"), "utf8");
  const source = workflow.match(/node - "\$version" <<'NODE'\r?\n([\s\S]*?)\r?\n\s+NODE/)[1];
  const cases = [
    { tag: `v${pkg.version}`, channel: "latest", valid: true },
    { tag: "v0.1.11", channel: "latest", valid: true },
    { tag: "v0.1.11-blue.1", channel: "latest" },
    { tag: "v0.1.11-beta.1", channel: "latest" },
    { tag: "v0.1.11-release.1", channel: "latest" },
    { tag: "v0.1.11+build", channel: "latest" },
    { tag: "v01.1.11", channel: "latest" },
    { tag: "v0.1.11", channel: "blue" },
    { tag: "v0.1.11", channel: "beta" },
  ];
  for (const scenario of cases) {
    let output = "";
    const version = scenario.tag.slice(1);
    const run = () => runInNewContext(source, {
      process: { argv: ["node", "-", version], env: { GITHUB_OUTPUT: "fixture-output" } },
      require: (name) => {
        if (name === "node:fs") return { appendFileSync: (file, content) => { assert.equal(file, "fixture-output"); output += content; } };
        if (name === "./package.json") return { ...pkg, publishConfig: { tag: scenario.channel } };
        if (name === "./scripts/package-version") return versions;
        throw new Error(`未预期的依赖: ${name}`);
      },
    });
    if (scenario.valid) {
      run();
      assert.equal(output, `version=${version}\nchannel=latest\n`);
      assert.match(workflow, /npm version "\$version" --no-git-tag-version --allow-same-version/);
    } else {
      assert.throws(run, /发布|版本|渠道/);
      assert.equal(output, "");
    }
  }
});

/**
 * 验证 npm 源码发布入口强制正式版本与 latest 渠道，普通 CI 构建包仍可本地打包。
 * 入参：t（TestContext）负责临时包目录清理。
 * 返回值：void，发布绕过校验或 CI 制品被误拒绝时断言失败。
 */
test("npm 发布入口校验 latest 正式版本且保留 CI 本地打包", (t) => {
  const root = mkdtempSync(join(tmpdir(), "everyline-release-channel-"));
  t.after(() => rmSync(root, { recursive: true, force: true }));
  mkdirSync(join(root, "scripts"));
  for (const name of ["package-version.js", "verify-package-version.js"]) {
    copyFileSync(join(__dirname, "../../scripts", name), join(root, "scripts", name));
  }
  assert.equal(pkg.publishConfig.tag, "latest");
  assert.equal(pkg.scripts.prepublishOnly, "node scripts/verify-package-version.js --publish");
  for (const [version, tag, valid] of [[pkg.version, "latest", true], ["0.1.11", "latest", true], ["0.1.11", "blue", false], ["0.1.11-beta.1", "latest", false], ["0.1.11-blue.1", "latest", false], ["0.1.11+build", "latest", false], ["0.0.0-build-abcdef", "latest", false]]) {
    writeFileSync(join(root, "package.json"), JSON.stringify({ version, publishConfig: { tag } }));
    const result = spawnSync(process.execPath, [join(root, "scripts/verify-package-version.js"), "--publish"], { encoding: "utf8" });
    assert.equal(result.status === 0, valid, result.stderr);
    // CI 本地打包不发布 npm，仍接受可追溯的 build 版本。
    assert.doesNotThrow(() => execFileSync(process.execPath, [join(root, "scripts/verify-package-version.js")]));
  }
});

/**
 * 验证 npm 发布与其他环境保持一致，使用独立 Workflow 和 GitHub OIDC，不依赖长期 Token。
 * 入参：无；读取实际 npm 发布工作流。
 * 返回值：void，OIDC 权限、Tag 版本写入或 provenance 发布缺失时断言失败。
 */
test("npm 发布使用独立 OIDC 工作流", () => {
  const workflow = readFileSync(join(__dirname, "../../.github/workflows/npm-publish.yml"), "utf8");
  assert.match(workflow, /id-token:\s*write/);
  assert.match(workflow, /node-version:\s*"24"/);
  assert.match(workflow, /packageData\.version = process\.argv\[2\]/);
  assert.match(workflow, /npm publish --provenance --access public/);
  assert.doesNotMatch(workflow, /NPM_TOKEN|NODE_AUTH_TOKEN/);
});
