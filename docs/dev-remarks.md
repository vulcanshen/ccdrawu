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

下面這些是 2026-10-14 從 ccstatusline 搬的（user 挑的），預設 dock、inline 都關，免得側邊欄一下子變長：

| 列 | 來源 | 更新時機 |
|---|---|---|
| Style | `$.config.list()` 裡 `outputStyle` 那一列 | session 開始、每輪結束 |
| ID、Account | `$.session.id()` 的前 8 字；`$.session.identity()` 的 `email` | session 開始 |
| Window | `usage().context` 的 `tokens`／`window` | 跟 Context 一樣 |
| Rate | 花費 ÷ 開了幾小時，滿一分鐘才算（太早一個請求就讓數字亂跳） | 跟 Cost、Uptime 一樣 |
| Tokens、Speed | `turn.step` 結果的 `usage` 累計（主對話；`SessionCost` 只有 usd，沒有累計 token，所以只能從 mod 載入後自己加，reload 或接續舊 session 會少算）；速度是 output tokens ÷ 從送出到收完的秒數（含等第一個字） | 每次主對話送出請求 |
| Skills | `skill.prompt` 的次數與最後的 `skill` | 每次跑 skill |
| Sync、Files、Commit | 一個 `git status --porcelain=v2 --branch`：`branch.ab`、XY 碼、`?`、`branch.oid` | 每 5 秒 |
| Repo | `git remote get-url origin` 解析成 owner/repo（`$.session.repo().name` 只有內部 repo 才有）；`--git-dir` 跟 `--git-common-dir` 不同就是 linked worktree，取 toplevel 的資料夾名；工作目錄沒變就不重查 | 工作目錄改變時 |
| Status | `https://status.claude.com/api/v2/status.json` 的 `status.indicator`／`description` | session 開始、每 5 分鐘 |

## 設計決定

