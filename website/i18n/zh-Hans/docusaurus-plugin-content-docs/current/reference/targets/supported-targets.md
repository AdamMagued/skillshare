---
sidebar_position: 2
---

# Supported Targets

skillshare 开箱即用支持的 AI CLI 完整列表。

## 概览

skillshare 支持 **65+ 种 AI CLI 工具**。运行 `skillshare init` 时，它会自动检测并配置任何已安装的工具。

---

## 内置 Target

以下工具会在 `skillshare init` 期间被自动检测：

<div className="target-grid">
  <a className="target-badge" href="#target-adal">AdaL</a>
  <a className="target-badge" href="#target-universal">Universal</a>
  <a className="target-badge" href="#target-amp">Amp</a>
  <a className="target-badge" href="#target-antigravity">Antigravity</a>
  <a className="target-badge" href="#target-antigravity-cli">Antigravity CLI</a>
  <a className="target-badge" href="#target-aiderdesk">AiderDesk</a>
  <a className="target-badge" href="#target-astrbot">AstrBot</a>
  <a className="target-badge" href="#target-augment">Augment</a>
  <a className="target-badge" href="#target-autohand-code">Autohand Code</a>
  <a className="target-badge" href="#target-bob">Bob</a>
  <a className="target-badge" href="#target-claude">Claude</a>
  <a className="target-badge" href="#target-codearts">CodeArts</a>
  <a className="target-badge" href="#target-cline">Cline</a>
  <a className="target-badge" href="#target-codebuddy">CodeBuddy</a>
  <a className="target-badge" href="#target-codestudio">Code Studio</a>
  <a className="target-badge" href="#target-comate">COMATE</a>
  <a className="target-badge" href="#target-codex">Codex</a>
  <a className="target-badge" href="#target-commandcode">Cmd Code</a>
  <a className="target-badge" href="#target-continue">Continue</a>
  <a className="target-badge" href="#target-copilot">Copilot</a>
  <a className="target-badge" href="#target-cortex">Cortex</a>
  <a className="target-badge" href="#target-crush">Crush</a>
  <a className="target-badge" href="#target-cursor">Cursor</a>
  <a className="target-badge" href="#target-deepagents">Deep Agents</a>
  <a className="target-badge" href="#target-devin">Devin</a>
  <a className="target-badge" href="#target-dexto">Dexto</a>
  <a className="target-badge" href="#target-droid">Droid</a>
  <a className="target-badge" href="#target-firebender">Firebender</a>
  <a className="target-badge" href="#target-forgecode">ForgeCode</a>
  <a className="target-badge" href="#target-fx">fx</a>
  <a className="target-badge" href="#target-gemini">Gemini CLI</a>
  <a className="target-badge" href="#target-goose">Goose</a>
  <a className="target-badge" href="#target-grok">Grok</a>
  <a className="target-badge" href="#target-hermes">Hermes</a>
  <a className="target-badge" href="#target-iflow">iFlow</a>
  <a className="target-badge" href="#target-jazz">Jazz</a>
  <a className="target-badge" href="#target-junie">Junie</a>
  <a className="target-badge" href="#target-kilocode">Kilocode</a>
  <a className="target-badge" href="#target-kimchi">Kimchi</a>
  <a className="target-badge" href="#target-kimi">Kimi</a>
  <a className="target-badge" href="#target-kimi-code">Kimi Code</a>
  <a className="target-badge" href="#target-kiro">Kiro</a>
  <a className="target-badge" href="#target-kode">Kode</a>
  <a className="target-badge" href="#target-letta">Letta</a>
  <a className="target-badge" href="#target-lingma">Lingma</a>
  <a className="target-badge" href="#target-mcpjam">MCPJam</a>
  <a className="target-badge" href="#target-mux">Mux</a>
  <a className="target-badge" href="#target-neovate">Neovate</a>
  <a className="target-badge" href="#target-omp">oh-my-pi</a>
  <a className="target-badge" href="#target-ona">Ona</a>
  <a className="target-badge" href="#target-openclaw">OpenClaw</a>
  <a className="target-badge" href="#target-opencode">OpenCode</a>
  <a className="target-badge" href="#target-openhands">OpenHands</a>
  <a className="target-badge" href="#target-pi">Pi</a>
  <a className="target-badge" href="#target-pochi">Pochi</a>
  <a className="target-badge" href="#target-posit-assistant">Posit Assistant</a>
  <a className="target-badge" href="#target-purecode">Purecode AI</a>
  <a className="target-badge" href="#target-qoder">Qoder</a>
  <a className="target-badge" href="#target-qoder-cn">Qoder CN</a>
  <a className="target-badge" href="#target-qwen">Qwen</a>
  <a className="target-badge" href="#target-replit">Replit</a>
  <a className="target-badge" href="#target-reasonix">Reasonix</a>
  <a className="target-badge" href="#target-roo">Roo</a>
  <a className="target-badge" href="#target-rovodev">Rovo Dev</a>
  <a className="target-badge" href="#target-tabnine">Tabnine</a>
  <a className="target-badge" href="#target-trae">Trae</a>
  <a className="target-badge" href="#target-trae-cn">Trae CN</a>
  <a className="target-badge" href="#target-vibe">Vibe</a>
  <a className="target-badge" href="#target-verdent">Verdent</a>
  <a className="target-badge" href="#target-warp">Warp</a>
  <a className="target-badge" href="#target-windsurf">Windsurf</a>
  <a className="target-badge" href="#target-witsy">Witsy</a>
  <a className="target-badge" href="#target-xcode-claude">Xcode Claude</a>
  <a className="target-badge" href="#target-xcode-codex">Xcode Codex</a>
  <a className="target-badge" href="#target-zcode">ZCode</a>
  <a className="target-badge" href="#target-zed">Zed</a>
  <a className="target-badge" href="#target-zencoder">Zencoder</a>
