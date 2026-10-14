# ccdrawu 開發者備忘

開發 ccdrawu 時要提醒自己、以及與 AI 協作時記下的決策。使用者要知道的在 README，這裡只放開發者需要的。

---

## 運作方式

ccdrawu 是一個 Claude Code mod（function hooks 外掛）：`hooks/register.tsx` 在 `session.start` 開一個 pane，`ui.render` 依 `$.state` 裡的值畫它。每個值由下表的來源更新，寫入 `$.state` 後引擎只重畫讀到它的地方。

| 列 | 來源 | 更新時機 |
|---|---|---|
| Context、Session、Weekly、Cost | `$.session.usage()` | `session.measure`（每輪結束、rate limit 變動）；compact 完成後（`classic.PostCompact`，`/compact` 不算一輪，沒有 measure） |
| Model | `$.session.model()` | session 開始、每輪結束 |
| Effort | `turn.step` 的 `e.effort`，依等級上色（`effortColor`：low 綠到 max 紅）；session 剛開始還沒送過請求時，先用 settings 的 `modelSettings[model].effortLevel` 或 `effortLevel` | 每次主對話送出請求 |
| Mode | 舊式 hook 事件（`classic.SessionStart`、`classic.UserPromptSubmit`、`classic.PostToolBatch`、`classic.Stop`）帶的 `permission_mode`，只認主對話 | 上述四個事件；PostToolBatch 是每一步的工具跑完後一次（`classic.PreToolUse` 在 mod 裡只有工具的內容、沒有 permission_mode，PostToolUse 每個工具一次、還帶整份輸出） |
| Uptime | `usage().startedAt` 與 `$.session.turns()` | 每 30 秒、每輪結束 |
| Cache | `turn.step` 結果的 `usage`：`cache_read / (input + cache_read + cache_creation)` | 每次主對話送出請求 |
| Compact | `session.compact` 成功（非 `precompute`、非 skip、非子 agent）時加一 | 每次壓縮 |
| Tasks | subagent：`$.agent.list()` 裡 `pending`／`running`／`waiting` 的數量（只算這個 session 的；workflow 的 agent 不在裡面）。其他：`classic.Stop` 的 `background_tasks` 按種類計數，不算 subagent——等自己背景工作的 subagent 不在這份清單裡，它開的 Monitor 算成 shell（2026-10-14 實測）。顯示用 `｜` 分隔（user 定的），它佔兩格，寬度用 `cellWidth` 算 | subagent 每 5 秒、`agent.spawn`、每輪結束；其他在每次回覆結束 |
| Git、Path | `git rev-parse --abbrev-ref HEAD`、`git diff HEAD --shortstat`；`$.session.cwd()` | 每 5 秒 |
| Claude | `$.session.version()`、npm registry 的 `@anthropic-ai/claude-code/latest` | session 開始、每 6 小時 |
| ccdrawu | 外掛自己的 `.claude-plugin/plugin.json`（`$.plugin` 只有 name 和 root，沒有版本） | session 開始 |

## 設計決定

