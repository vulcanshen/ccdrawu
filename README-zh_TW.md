# ccdrawu

[![License](https://img.shields.io/github/license/vulcanshen/ccdrawu)](LICENSE)

[English](README.md)

**Claude Code 的側邊欄。** 用量上限、context、模型、花費、git 和 session 統計都放在對話旁邊的面板裡，不再擠在輸入框下方。

```
Model    claude-opus-5-5
Effort   xhigh
Mode     auto
Context  ▰▰▰▰▰▰▰▰▱▱▱▱▱▱  65%
Session  ▰▱▱▱▱▱▱▱▱▱▱▱▱▱   6%
         resets 2h12m
Weekly   ▰▰▰▱▱▱▱▱▱▱▱▱▱▱  23%
         resets 3d17h
Cost     $35.46
Uptime   11h39m · 43 turns
Cache    100%
Compact  0
Tasks    none
Git      main  +12 -3
Path     ~/projects/app
Claude   2.1.299  ↑ 2.1.300
ccdrawu  0.1.0
```

## 顯示的內容

| 列 | 內容 |
|---|---|
| Model、Effort | 這個 session 的模型，以及上一次請求的 effort 等級：`low` 綠色，一路到 `max` 紅色 |
| Mode | 權限模式：`auto`、`plan`、`accept edits`… |
| Context、Session、Weekly | context 用量，以及方案的 5 小時和每週用量與重置倒數。bar 在 50%、70%、90% 分別從綠變黃、橘、紅 |
| Cost | 這個 session 到目前為止的花費 |
| Uptime | session 開了多久、你送出了幾則訊息 |
| Cache | 上一次請求有多少是從 prompt cache 讀的 |
| Compact | 對話被壓縮過幾次 |
| Tasks | 正在跑的 subagent，以及 Claude 上一次回覆結束時還在背景跑的 shell、monitor、workflow，各種類分別計數：`2 agents｜1 shell` |
| Git、Path | branch 和增減行數、工作目錄 |
| Claude、ccdrawu | 你的 Claude Code 版本（有新版時一併顯示），以及這個側邊欄自己的版本 |

下面這些預設是關的，到 `ccdrawu` 裡打開：

| 列 | 內容 |
|---|---|
| Style | `/config` 裡設的 output style |
| ID、Account | 這個 session 的 id 開頭，以及登入的 Claude 帳號 |
| Window | context 用了多少 token、視窗多大：`420k / 1M` |
| Rate | 這個 session 到目前為止，平均每小時花多少 |
| Tokens、Speed | 側邊欄載入後，對話累計的 input、output、cache token，以及上一次回覆每秒輸出幾個 token |
| Skills | 這個 session 用過幾次 skill，以及最後一個 |
| Sync、Files、Commit、Repo | 跟 upstream 差幾個 commit、staged／changed／new 的檔案數、HEAD 的短 hash，以及 origin 的 owner/repo（在 worktree 裡時加上 worktree 名稱）；沒有 upstream、commit 或 origin 時顯示 `no upstream`、`no commits`、`no origin`，不在 repo 裡時顯示 `not a repo` |
| Status | [status.claude.com](https://status.claude.com) 的服務狀態：正常是綠色，有事故是橘色或紅色 |

## 需求

- Claude Code 2.1.287 以上
- fullscreen 渲染模式：執行一次 `/tui fullscreen`。沒開的話，面板會放在輸入框上方，而不是對話旁邊
- 終端機要夠寬：自動開啟的面板會等終端機寬到大約 110～144 欄才出現
- `ccdrawu` 指令，用來挑側邊欄顯示的項目；它需要終端機用 [Nerd Font](https://www.nerdfonts.com)
- macOS、Linux 或 Windows

## 安裝

先安裝 `ccdrawu` 指令：

```sh
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/vulcanshen/ccdrawu/main/install.sh | sh

# Homebrew（macOS / Linux）
brew install vulcanshen/tap/ccdrawu
```

```powershell
# Windows PowerShell
irm https://raw.githubusercontent.com/vulcanshen/ccdrawu/main/install.ps1 | iex

# Scoop（Windows）
scoop bucket add vulcanshen https://github.com/vulcanshen/scoop-bucket
scoop install ccdrawu
```

再把側邊欄加進 Claude Code：

```sh
ccdrawu install claude
```

它會幫你的使用者裝好 ccdrawu 外掛，之後每個 session 都有側邊欄；已經開著的 session 跑一次 `/reload-plugins`。接著執行 `ccdrawu` 挑要顯示的項目。

## 使用方式

- session 開始時會自動打開。
- `/ccdrawu` 開關側邊欄，Claude 回覆到一半也能用。
- 點右上角的 `×` 關閉，或按 `Ctrl+X Tab` 移過去再按 `Ctrl+X X`。
- 焦點在側邊欄上時，`Ctrl+X ←`、`Ctrl+X →` 可以調寬或調窄。

### 比較短的指令

執行 `ccdrawu`，移到 Settings 底下的 **preferences** 按 `Enter`，在 **Alias** 上再按 `Enter`，輸入例如 `ccd`。一秒內，每個開著的 session 裡 `/ccd` 就跟 `/ccdrawu` 一樣能用。改掉或拿掉的名字，在 session 重開之前都還能用。

別名存在 `~/.config/ccdrawu/config.json` 的 `"alias": "ccd"`，也可以直接改那裡。

### 讓底色跟終端機一樣

側邊欄的底色由 Claude Code 的主題決定，不是外掛。要讓它跟終端機同色，執行 `ccdrawu`，在 Settings 底下的 **preferences** 設定：

- **Claude theme**：選一個 Claude Code 內建主題當底（`dark-ansi`、`dark`…）。
- **Drawer background**：選你終端機的底色；在它上面按 `x` 會改回主題原本的底色。

ccdrawu 會寫一份自己的主題，就是那個內建主題加上你選的底色（`~/.claude/themes/ccdrawu.json`），並在 Claude Code 裡選用它。把 Claude theme 設回 `off`，或執行 `ccdrawu uninstall claude`，會換回你原本的主題。開著的 Claude Code 會立刻跟著換。

在 tmux 裡，Claude Code 會把顏色限制在 256 色，像 `#11111b` 這種顏色會被換成別的顏色。如果你的 tmux 會傳遞全彩，在 `~/.claude/settings.json` 的 `env` 加上 `"CLAUDE_CODE_TMUX_TRUECOLOR": "1"`，再重開 Claude Code。

## 資料存放位置

- `~/.config/ccdrawu/config.json`：你在 `ccdrawu` 裡挑的項目和別名。`CCDRAWU__CONFIG` 或 `XDG_CONFIG_HOME` 可以改放別的資料夾。
- `~/.ccdrawu/`：外掛寫的東西，`ccdrawu` 列出的項目清單（`catalog.json`）。`CCDRAWU__DATA` 可以改放別的資料夾。

在 Windows 上，`~` 就是你的使用者資料夾（`%USERPROFILE%`）。

側邊欄每 6 小時會向 npm registry 查一次 Claude Code 的最新版本。

## 限制

- 用 `shift+tab` 切換權限模式後，**Mode** 要等 Claude 跑完目前這一步的工具才會更新；Claude 閒著的時候，要等你送出下一則訊息。
- 配色是 Catppuccin Mocha，目前不能更改。
- 由組織管理的 Claude Code 環境裡，**Mode**、**Compact** 和 **Skills** 可能會是空的，壓縮對話後 **Context** 也要等下一次回覆才會更新：那些事件會略過使用者自己安裝的外掛。
- Claude Code 的 mod 介面還在搶先體驗階段，版本更新時可能會變動。

## 移除

先把側邊欄從 Claude Code 拿掉：

```sh
ccdrawu uninstall claude
```

再移除 `ccdrawu` 指令：

```sh
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/vulcanshen/ccdrawu/main/uninstall.sh | sh
```

```powershell
# Windows PowerShell
irm https://raw.githubusercontent.com/vulcanshen/ccdrawu/main/uninstall.ps1 | iex
```

兩者都會先問過你，才刪 `~/.config/ccdrawu/` 和 `~/.ccdrawu/`。用 Homebrew 或 Scoop 裝的，用 `brew uninstall ccdrawu` 或 `scoop uninstall ccdrawu`。

## 更多

- [開發者備忘](docs/dev-remarks.md)
- [授權：GPL-3.0](LICENSE)
