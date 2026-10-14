# statusdrawer 開發者備忘

開發 statusdrawer 時要提醒自己、以及與 AI 協作時記下的決策。使用者要知道的在 README，這裡只放開發者需要的。

---

## 運作方式

statusdrawer 是一個 Claude Code mod（function hooks 外掛）：`hooks/register.tsx` 在 `session.start` 開一個 pane，`ui.render` 依 `$.state` 裡的值畫它。每個值由下表的來源更新，寫入 `$.state` 後引擎只重畫讀到它的地方。

| 列 | 來源 | 更新時機 |
|---|---|---|
| Context、Session、Weekly、Cost | `$.session.usage()` | `session.measure`（每輪結束、rate limit 變動） |
| Model | `$.session.model()` | session 開始、每輪結束 |
| Thinking | `turn.step` 的 `e.effort`；session 剛開始還沒送過請求時，先用 settings 的 `modelSettings[model].effortLevel` 或 `effortLevel` | 每次主對話送出請求 |
| Mode | 舊式 hook 事件（`classic.SessionStart`、`classic.UserPromptSubmit`、`classic.Stop`）帶的 `permission_mode` | 上述三個事件 |
| Editor | settings 的 `editorMode`；vim 即時狀態讀 `~/.claude/statusdrawer/vim/<session id>` | session 開始；即時狀態每 0.5 秒 |
| Uptime | `usage().startedAt` 與 `$.session.turns()` | 每 30 秒、每輪結束 |
| Cache | `turn.step` 結果的 `usage`：`cache_read / (input + cache_read + cache_creation)` | 每次主對話送出請求 |
| Compact | `session.compact` 成功（非 `precompute`、非 skip、非子 agent）時加一 | 每次壓縮 |
| Agents | `$.agent.list()` 裡 `pending`／`running`／`waiting` 的數量 | 每 5 秒、`agent.spawn`、每輪結束 |
| Git、Path | `git rev-parse --abbrev-ref HEAD`、`git diff HEAD --shortstat`；`$.session.cwd()` | 每 5 秒 |
| Version | `$.session.version()`、npm registry 的 `@anthropic-ai/claude-code/latest` | session 開始、每 6 小時 |

### 狀態列轉接（`statusline/statusline.sh`）

mod 的 API 沒有 vim 的即時狀態，但 Claude Code 每次重跑狀態列都會從 stdin 給一包 JSON，裡面有 `vim.mode` 和 `session_id`，而且 vim 模式切換時會立刻重跑。轉接程式只把 `vim.mode` 寫進 `~/.claude/statusdrawer/vim/<session_id>`（先寫暫存檔再 `mv`，mod 不會讀到寫一半的內容），不輸出任何東西。

mod 在 `session.start` 把外掛附帶的那份複製到 `~/.claude/statusdrawer/statusline.sh`，內容不同才覆寫。使用者的 `statusLine` 設定指向這個固定路徑，外掛更新、快取路徑變動都不受影響。用 `sh <path>` 執行，所以不需要執行權限。

權限模式不在那包 JSON 裡（Claude Code 在權限模式改變時會重跑狀態列，但沒有告訴它改成什麼），所以 Mode 只能走舊式 hook 事件。

## 設計決定

- **指令叫 `/sdrawer`，別名交給使用者（2026-10-14）。** 要開源，所以選不容易跟內建指令或其他外掛撞名、又看得出是 statusdrawer 的名字。嫌長的人在 `/config` 的 Command alias 自己設，例如 `sdr`。`command.run` 因此是不帶 matcher 的 hook，每個斜線指令都會經過，所以加了 `.catch`：出錯時不是自己的指令就照常放行。
- **不畫底色（2026-10-14）。** 停靠的 pane 外框（上方 × 那列、左邊分隔欄、下方）由引擎的容器畫，底色是主題的 `composerSidebarBackground`（`dark-ansi` 是 `ansi:blackBright`，`dark` 是 `rgb(38, 38, 38)`）。mod 只能畫 body，蓋不到外框；只蓋 body 會變成「外框一個色、內容一個色」。所以底色交給主題，README 教使用者自訂主題。這個 token 不在官方文件的清單裡，是從 Claude Code 程式找到的，改名時主題會忽略它，退回灰底。
- **不用 Fragment。** 這個 surface 把 `<>…</>` 當直排容器排版，放在 row 裡的「·」和 effort 會被疊成上下兩行。同一列的元素都直接放進 `flexDirection="row"` 的 Box。
- **bar 寬度要扣 paddingX。** 外層 Box 左右各 1 格 padding，`bodyColumns` 沒扣掉的話，bar 會超出寬度、折到下一行，看起來像兩條 bar。
- **key 欄寬 9。** 最長的 key 是 `Thinking`，8 個字加一格空白。
- **會擋住引擎動作的 hook 都有 `.catch`。** `agent.spawn`、`session.compact`、三個 `classic.*` 事件都是 gating hook；側邊欄出錯時 `.catch(($, e, next) => next(e))` 讓引擎照常進行（`next` 已呼叫過就沿用那次結果，不會重跑）。

## 已知的牆

- **權限模式有延遲。** 沒有「權限模式改變」的事件，`shift+tab` 之後要等下一個 prompt 或回覆結束。
- **tmux 256 色。** Claude Code 偵測到 `$TMUX` 就把顏色限制在 256 色，hex 色會被換成最接近的 256 色（`#11111b` 會變成第 17 號深藍）。要全彩得設 `CLAUDE_CODE_TMUX_TRUECOLOR=1`。
- **空的狀態列仍佔一行。** 轉接程式不輸出任何東西，Claude Code 還是留一行空白。
- **組織管理的環境。** `session.compact` 與 `classic.*` 事件在組織管理的 Claude Code 裡會略過使用者安裝的外掛，Mode 和 Compact 會是空的。

## 未做

- 配色可設定（目前寫死 Catppuccin Mocha）
- README 的截圖或 demo
- CHANGELOG

## 開發

### 載入

- 只跑一次：`claude --plugin-dir ~/Documents/sideproj/claude-code-statusdrawer`
- 長期：把這個資料夾加成本機市集再安裝；市集是資料夾、`source` 是相對路徑時，Claude Code 直接讀資料夾本身，改完在 session 裡跑 `/reload-plugins` 就生效，不用重新安裝：

  ```sh
  claude plugin marketplace add ~/Documents/sideproj/claude-code-statusdrawer
  claude plugin install statusdrawer@claude-code-statusdrawer
  ```

### 檢查

- `claude plugin validate .`：讀 manifest、市集檔和 hooks module，列出 hook、呼叫的 API 和引擎會拒絕的地方。
- `claude plugin test .`：跑 `tests/*.test.ts`。
- 型別檢查：Claude Code 載入外掛後會在 `.claude-plugin/types/` 放型別檔、在根目錄放 `tsconfig.json`（兩者都在 `.gitignore`），之後 `tsc -p .` 即可。