</div>

---

## Target 路径

<table>
<thead>
<tr><th>Target</th><th>全局路径</th><th>项目路径</th></tr>
</thead>
<tbody>
<tr id="target-adal"><td>adal</td><td><code>&#126;/.adal/skills</code></td><td><code>.adal/skills</code></td></tr>
<tr id="target-universal"><td>universal</td><td><code>&#126;/.agents/skills</code></td><td><code>.agents/skills</code></td></tr>
<tr id="target-amp"><td>amp</td><td><code>&#126;/.config/agents/skills</code></td><td><code>.agents/skills</code></td></tr>
<tr id="target-antigravity"><td>antigravity</td><td><code>&#126;/.gemini/config/skills</code></td><td><code>.agents/skills</code></td></tr>
<tr id="target-antigravity-cli"><td>antigravity-cli</td><td><code>&#126;/.gemini/antigravity-cli/skills</code></td><td><code>.agents/skills</code></td></tr>
<tr id="target-aiderdesk"><td>aiderdesk</td><td><code>&#126;/.aider-desk/skills</code></td><td><code>.aider-desk/skills</code></td></tr>
<tr id="target-astrbot"><td>astrbot</td><td><code>&#126;/.astrbot/data/skills</code></td><td><code>data/skills</code></td></tr>
<tr id="target-augment"><td>augment</td><td><code>&#126;/.augment/skills</code></td><td><code>.augment/skills</code></td></tr>
<tr id="target-autohand-code"><td>autohand-code</td><td><code>&#126;/.autohand/skills</code></td><td><code>.autohand/skills</code></td></tr>
<tr id="target-bob"><td>bob</td><td><code>&#126;/.bob/skills</code></td><td><code>.bob/skills</code></td></tr>
<tr id="target-claude"><td>claude</td><td><code>&#126;/.claude/skills</code></td><td><code>.claude/skills</code></td></tr>
<tr id="target-codearts"><td>codearts</td><td><code>&#126;/.codeartsdoer/skills</code></td><td><code>.codeartsdoer/skills</code></td></tr>
<tr id="target-cline"><td>cline</td><td><code>&#126;/.cline/skills</code></td><td><code>.cline/skills</code></td></tr>
<tr id="target-codebuddy"><td>codebuddy</td><td><code>&#126;/.codebuddy/skills</code></td><td><code>.codebuddy/skills</code></td></tr>
<tr id="target-codestudio"><td>codestudio</td><td><code>&#126;/.codestudio/skills</code></td><td><code>.codestudio/skills</code></td></tr>
<tr id="target-comate"><td>comate</td><td><code>&#126;/.comate/skills</code></td><td><code>.comate/skills</code></td></tr>
<tr id="target-codex"><td>codex</td><td><code>&#126;/.agents/skills</code></td><td><code>.agents/skills</code></td></tr>
<tr id="target-commandcode"><td>commandcode</td><td><code>&#126;/.commandcode/skills</code></td><td><code>.commandcode/skills</code></td></tr>
<tr id="target-continue"><td>continue</td><td><code>&#126;/.continue/skills</code></td><td><code>.continue/skills</code></td></tr>
<tr id="target-cortex"><td>cortex</td><td><code>&#126;/.snowflake/cortex/skills</code></td><td><code>.cortex/skills</code></td></tr>
<tr id="target-copilot"><td>copilot</td><td><code>&#126;/.copilot/skills</code></td><td><code>.github/skills</code></td></tr>
<tr id="target-crush"><td>crush</td><td><code>&#126;/.config/crush/skills</code></td><td><code>.crush/skills</code></td></tr>
<tr id="target-cursor"><td>cursor</td><td><code>&#126;/.cursor/skills</code></td><td><code>.agents/skills</code></td></tr>
<tr id="target-deepagents"><td>deepagents</td><td><code>&#126;/.deepagents/agent/skills</code></td><td><code>.deepagents/skills</code></td></tr>
<tr id="target-devin"><td>devin</td><td><code>&#126;/.config/devin/skills</code></td><td><code>.devin/skills</code></td></tr>
<tr id="target-dexto"><td>dexto</td><td><code>&#126;/.agents/skills</code></td><td><code>.agents/skills</code></td></tr>
<tr id="target-droid"><td>droid</td><td><code>&#126;/.factory/skills</code></td><td><code>.factory/skills</code></td></tr>
<tr id="target-firebender"><td>firebender</td><td><code>&#126;/.firebender/skills</code></td><td><code>.firebender/skills</code></td></tr>
<tr id="target-forgecode"><td>forgecode</td><td><code>&#126;/forge/skills</code></td><td><code>.forge/skills</code></td></tr>
<tr id="target-fx"><td>fx</td><td><code>&#126;/.fx/skills</code></td><td><code>.fx/skills</code></td></tr>

