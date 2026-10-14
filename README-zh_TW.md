# statusdrawer

[![License](https://img.shields.io/github/license/vulcanshen/claude-code-statusdrawer)](LICENSE)

[English](README.md)

**Claude Code 的側邊欄。** 用量上限、context、模型、花費、git 和 session 統計都放在對話旁邊的面板裡，不再擠在輸入框下方。

```
Model    claude-opus-5-5
Thinking xhigh
Mode     auto
Editor   vim · INSERT

Context  ████████░░░░░░  65%
Session  █░░░░░░░░░░░░░   6%
         resets 2h12m
Weekly   ███░░░░░░░░░░░  23%
         resets 3d17h
Cost     $35.46

Uptime   11h39m · 43 turns
Cache    100%
Compact  0
Agents   none

Git      main  +12 -3
Path     ~/projects/app

Version  2.1.299  ↑ 2.1.300
```

## 顯示的內容

| 列 | 內容 |
|---|---|
| Model、Thinking | 這個 session 的模型和 effort 等級 |
| Mode | 權限模式：`auto`、`plan`、`accept edits`… |
| Editor | `vim` 或 `normal`；開了[狀態列轉接](#即時-vim-狀態選用)時，vim 會加上即時狀態 |
| Context、Session、Weekly | context 用量，以及方案的 5 小時和每週用量與重置倒數。bar 在 50%、70%、90% 分別從綠變黃、橘、紅 |
| Cost | 這個 session 到目前為止的花費 |
| Uptime | session 開了多久、你送出了幾則訊息 |
| Cache | 上一次請求有多少是從 prompt cache 讀的 |
| Compact | 對話被壓縮過幾次 |
| Agents | 正在執行的子 agent |
| Git、Path | branch 和增減行數、工作目錄 |
| Version | 你的 Claude Code 版本，有新版時一併顯示 |

## 需求

- Claude Code 2.1.287 以上
- fullscreen 渲染模式：執行一次 `/tui fullscreen`。沒開的話，面板會放在輸入框上方，而不是對話旁邊
- 終端機要夠寬：自動開啟的面板會等終端機寬到大約 110～144 欄才出現
- `jq`，只有即時 vim 狀態需要

## 安裝

在 Claude Code 的輸入框輸入：

```
/plugin install statusdrawer --marketplace vulcanshen/claude-code-statusdrawer
```

詢問是否加入市集時回答 `y`，再選 user scope，每個 session 就都會有。

## 使用方式

- session 開始時會自動打開。
- `/sdrawer` 開關側邊欄，Claude 回覆到一半也能用。
- 點右上角的 `×` 關閉，或按 `Ctrl+X Tab` 移過去再按 `Ctrl+X X`。
- 焦點在側邊欄上時，`Ctrl+X ←`、`Ctrl+X →` 可以調寬或調窄。

### 比較短的指令

`/config` 裡有 statusdrawer 的 **Command alias** 欄位。填入例如 `sdr`，`/sdr` 就跟 `/sdrawer` 一樣能用。

### 即時 vim 狀態（選用）

外掛拿不到 vim 現在是 INSERT 還是 NORMAL，但狀態列程式拿得到。statusdrawer 附了一支小程式，它不畫任何東西，只負責把狀態轉交過來。在 `~/.claude/settings.json` 把狀態列指向它：

```json
"statusLine": {
  "type": "command",
  "command": "sh ~/.claude/statusdrawer/statusline.sh"
}
```

這會取代你現在的狀態列。它不輸出任何東西，但原本狀態列的位置會留下一行空白。

### 讓底色跟終端機一樣

側邊欄的底色由 Claude Code 的主題決定，不是外掛。要讓它跟終端機同色，在 `~/.claude/themes/` 建一個自訂主題，例如 `~/.claude/themes/my-theme.json`：

```json
{
  "name": "My theme",
  "base": "dark-ansi",
  "overrides": { "composerSidebarBackground": "#11111b" }
}
```

再用 `/theme` 選它。`base` 填你現在用的主題，值填你終端機的底色。

在 tmux 裡，Claude Code 會把顏色限制在 256 色，像 `#11111b` 這種顏色會被換成別的顏色。如果你的 tmux 會傳遞全彩，在 `~/.claude/settings.json` 的 `env` 加上 `"CLAUDE_CODE_TMUX_TRUECOLOR": "1"`，再重開 Claude Code。

## 資料存放位置

`~/.claude/statusdrawer/` 放狀態列轉接程式（`statusline.sh`，由外掛複製過去），以及每個 session 一個記錄 vim 狀態的小檔案（`vim/`）。超過一週沒更新的檔案會自動刪除。

側邊欄每 6 小時會向 npm registry 查一次 Claude Code 的最新版本。

## 限制

- 用 `shift+tab` 切換權限模式後，**Mode** 要等你送出下一則訊息或 Claude 回覆結束才會更新。
- 配色是 Catppuccin Mocha，目前不能更改。
- 由組織管理的 Claude Code 環境裡，**Mode** 和 **Compact** 可能會是空的：那些事件會略過使用者自己安裝的外掛。
- Claude Code 的 mod 介面還在搶先體驗階段，版本更新時可能會變動。

## 移除

```
claude plugin uninstall statusdrawer@claude-code-statusdrawer
```

再移除你加的 `statusLine` 設定，以及 `~/.claude/statusdrawer/` 資料夾。

## 更多

- [開發者備忘](docs/dev-remarks.md)
- [授權：GPL-3.0](LICENSE)