- **改名 ccdrawu，加入 terminu 家族（2026-10-14）。** 原名 statusdrawer、指令 `/sdrawer`。設定介面改用 terminu-go 做成獨立執行檔，由它安裝、設定 mod，入口本身就是 terminu app，所以照家族慣例取 `*u` 的名字。`cc` 點出是給 Claude Code 用的，又不把「Claude」放進產品名；repo 不加 `claude-code-` 前綴，跟家族其他 repo 一致。執行檔、外掛、斜線指令都叫 ccdrawu。
- **指令叫 `/ccdrawu`，別名交給使用者（2026-10-14）。** 要開源，所以選不容易跟內建指令或其他外掛撞名的名字。嫌長的人自己設別名，例如 `ccd`。`command.run` 因此是不帶 matcher 的 hook，每個斜線指令都會經過，所以加了 `.catch`：出錯時不是自己的指令就照常放行。
- **別名存在 config.json 的 `alias`，不用 `userConfig`（2026-10-14）。** 跟項目開關放同一個檔、由 TUI 設定，開著的 session 1 秒內讀到新名字就 `$.command.register`（`hooks/commands.ts`）。mod API 沒有移除指令的方法，所以一個名字註冊過就留到 session 結束：改掉或清空的別名在那之前照樣能用、也還列在選單裡。否決：TUI 改 Claude Code 的 `settings.json` 的 `pluginConfigs`（跟 `/config` 那一列同步）——要動別人的設定檔，開著的 session 也不確定會不會跟著換。
- **設定 TUI 是側欄＋明細（2026-10-14）。** `[1] PREFERENCES` 分組：DRAWER（sidebar、inline）、SETTINGS（preferences：所有設定放在同一頁——Alias、Claude theme、Drawer background、Plugin，user 要的，跟 locku 的 Settings › preference 一樣）；`[2]` 是游標那一項，即時換內容與 label（tdp v0.6.2 `components/layout/panel` 的「分組」「明細」，terminu-go v0.1.2 的 `Row.Group`、`List.OnMove`）。側欄寬 24，跟 locku、webu 一樣。在 `[1]` 按 `Enter` 進 `[2]`（tdp v0.6.3 給 K3 開的明細例外，terminu-go v0.1.3 的 `Panel.Details`、`Action.ToDetails`），`Tab`、`2` 也行；每一項都有明細，所以都進得去。別名的檢查（`config.CleanAlias`）照 mod 的 `commandNames`，mod 不收的名字在輸入框就擋下。
- **sidebar、inline 的 `[2]` 每組一列，`Enter` 開一個全部項目的 checkbox popup（2026-10-14）。** 原本每一組是畫在 panel 上的 checkbox，照 tdp `components/input/checkbox` 在 panel 上要按 `Enter` 翻（`Space` 開 Space menu），user 覺得不好用；照規範「五個以內畫在原地、超過五個開 popup」改成混著用，同一頁又會有兩種操作。user 定的做法：`[2]` 每組一列、列出勾了哪些（用 drawer 上的 key，跟側邊欄看到的同一個字），panel operation `e` Edit 開一個 popup 列出全部項目（一定超過五個，照規範就是 popup），`Space` 勾、`Enter` 一次套用、`Esc` 不改。popup 每列是「組名、key、名稱」對齊成三欄（名稱跟 key 一樣就省掉），`/` 篩選打組名也篩得到。這兩頁的 Space menu 只有這一項。之後 user 要每一列的 `Enter` 也開同一個 popup：每組的值本來就只在這個 popup 裡改，跟 Alias 那種「值不畫在原地的欄位，`Enter` 開它的 input popup」是同一套（terminu-go 的 `Field.Edit`），不違反規範；`e` 留在 Space menu，但不寫進 hint，免得 hint 變成 `Enter:edit e:edit`。從哪一列打開，cursor 就停在那一組的第一項（`e` 從第一項開始）：terminu-go v0.1.5 的 `Checks.Cursor`，ccdrawu 提的 CD8，規範沒缺、只補 lib。
- **也給 Windows 用，建置照 kbu（2026-10-14）。** Claude Code 三個平台都有，家族其他 app 是 unix-first，只有 kbu 出 Windows 版。goreleaser 多出 windows 的 amd64、arm64（zip）、發到 Scoop（`vulcanshen/scoop-bucket`），`install.ps1`／`uninstall.ps1` 照 kbu（裝到 `%LOCALAPPDATA%\ccdrawu`、加進使用者的 PATH）；release 的測試也在 windows-latest 跑（那邊沒有 cgo，不加 `-race`）。位置照 E3：Windows 上 `~` 就是 `%USERPROFILE%`，不用 `%APPDATA%`（tdp v0.6.4，ccdrawu 提的 CD4，user 選的）。mod 先讀 `USERPROFILE` 再讀 `HOME`，跟 Go 的 `os.UserHomeDir` 同一個順序，兩邊才會算出同一個資料夾。
- **先裝 ccdrawu，再 `ccdrawu install claude`（2026-10-14）。** 入口是這支程式，外掛由它裝（user 定的）：`claude plugin install ccdrawu --marketplace <來源> --json`（user scope；市集還沒加會先加；已經裝了就是沒變化），來源預設 `vulcanshen/ccdrawu`，`CCDRAWU__MARKETPLACE` 改指別處（開發時指本機資料夾）。`uninstall claude` 先 `plugin uninstall <id>` 再 `marketplace remove <市集名>`（市集只有 ccdrawu），設定與資料留著、告訴使用者在哪。移除腳本刪執行檔前先跑一次 `uninstall claude`。不碰 `statusLine`。`claude` 這個目標先佔著位置，之後要裝到別的地方再加。
- **不顯示 vim 的 INSERT／NORMAL（2026-10-14）。** mod API 拿不到 vim 的即時狀態，唯一的來源是 Claude Code 跑 `statusLine` 指令時給的 JSON：先做過外掛附的 sh 轉接（要 sh 與 jq），再改成 `ccdrawu bridge` 子指令（Windows 也能用）。但 `statusLine` 只能填一個指令，佔用它就會跟使用者自己的狀態列（例如 ccstatusline）互相蓋掉；而 Claude Code 輸入框下方本來就顯示 vim 模式。所以整列（原本叫 Editor，後來叫 Vim）連同轉接一起拿掉，ccdrawu 不碰 `statusLine`。
- **不畫底色（2026-10-14）。** 停靠的 pane 外框（上方 × 那列、左邊分隔欄、下方）由引擎的容器畫，底色是主題的 `composerSidebarBackground`（`dark-ansi` 是 `ansi:blackBright`，`dark` 是 `rgb(38, 38, 38)`）。mod 只能畫 body，蓋不到外框；只蓋 body 會變成「外框一個色、內容一個色」。所以底色交給主題，由 TUI 的 theme／background 寫自訂主題（下一條）。這個 token 不在官方文件的清單裡，是從 Claude Code 程式找到的，改名時主題會忽略它，退回灰底。
- **主題由 TUI 寫進 Claude Code（2026-10-14）。** user 選的做法：選一個內建主題當底，加上側邊欄底色。`internal/claude` 寫 `<Claude 資料夾>/themes/ccdrawu.json`（`CLAUDE_CONFIG_DIR`，否則 `~/.claude`），把 `settings.json` 的 `theme` 換成 `custom:ccdrawu`，換之前的值記在 ccdrawu 的 config.json（`theme.before`），選 off 或 `uninstall claude` 時換回去；使用者之後在 `/theme` 自己換過的就不動。改 `settings.json` 只改最外層 `theme` 那一個值，其他位元組原樣（`json.Decoder` 的 offset，測試守著「換過去再換回來，檔案一模一樣」），用暫存檔 rename、保留權限。否決：交給 mod 用 `$.config.set` 換——2026-10-14 實測 theme 那一列是 `choice`，只收 `auto` 與 6 個內建主題，`custom:…` 會被拒絕。內建主題清單（`claude.Themes`）是從 Claude Code 2.1.299 的執行檔裡找出來的。2026-10-14 實測：TUI 改了 `settings.json` 的 `theme` 與主題檔，開著的 session 立刻換色，不用重開。
- **寵物 pet（2026-10-15）。** Claude Code 的 `/buddy` 已經移除，user 要 ccdrawu 自己養一隻：叫 pet（不叫 buddy、不抽卡），預設開、是貓，放在 sidebar 最上面（台詞 1 列＋走道），inline 佔最左一欄（寬 14，不畫台詞與球）。圖是自己畫的：vscode-pets 的圖是 GIF、授權屬於各個畫師，只參考玩法。
  - **畫法**：每格 12×8 px，用 `▀▄` 半格字元畫成 12 欄×4 列；透明的像素是空白 Box，透出 drawer 的底色。5 種動物（`SPECIES`：cat dog rabbit duck penguin）× 7 格（walk1 walk2 sit sleep lie swipe1 swipe2）。記號：`#` 毛、`s` 斑紋／耳朵、`w` 胸腹、`x` 賓士貓多的白、`e` 眼、`n` 鼻、`p` 粉紅、`b` 喙與鴨／企鵝的腳、`q` 四肢（貓狗兔的四隻腳、鴨的翅膀、企鵝的鰭）。每組花色都明訂 `q` 的顏色，所有姿勢一開始就一起上色（user 要的；測試守著每格都有 `q`、`q` 跟 `#` 不同色、每個記號在每組花色都有顏色）。花色在 session.start 隨機抽，不給設定（user 定的）。
  - **否決**：ASCII（user 算過格數，像素佔的差不多、好看得多）；另外畫的跳格（看起來扁，改成 walk2 整隻抬高 1→2→1 px）；爬牆（另畫的攀爬格讓動物變小、用揮爪格抬高又很怪，三種畫法都否決，整個拿掉）；只有揮爪時手才換色（四肢一開始就要不同色）；從毛色自動算四肢色（每組花色明訂）；Raster＋blit（背景色不保證跟 drawer 一樣）。
  - **行為**是純函式（`hooks/pets.ts` 的 `petEvent`、`petTick`，亂數由參數注入，測試直接跑），`$.clock.every(200)` 走一步；沒變化就回同一個物件，不寫 state、不重畫。事件：`turn.step`（主對話）→ 工作中，閒逛、偶爾休息、跳、揮爪、跑；`turn.complete` 依 `reason`：answer 坐下說完成的話、aborted 安靜坐下、error／refusal 跳一下說擔心的話；`classic.PostToolUseFailure`（不是 interrupt）跳；`session.compact` 說一句；閒著 2 分鐘趴、10 分鐘睡（z 用文字 `z`→`zz`→`zzZ`，user 說不用像素畫）。台詞固定，不叫模型、不花 token（user 定的）。`/ccdrawu pat`、`/ccdrawu ball` 從 `command.run` 的 `e.args` 來，其他參數照舊是開關 drawer。
  - **球**從寵物遠的那一側滾進來（1.2 秒減速、彈 2,1,0,1,0 px），停在中段；寵物跑到球的近側 5 格（`BALL_REACH`，畫面上剛好碰到球），等球停了揮爪 1.6 秒，球消失、說一句；20 秒沒拿到就自己消失。球是 `position="absolute"` 的 Box，疊在走道上不佔位置，位置跟寵物一樣依寬度比例換算（`along()`）。
  - **框固定寬高、每列 `height={1}`**：只有空白 Box 的列會被排版壓成 0 高，睡覺（頭頂空 2 px）或跳的時候框會變高變矮，下面的清單就跟著跳。
  - **舊 state 要補齊**：`/reload-plugins` 後 `$.state` 保留舊值，session.start 用 `{ ...PET_START, ...old }` 補上新加的欄位，否則讀到 undefined 會算出 NaN。
  - **物種在 TUI 設**：preferences 的 Pet（radio，畫在原地），放在 Drawer background 與 Plugin 之間，寫進 config.json 的 `pet.species`。Go 的 `config.PetSpecies` 要跟 `SPECIES` 同順序；mod 不認得的物種畫成貓。
  - **不給取名（2026-10-15）**：做過 Pet name（最多 16 字），user 拿掉：花色每個 session 隨機，每次開都不是同一隻，給名字很怪；留給之後的版本再看。台詞那一列只有台詞。