<tr id="target-gemini"><td>gemini</td><td><code>&#126;/.gemini/skills</code></td><td><code>.gemini/skills</code></td></tr>
<tr id="target-goose"><td>goose</td><td><code>&#126;/.agents/skills</code></td><td><code>.agents/skills</code></td></tr>
<tr id="target-grok"><td>grok</td><td><code>&#126;/.grok/skills</code></td><td><code>.grok/skills</code></td></tr>
<tr id="target-hermes"><td>hermes</td><td><code>&#126;/.hermes/skills</code></td><td><code>.hermes/skills</code></td></tr>
<tr id="target-iflow"><td>iflow</td><td><code>&#126;/.iflow/skills</code></td><td><code>.iflow/skills</code></td></tr>
<tr id="target-jazz"><td>jazz</td><td><code>&#126;/.jazz/skills</code></td><td><code>skills</code></td></tr>
<tr id="target-junie"><td>junie</td><td><code>&#126;/.junie/skills</code></td><td><code>.junie/skills</code></td></tr>
<tr id="target-kilocode"><td>kilocode</td><td><code>&#126;/.kilo/skills</code></td><td><code>.kilo/skills</code></td></tr>
<tr id="target-kimchi"><td>kimchi</td><td><code>&#126;/.config/kimchi/harness/skills</code></td><td><code>.kimchi/skills</code></td></tr>
<tr id="target-kimi"><td>kimi</td><td><code>&#126;/.config/agents/skills</code></td><td><code>.agents/skills</code></td></tr>
<tr id="target-kimi-code"><td>kimi-code</td><td><code>&#126;/.kimi-code/skills</code></td><td><code>.kimi-code/skills</code></td></tr>
<tr id="target-kiro"><td>kiro</td><td><code>&#126;/.kiro/skills</code></td><td><code>.kiro/skills</code></td></tr>
<tr id="target-kode"><td>kode</td><td><code>&#126;/.kode/skills</code></td><td><code>.kode/skills</code></td></tr>
<tr id="target-letta"><td>letta</td><td><code>&#126;/.letta/skills</code></td><td><code>.skills</code></td></tr>
<tr id="target-lingma"><td>lingma</td><td><code>&#126;/.lingma/skills</code></td><td><code>.lingma/skills</code></td></tr>
<tr id="target-mcpjam"><td>mcpjam</td><td><code>&#126;/.mcpjam/skills</code></td><td><code>.mcpjam/skills</code></td></tr>
<tr id="target-mux"><td>mux</td><td><code>&#126;/.mux/skills</code></td><td><code>.mux/skills</code></td></tr>
<tr id="target-neovate"><td>neovate</td><td><code>&#126;/.neovate/skills</code></td><td><code>.neovate/skills</code></td></tr>
<tr id="target-omp"><td>omp</td><td><code>&#126;/.omp/agent/skills</code></td><td><code>.omp/skills</code></td></tr>
<tr id="target-ona"><td>ona</td><td>—</td><td><code>.ona/skills</code></td></tr>
<tr id="target-openclaw"><td>openclaw</td><td><code>&#126;/.openclaw/skills</code></td><td><code>skills</code></td></tr>
<tr id="target-opencode"><td>opencode</td><td><code>&#126;/.config/opencode/skills</code></td><td><code>.opencode/skills</code></td></tr>
<tr id="target-openhands"><td>openhands</td><td><code>&#126;/.agents/skills</code></td><td><code>.agents/skills</code></td></tr>
<tr id="target-pi"><td>pi</td><td><code>&#126;/.pi/agent/skills</code></td><td><code>.pi/skills</code></td></tr>
<tr id="target-pochi"><td>pochi</td><td><code>&#126;/.pochi/skills</code></td><td><code>.pochi/skills</code></td></tr>
<tr id="target-posit-assistant"><td>posit-assistant</td><td><code>&#126;/.posit/assistant/skills</code></td><td><code>.posit/assistant/skills</code></td></tr>
<tr id="target-purecode"><td>purecode</td><td><code>&#126;/.purecode/skills</code></td><td><code>.agents/skills</code></td></tr>
<tr id="target-qoder"><td>qoder</td><td><code>&#126;/.qoder/skills</code></td><td><code>.qoder/skills</code></td></tr>
<tr id="target-qoder-cn"><td>qoder-cn</td><td><code>&#126;/.qoder-cn/skills</code></td><td><code>.qoder/skills</code></td></tr>
<tr id="target-qwen"><td>qwen</td><td><code>&#126;/.qwen/skills</code></td><td><code>.qwen/skills</code></td></tr>
<tr id="target-replit"><td>replit</td><td>—</td><td><code>.agents/skills</code></td></tr>
<tr id="target-reasonix"><td>reasonix</td><td><code>&#126;/.reasonix/skills</code></td><td><code>.reasonix/skills</code></td></tr>
<tr id="target-roo"><td>roo</td><td><code>&#126;/.roo/skills</code></td><td><code>.roo/skills</code></td></tr>
<tr id="target-rovodev"><td>rovodev</td><td><code>&#126;/.rovodev/skills</code></td><td><code>.rovodev/skills</code></td></tr>
<tr id="target-tabnine"><td>tabnine</td><td><code>&#126;/.tabnine/agent/skills</code></td><td><code>.tabnine/agent/skills</code></td></tr>
<tr id="target-trae"><td>trae</td><td><code>&#126;/.trae/skills</code></td><td><code>.trae/skills</code></td></tr>
<tr id="target-trae-cn"><td>trae-cn</td><td><code>&#126;/.trae-cn/skills</code></td><td><code>.trae/skills</code></td></tr>
<tr id="target-vibe"><td>vibe</td><td><code>&#126;/.vibe/skills</code></td><td><code>.vibe/skills</code></td></tr>
<tr id="target-verdent"><td>verdent</td><td><code>&#126;/.verdent/skills</code></td><td><code>.verdent/skills</code></td></tr>
<tr id="target-warp"><td>warp</td><td><code>&#126;/.agents/skills</code></td><td><code>.agents/skills</code></td></tr>
<tr id="target-windsurf"><td>windsurf</td><td><code>&#126;/.codeium/windsurf/skills</code></td><td><code>.windsurf/skills</code></td></tr>
<tr id="target-witsy"><td>witsy</td><td><code>&#126;/.agents/skills</code></td><td><code>.agents/skills</code></td></tr>
<tr id="target-xcode-claude"><td>xcode-claude</td><td><code>&#126;/Library/Developer/Xcode/CodingAssistant/ClaudeAgentConfig/skills</code></td><td><code>.claude/skills</code></td></tr>
<tr id="target-xcode-codex"><td>xcode-codex</td><td><code>&#126;/Library/Developer/Xcode/CodingAssistant/codex/skills</code></td><td><code>.codex/skills</code></td></tr>
<tr id="target-zcode"><td>zcode</td><td><code>&#126;/.zcode/skills</code></td><td><code>.zcode/skills</code></td></tr>
<tr id="target-zed"><td>zed</td><td><code>&#126;/.agents/skills</code></td><td><code>.agents/skills</code></td></tr>
<tr id="target-zencoder"><td>zencoder</td><td><code>&#126;/.agents/skills</code></td><td><code>.agents/skills</code></td></tr>
</tbody>
</table>