- **改名 ccdrawu，加入 terminu 家族（2026-10-14）。** 原名 statusdrawer、指令 `/sdrawer`。設定介面改用 terminu-go 做成獨立執行檔，由它安裝、設定 mod，入口本身就是 terminu app，所以照家族慣例取 `*u` 的名字。`cc` 點出是給 Claude Code 用的，又不把「Claude」放進產品名；repo 不加 `claude-code-` 前綴，跟家族其他 repo 一致。執行檔、外掛、斜線指令都叫 ccdrawu。
- **指令叫 `/ccdrawu`，別名交給使用者（2026-10-14）。** 要開源，所以選不容易跟內建指令或其他外掛撞名的名字。嫌長的人自己設別名，例如 `ccd`。`command.run` 因此是不帶 matcher 的 hook，每個斜線指令都會經過，所以加了 `.catch`：出錯時不是自己的指令就照常放行。
- **別名存在 config.json 的 `alias`，不用 `userConfig`（2026-10-14）。** 跟項目開關放同一個檔、由 TUI 設定，開著的 session 1 秒內讀到新名字就 `$.command.register`（`hooks/commands.ts`）。mod API 沒有移除指令的方法，所以一個名字註冊過就留到 session 結束：改掉或清空的別名在那之前照樣能用、也還列在選單裡。否決：TUI 改 Claude Code 的 `settings.json` 的 `pluginConfigs`（跟 `/config` 那一列同步）——要動別人的設定檔，開著的 session 也不確定會不會跟著換。
- **設定 TUI 是側欄＋明細（2026-10-14）。** `[1] PREFERENCES` 分組：DRAWER（sidebar、inline）、SETTINGS（alias）；`[2]` 是游標那一項，即時換內容與 label（tdp v0.6.2 `components/layout/panel` 的「分組」「明細」，terminu-go v0.1.2 的 `Row.Group`、`List.OnMove`）。側欄寬 24，跟 locku、webu 一樣。在 `[1]` 按 `Enter` 進 `[2]`（tdp v0.6.3 給 K3 開的明細例外，terminu-go v0.1.3 的 `Panel.Details`、`Action.ToDetails`），`Tab`、`2` 也行；三項都有明細，所以每一項都進得去。別名的檢查（`config.CleanAlias`）照 mod 的 `commandNames`，mod 不收的名字在輸入框就擋下。
- **也給 Windows 用，建置照 kbu（2026-10-14）。** Claude Code 三個平台都有，家族其他 app 是 unix-first，只有 kbu 出 Windows 版。goreleaser 多出 windows 的 amd64、arm64（zip）、發到 Scoop（`vulcanshen/scoop-bucket`），`install.ps1`／`uninstall.ps1` 照 kbu（裝到 `%LOCALAPPDATA%\ccdrawu`、加進使用者的 PATH）；release 的測試也在 windows-latest 跑（那邊沒有 cgo，不加 `-race`）。位置照 E3：Windows 上 `~` 就是 `%USERPROFILE%`，不用 `%APPDATA%`（tdp v0.6.4，ccdrawu 提的 CD4，user 選的）。mod 先讀 `USERPROFILE` 再讀 `HOME`，跟 Go 的 `os.UserHomeDir` 同一個順序，兩邊才會算出同一個資料夾。
- **不顯示 vim 的 INSERT／NORMAL（2026-10-14）。** mod API 拿不到 vim 的即時狀態，唯一的來源是 Claude Code 跑 `statusLine` 指令時給的 JSON：先做過外掛附的 sh 轉接（要 sh 與 jq），再改成 `ccdrawu bridge` 子指令（Windows 也能用）。但 `statusLine` 只能填一個指令，佔用它就會跟使用者自己的狀態列（例如 ccstatusline）互相蓋掉；而 Claude Code 輸入框下方本來就顯示 vim 模式。所以整列（原本叫 Editor，後來叫 Vim）連同轉接一起拿掉，ccdrawu 不碰 `statusLine`。
- **不畫底色（2026-10-14）。** 停靠的 pane 外框（上方 × 那列、左邊分隔欄、下方）由引擎的容器畫，底色是主題的 `composerSidebarBackground`（`dark-ansi` 是 `ansi:blackBright`，`dark` 是 `rgb(38, 38, 38)`）。mod 只能畫 body，蓋不到外框；只蓋 body 會變成「外框一個色、內容一個色」。所以底色交給主題，README 教使用者自訂主題。這個 token 不在官方文件的清單裡，是從 Claude Code 程式找到的，改名時主題會忽略它，退回灰底。
- **不用 Fragment。** 這個 surface 把 `<>…</>` 當直排容器排版，放在 row 裡的「·」和 effort 會被疊成上下兩行。同一列的元素都直接放進 `flexDirection="row"` 的 Box。
- **bar 寬度要扣 paddingX。** 外層 Box 左右各 1 格 padding，`bodyColumns` 沒扣掉的話，bar 會超出寬度、折到下一行，看起來像兩條 bar。
- **key 欄寬 9。** key 最多 8 個字（測試守著），加一格空白。
- **會擋住引擎動作的 hook 都有 `.catch`。** `agent.spawn`、`session.compact`、`classic.*` 事件都是 gating hook；側邊欄出錯時 `.catch(($, e, next) => next(e))` 讓引擎照常進行（`next` 已呼叫過就沿用那次結果，不會重跑）。