- **不用 Fragment。** 這個 surface 把 `<>…</>` 當直排容器排版，放在 row 裡的「·」和 effort 會被疊成上下兩行。同一列的元素都直接放進 `flexDirection="row"` 的 Box。
- **bar 寬度要扣 paddingX。** 外層 Box 左右各 1 格 padding，`bodyColumns` 沒扣掉的話，bar 會超出寬度、折到下一行，看起來像兩條 bar。
- **key 欄寬 9。** key 最多 8 個字（測試守著），加一格空白。
- **會擋住引擎動作的 hook 都有 `.catch`。** `agent.spawn`、`session.compact`、`classic.*` 事件都是 gating hook；側邊欄出錯時 `.catch(($, e, next) => next(e))` 讓引擎照常進行（`next` 已呼叫過就沿用那次結果，不會重跑）。

## 已知的牆

- **權限模式有延遲。** 沒有「權限模式改變」的事件。Claude 工作中按 `shift+tab`，等目前這一步的工具跑完才更新；閒置時要等下一個 prompt。mod API 沒有讀目前模式的方法；`statusLine` 指令收到的 JSON 也沒有：2026-10-14 在 Claude Code 2.1.299 把它錄下來看過，按 `shift+tab` 時狀態列馬上重跑，但前後只差 `cost.total_duration_ms`，沒切換時它也會這樣重跑，連「重跑了」都不能當成模式變了的訊號。
- **tmux 256 色。** Claude Code 偵測到 `$TMUX` 就把顏色限制在 256 色，hex 色會被換成最接近的 256 色（`#11111b` 會變成第 17 號深藍）。要全彩得設 `CLAUDE_CODE_TMUX_TRUECOLOR=1`。
- **組織管理的環境。** `session.compact`、`skill.prompt` 與 `classic.*` 事件在組織管理的 Claude Code 裡會略過使用者安裝的外掛，Mode、Compact 和 Skills 會是空的，compact 後的 Context 也要等下一輪結束才更新，寵物不會在工具失敗時跳、compact 後說話。