:::info Universal target
**universal** target（`&#126;/.agents/skills`）是一个共享的 Agent 目录，多个 AI CLI 都可以从中读取。当检测到任何其他 Agent 存在时，`skillshare init` 会自动检测到它。在 Project mode 下，`amp`、`codex`、`cursor`、`dexto`、`kimi`、`purecode`、`replit`、`warp`、`witsy`、`zed` 和 `zencoder` 共享同一个 `.agents/skills` 路径，并会被自动归类到 `universal` 下。

这与 [npx skills CLI](https://github.com/vercel-labs/skills) 使用的路径相同。共存相关细节参见 [FAQ: Using universal alongside npx skills](/docs/troubleshooting/faq#using-universal-alongside-npx-skills)。
:::

## 别名

某些 Target 有替代名称，用于向后兼容或方便使用：

| 别名 | 解析为 | 说明 |
|-------|-------------|-------|
| `agents` | `universal` | 旧版名称 |
| `claude-code` | `claude` | 旧版名称 |
| `aider-desk` | `aiderdesk` | 带连字符的变体 |
| `codearts-agent` | `codearts` | 带 agent 后缀 |
| `code-studio` | `codestudio` | 带连字符的变体 |
| `command-code` | `commandcode` | 带连字符的变体 |
| `devin-terminal` | `devin` | 带 terminal 后缀 |
| `deep-agents` | `deepagents` | 带连字符的变体 |
| `factory` | `droid` | 品牌 / 配置目录名 |
| `forge-code` | `forgecode` | 带连字符的变体 |
| `gemini-cli` | `gemini` | 带 CLI 后缀 |
| `github-copilot` | `copilot` | 完整产品名称 |
| `iflow-cli` | `iflow` | 带 CLI 后缀 |
| `kilo` | `kilocode` | 简称 |
| `kimi-cli` | `kimi` | 带 CLI 后缀 |
| `kiro-cli` | `kiro` | 带 CLI 后缀 |
| `mistral-vibe` | `vibe` | 完整产品名称 |
| `oh-my-pi` | `omp` | 完整产品名称 |
| `purecode-ai` | `purecode` | 带连字符的变体 |
| `qwen-code` | `qwen` | 带 code 后缀 |
| `rovo-dev` | `rovodev` | 带连字符的变体 |
| `tabnine-cli` | `tabnine` | 带 CLI 后缀 |
| `zenflow` | `zencoder` | 完整产品名称 |

在所有命令中，你可以使用别名或规范名称：

```bash
skillshare target add claude           # 规范名称
skillshare target add claude-code      # 别名——效果相同
```

别名会被自动解析。配置文件和状态输出中使用的是规范名称。

---

## 检查 Target 路径

对于任何 Target，运行：

```bash
skillshare target <name>
```

---

## 自定义 Target

没有找到你的 AI CLI？手动添加它：

```bash
skillshare target add myapp ~/.myapp/skills
```

详见 [Adding Custom Targets](./adding-custom-targets.md)。

---

## 相关内容

- [Adding Custom Targets](./adding-custom-targets.md) — 添加不支持的工具
- [Configuration](./configuration.md) — 配置文件参考
- [Commands: target](/docs/reference/commands/target) — Target 命令