## 已知的牆

- **權限模式有延遲。** 沒有「權限模式改變」的事件。Claude 工作中按 `shift+tab`，等目前這一步的工具跑完才更新；閒置時要等下一個 prompt。mod API 沒有讀目前模式的方法；`statusLine` 指令收到的 JSON 也沒有：2026-10-14 在 Claude Code 2.1.299 把它錄下來看過，按 `shift+tab` 時狀態列馬上重跑，但前後只差 `cost.total_duration_ms`，沒切換時它也會這樣重跑，連「重跑了」都不能當成模式變了的訊號。
- **tmux 256 色。** Claude Code 偵測到 `$TMUX` 就把顏色限制在 256 色，hex 色會被換成最接近的 256 色（`#11111b` 會變成第 17 號深藍）。要全彩得設 `CLAUDE_CODE_TMUX_TRUECOLOR=1`。
- **組織管理的環境。** `session.compact` 與 `classic.*` 事件在組織管理的 Claude Code 裡會略過使用者安裝的外掛，Mode 和 Compact 會是空的，compact 後的 Context 也要等下一輪結束才更新。

## 未做

- 配色可設定（目前寫死 Catppuccin Mocha）
- README 的截圖或 demo
- CHANGELOG：release workflow 從它取 release notes，第一次發布前要有
- `ccdrawu install`／`uninstall`：在那之前，安裝腳本、Homebrew caveats、README 都教使用者在 Claude Code 裡 `/plugin install`
- 沒在真的 Windows 上跑過：mod 的 `$.fs` 收不收 `C:\Users\me/.ccdrawu/...` 這種混用分隔的路徑、Windows 上 rename 蓋過 mod 正在讀的檔會不會失敗（TUI 存設定靠 rename）、`.ps1` 的語法（這台 Mac 沒有 PowerShell）

## 開發

### 載入

- 只跑一次：`claude --plugin-dir ~/Documents/sideproj/ccdrawu`
- 長期：把這個資料夾加成本機市集再安裝；市集是資料夾、`source` 是相對路徑時，Claude Code 直接讀資料夾本身，改完在 session 裡跑 `/reload-plugins` 就生效，不用重新安裝：

  ```sh
  claude plugin marketplace add ~/Documents/sideproj/ccdrawu
  claude plugin install ccdrawu@ccdrawu
  ```

### 設定程式（Go）

- 跟家族其他 app 同一套建置：`cmd/ccdrawu`（命令列）、`internal/config`（E3 位置、catalog 與設定檔）、`internal/ui`（TUI 與 splash）、`internal/version`（release 時由 goreleaser 的 ldflags 注入）。`make` 列出所有指令；`make run` 跑 TUI，`make build` 編出 `./ccdrawu`。
- Windows 版編不編得過：`GOOS=windows GOARCH=amd64 go build ./cmd/ccdrawu`（arm64 同），或 `make snapshot` 一次打六個平台的包到 `dist/`。`make release-check`（`goreleaser check`）要有 GitHub remote 才過。
- 本機同時改 terminu-go 時用 `go work init . ../terminu-go`（`go.work` 不進版控）。
- `internal/ui/icon.svg` 是 `docs/icon.svg` 的複本：splash 用 `go:embed` 讀 icon，讀不到上層目錄。換 icon 時兩個一起換，`TestSplashIconIsDocsIcon` 會抓不一致。

### 檢查

- `make check`：gofmt、`go vet`、`go test` 加上 `make mod-check`（下面兩個 plugin 指令），commit 前跑。有 `go.work` 時另外跑 `GOWORK=off make check`，確認對 go.mod 裡的 terminu-go 版本也過。
- `claude plugin validate .`：讀 manifest、市集檔和 hooks module，列出 hook、呼叫的 API 和引擎會拒絕的地方。
- `claude plugin test .`：跑 `tests/*.test.ts`。
- 型別檢查：Claude Code 載入外掛後會在 `.claude-plugin/types/` 放型別檔、在根目錄放 `tsconfig.json`（兩者都在 `.gitignore`），之後 `tsc -p .` 即可。
