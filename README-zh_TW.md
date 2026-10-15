# ccdrawu

<p align="center"><img src="docs/icon.svg" width="128" alt="ccdrawu icon" /></p>

[![GitHub Release](https://img.shields.io/github/v/release/vulcanshen/ccdrawu)](https://github.com/vulcanshen/ccdrawu/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/vulcanshen/ccdrawu)](https://go.dev/)
[![License](https://img.shields.io/badge/license-GPL--3.0-blue)](LICENSE)

**語言**：[English](README.md) · 繁體中文

**Claude Code 的側邊欄。** 用量上限、context、模型、花費、git 和 session 統計都放在對話旁邊的面板裡，不再擠在輸入框下方。

![demo](docs/demo.gif)

## 顯示的內容

| 列 | 內容 |
|---|---|
| Pet | 陪著你的像素寵物，見[寵物](#寵物) |
| Model、Effort | 這個 session 的模型，以及上一次請求的 effort 等級：`low` 綠色，一路到 `max` 紅色 |
| Mode | 權限模式：`auto`、`plan`、`accept edits`… |
| Context、Session、Weekly | context 用量，以及方案的 5 小時和每週用量與重置倒數。bar 在 50%、70%、90% 分別從綠變黃、橘、紅 |
| Spend | 透過有消費上限的 gateway 使用時：用了多少、什麼時候重置 |
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
- 終端機要用 [Nerd Font](https://www.nerdfonts.com)：側邊欄的 bar 和 `ccdrawu` 指令都用到它的圖示。沒有的話，在 Settings 底下的 **preferences** 把 **Bar** 設成 `line`：每個選項旁邊都畫著它的 bar，你的終端機畫得出哪一種，一看就知道
- `ccdrawu` 指令，用來挑側邊欄顯示的項目
- macOS、Linux 或 Windows（Windows 還沒在真的機器上測過，見[限制](#限制)）

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

也可以只在 Claude Code 裡裝外掛：`/plugin install ccdrawu --marketplace vulcanshen/ccdrawu`。不建議這樣裝：側邊欄只會顯示預設的項目，要挑項目、排順序、換寵物、bar 或主題，都得用 `ccdrawu` 指令，或自己改 `~/.config/ccdrawu/config.json`。

## 使用方式

- session 開始時會自動打開。
- `/ccdrawu` 開關側邊欄，Claude 回覆到一半也能用。
- 點右上角的 `×` 關閉，或按 `Ctrl+X Tab` 移過去再按 `Ctrl+X X`。
- 焦點在側邊欄上時，`Ctrl+X ←`、`Ctrl+X →` 可以調寬或調窄。

### 挑要顯示的項目

執行 `ccdrawu`，在 Drawer 底下選 **sidebar**（對話旁邊）或 **inline**（輸入框上方）按 `Enter`。右邊有兩列：**Items** 是勾了的項目，**Order** 是同樣這些，照側邊欄顯示的順序列出。在 **Items** 上按 `Enter`，打開所有項目的清單：`Space` 勾或取消勾、`Enter` 套用、`Esc` 不改，`/` 可以打字篩選。新勾的項目會排在清單裡它前一項的後面。要調整順序，在 **Order** 上按 `Enter`：`D` 拿起一項，`j`、`k` 移動，`Enter` 放下，再按一次 `Enter` 存起來（`Esc` 放回原樣）。一秒內，開著的 session 就會跟著換。

![settings](docs/settings.gif)

### 寵物

側邊欄最上面住著一隻像素寵物，沒選的話是貓。在輸入框上方時，它佔最左邊一欄，不顯示台詞和球。

- Claude 工作時它會到處走：停下來、跳一下、跑起來、揮揮爪。
- Claude 做完時它坐下來說一句話。你中斷 Claude 時它只會安靜坐下；回覆以錯誤結束，或工具執行失敗時，它會跳一下。壓縮對話後它也會說一句。
- 閒著 2 分鐘它會趴下，10 分鐘後睡著；中間偶爾會散步、轉頭看看或揮爪。
- `/ccdrawu pat` 摸摸它：它停下來，頭上冒出愛心。
- `/ccdrawu ball` 丟一顆球給它，它會跑過去玩。

台詞是固定的幾句，不花 token。要換動物（貓、狗、兔子、鴨子、企鵝），執行 `ccdrawu`，在 Settings 底下的 **preferences** 設定 **Pet**，一秒內開著的 session 就會跟著換。花色每個 session 隨機。不想要寵物的話，在項目清單裡取消勾選 **Pet**。

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

- `~/.config/ccdrawu/config.json`：你在 `ccdrawu` 裡挑的項目、別名、bar 和寵物。`CCDRAWU__CONFIG` 或 `XDG_CONFIG_HOME` 可以改放別的資料夾。
- `~/.ccdrawu/`：外掛寫的東西，`ccdrawu` 列出的項目清單（`catalog.json`）。`CCDRAWU__DATA` 可以改放別的資料夾。

在 Windows 上，`~` 就是你的使用者資料夾（`%USERPROFILE%`）。

側邊欄每 6 小時會向 npm registry 查一次 Claude Code 的最新版本。

## 限制

- 用 `shift+tab` 切換權限模式後，**Mode** 要等 Claude 跑完目前這一步的工具才會更新；Claude 閒著的時候，要等你送出下一則訊息。
- 配色是 Catppuccin Mocha，目前不能更改。
- 由組織管理的 Claude Code 環境裡，**Mode**、**Compact** 和 **Skills** 可能會是空的，壓縮對話後 **Context** 也要等下一次回覆才會更新，寵物也不會在工具失敗時跳、壓縮後說話：那些事件會略過使用者自己安裝的外掛。
- Claude Code 的 mod 介面還在搶先體驗階段，版本更新時可能會變動。
- Windows 版有建置、也能安裝，但還沒在真的 Windows 上測過，歡迎回報問題。

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

- [更新紀錄](CHANGELOG.md)
- [開發者備忘](docs/dev-remarks.md)
- [授權：GPL-3.0](LICENSE)