## 未做

- 配色可設定（目前寫死 Catppuccin Mocha）
- README 的截圖或 demo
- 寵物取名（做過又拿掉，見設計決定的寵物 pet；之後的版本再看）
- CHANGELOG：release workflow 從它取 release notes，第一次發布前要有
- 沒在真的 Windows 上跑過：mod 的 `$.fs` 收不收 `C:\Users\me/.ccdrawu/...` 這種混用分隔的路徑、Windows 上 rename 蓋過 mod 正在讀的檔會不會失敗（TUI 存設定靠 rename）、`.ps1` 的語法（這台 Mac 沒有 PowerShell）

## 開發

### 載入

- 只跑一次：`claude --plugin-dir ~/Documents/sideproj/ccdrawu`
- 長期：把這個資料夾當市集安裝；市集是資料夾、`source` 是相對路徑時，Claude Code 直接讀資料夾本身，改完在 session 裡跑 `/reload-plugins` 就生效，不用重新安裝：

  ```sh
  CCDRAWU__MARKETPLACE=~/Documents/sideproj/ccdrawu ccdrawu install claude
  ```

### 設定程式（Go）

- 跟家族其他 app 同一套建置：`cmd/ccdrawu`（命令列）、`internal/config`（E3 位置、catalog 與設定檔）、`internal/ui`（TUI 與 splash）、`internal/claude`（`ccdrawu install`／`uninstall claude`）、`internal/version`（release 時由 goreleaser 的 ldflags 注入）。`make` 列出所有指令；`make run` 跑 TUI，`make build` 編出 `./ccdrawu`。
- Windows 版編不編得過：`GOOS=windows GOARCH=amd64 go build ./cmd/ccdrawu`（arm64 同），或 `make snapshot` 一次打六個平台的包到 `dist/`。`make release-check`（`goreleaser check`）要有 GitHub remote 才過。
- 本機同時改 terminu-go 時用 `go work init . ../terminu-go`（`go.work` 不進版控）。
- `internal/ui/icon.svg` 是 `docs/icon.svg` 的複本：splash 用 `go:embed` 讀 icon，讀不到上層目錄。換 icon 時兩個一起換，`TestSplashIconIsDocsIcon` 會抓不一致。

### 檢查

- `make check`：gofmt、`go vet`、`go test` 加上 `make mod-check`（下面兩個 plugin 指令），commit 前跑。有 `go.work` 時另外跑 `GOWORK=off make check`，確認對 go.mod 裡的 terminu-go 版本也過。
- `claude plugin validate .`：讀 manifest、市集檔和 hooks module，列出 hook、呼叫的 API 和引擎會拒絕的地方。
- `claude plugin test .`：跑 `tests/*.test.ts`。
- 型別檢查：Claude Code 載入外掛後會在 `.claude-plugin/types/` 放型別檔、在根目錄放 `tsconfig.json`（兩者都在 `.gitignore`），之後 `tsc -p .` 即可。
