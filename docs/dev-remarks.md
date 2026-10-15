# ccdrawu 開發者備忘

開發 ccdrawu 時要提醒自己、以及與 AI 協作時記下的決策。使用者要知道的在 README，這裡只放開發者需要的。

---

## 運作方式

ccdrawu 是一個 Claude Code mod（function hooks 外掛）：`hooks/register.tsx` 在 `session.start` 開一個 pane（config 把 drawer 關掉時不開），`ui.render` 依 `$.state` 裡的值畫它。狀態列另外由 `ccdrawu statusline` 畫，見下面的「狀態列」。每個值由下表的來源更新，寫入 `$.state` 後引擎只重畫讀到它的地方。

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
| Git、Path、Directory | `git rev-parse --abbrev-ref HEAD`、`git diff HEAD --shortstat`；`$.session.cwd()`（Directory 是它的最後一層，`directoryName`） | 每 5 秒 |
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
| Status | `https://status.claude.com/api/v2/status.json` 的 `status.indicator`／`description`；只在勾了 Status 時查（預設關，沒勾的人不該有這個連線；2026-10-15 user 定的） | 讀到設定時剛打開就查一次，之後每 5 分鐘 |

### 狀態列

- Claude Code 的 `statusLine` 設成 `{"type":"command","command":"ccdrawu statusline","refreshInterval":N}`：Claude Code 在對話有進展時、以及沒有動靜時每 N 秒跑一次，stdin 給一份 JSON（`session_id`、`model`、`context_window`、`cost`…），stdout 的第一行就是狀態列。
- mod 在 config 的 `statusline.on` 為真時，每秒把每一項的值照 drawer 上方那種一列的寫法（`inlineValue(view(…))`，含顏色）算好，寫進 `<data>/sessions/<session id>.json`（`hooks/statusline.ts`，內容沒變就不寫）。session id 每次重問：`/clear` 之後 session 換了新的 id，Claude Code 給指令的也是新的那個。pet 不寫（狀態列只有一列）；Tasks 改用 `, ` 分隔（狀態列本身用 `|`）。
- `ccdrawu statusline`（`internal/statusline`）只負責排：讀 config 的項目與順序、Format（label／short／value）、Color、Separator，從 stdin 只拿 `session_id` 找檔案；檔案還沒有（外掛沒跑、或還沒寫）時，用 stdin 有的 Model（`model.id`，跟 drawer 一樣是 id）、Context、Cost 頂著。寬度是 `COLUMNS` 減 4：Claude Code 左右各留兩格，比這個長的它會自己從中間截斷、用值的顏色畫 `…`，所以我們先截、`…` 用分隔線的顏色。
- 跑完順便刪掉一週沒寫過的 session 檔：mod 的 `$.fs` 沒有刪檔，只能由指令這邊清。

## 設計決定

- **改名 ccdrawu，加入 terminu 家族（2026-10-14）。** 原名 statusdrawer、指令 `/sdrawer`。設定介面改用 terminu-go 做成獨立執行檔，由它安裝、設定 mod，入口本身就是 terminu app，所以照家族慣例取 `*u` 的名字。`cc` 點出是給 Claude Code 用的，又不把「Claude」放進產品名；repo 不加 `claude-code-` 前綴，跟家族其他 repo 一致。執行檔、外掛、斜線指令都叫 ccdrawu。
- **指令叫 `/ccdrawu`，別名交給使用者（2026-10-14）。** 要開源，所以選不容易跟內建指令或其他外掛撞名的名字。嫌長的人自己設別名，例如 `ccd`。`command.run` 因此是不帶 matcher 的 hook，每個斜線指令都會經過，所以加了 `.catch`：出錯時不是自己的指令就照常放行。
- **別名存在 config.json 的 `alias`，不用 `userConfig`（2026-10-14）。** 跟項目開關放同一個檔、由 TUI 設定，開著的 session 1 秒內讀到新名字就 `$.command.register`（`hooks/commands.ts`）。mod API 沒有移除指令的方法，所以一個名字註冊過就留到 session 結束：改掉或清空的別名在那之前照樣能用、也還列在選單裡。否決：TUI 改 Claude Code 的 `settings.json` 的 `pluginConfigs`（跟 `/config` 那一列同步）——要動別人的設定檔，開著的 session 也不確定會不會跟著換。
- **設定 TUI 是側欄＋明細（2026-10-14）。** `[1] PREFERENCES` 分組：DRAWER（sidebar、inline）、SETTINGS（preferences：所有設定放在同一頁——Alias、Claude theme、Drawer background、Plugin，user 要的，跟 locku 的 Settings › preference 一樣）；`[2]` 是游標那一項，即時換內容與 label（tdp v0.6.2 `components/layout/panel` 的「分組」「明細」，terminu-go v0.1.2 的 `Row.Group`、`List.OnMove`）。側欄寬 24，跟 locku、webu 一樣。在 `[1]` 按 `Enter` 進 `[2]`（tdp v0.6.3 給 K3 開的明細例外，terminu-go v0.1.3 的 `Panel.Details`、`Action.ToDetails`），`Tab`、`2` 也行；每一項都有明細，所以都進得去。別名的檢查（`config.CleanAlias`）照 mod 的 `commandNames`，mod 不收的名字在輸入框就擋下。
- **sidebar、inline 的 `[2]` 每組一列，`Enter` 開一個全部項目的 checkbox popup（2026-10-14）。** 原本每一組是畫在 panel 上的 checkbox，照 tdp `components/input/checkbox` 在 panel 上要按 `Enter` 翻（`Space` 開 Space menu），user 覺得不好用；照規範「五個以內畫在原地、超過五個開 popup」改成混著用，同一頁又會有兩種操作。user 定的做法：`[2]` 每組一列、列出勾了哪些（用 drawer 上的 key，跟側邊欄看到的同一個字），panel operation `e` Edit 開一個 popup 列出全部項目（一定超過五個，照規範就是 popup），`Space` 勾、`Enter` 一次套用、`Esc` 不改。popup 每列是「組名、key、名稱」對齊成三欄（名稱跟 key 一樣就省掉），`/` 篩選打組名也篩得到。這兩頁的 Space menu 只有這一項。之後 user 要每一列的 `Enter` 也開同一個 popup：每組的值本來就只在這個 popup 裡改，跟 Alias 那種「值不畫在原地的欄位，`Enter` 開它的 input popup」是同一套（terminu-go 的 `Field.Edit`），不違反規範；`e` 留在 Space menu，但不寫進 hint，免得 hint 變成 `Enter:edit e:edit`。從哪一列打開，cursor 就停在那一組的第一項（`e` 從第一項開始）：terminu-go v0.1.5 的 `Checks.Cursor`，ccdrawu 提的 CD8，規範沒缺、只補 lib。
- **`[2]` 改成 Items、Order 兩列（2026-10-15）。** user 要能調整 drawer 的顯示順序，`[2]` 不再每組一列：兩列的值照 tdp components/input 的顯示方式，用 `, ` 接起來、太長由 lib 截斷加 `…`：Items 是勾了的 key（照 catalog 的順序），`Enter`（或 `e`）開原本那個全部項目的 checkbox popup，從第一項開始（CD8 的 `Checks.Cursor` 因此不再用到）；Order 列出勾了的 key，照 drawer 畫的順序，不含 pet（mod 把 pet 另外畫在最上面或最左邊，在陣列哪裡都一樣）。排順序的 popup 叫 reorder（user 定的名字：sort 是照規則排出來的順序，reorder 是使用者自己改的），是 user 要的獨立元件，等於把 panel 的 Drag mode 放進 dialog；terminu-go 沒有，請它新增（CD9）。terminu 寫進 tdp v0.6.6 的 components/input/reorder（user 在 terminu 的 session 定的）：`D` 拿起（不用 `Space`：K5 規定 popup 上 `Space` 不作用、只有 checkbox 例外），`j|k` 跟上下一項交換、到頭尾繞回，拿著時 `Enter` 放下、`Esc` 放回原位；沒拿東西時 `Enter` 寫回整個順序，`Esc` 在改過時先問要不要放棄；不做 `/` 篩選；不叫 sort，因為 tdp 已有依欄位排序的「排序選擇器」。terminu-go v0.1.6 的 `App.Reorder`（2026-10-15 升上）。popup 列出勾了的項目（不含 pet），跟 checkbox popup 一樣寫成組名、key、名稱三欄；寫回時（`Config.Reorder`）把新順序放進那些項目在陣列裡原本的位置，pet 與 catalog 不認得的 id 留在原位；勾了不到兩項時 Order 停用。順序就是 config.json 的 `dock`／`inline` 陣列：mod 照陣列順序畫（`enabledItems`），TUI 打勾不再排回 catalog 的順序（`Config.Set`），新勾的放在 catalog 裡它前面最近一個有勾的後面，沒排過的清單因此維持 catalog 的順序。否決：照 kbu 把順序清單放在 panel 上、用 panel 的 Drag mode 拖（user 要的是 dialog）。
- **狀態列（2026-10-15）。** 有些人就是要狀態列，user 要 ccdrawu 也能當狀態列，挑項目、順序與寫法（狀態列很擠，原本要解決的就是這個問題，所以有只寫值的 Format）。跟 drawer 是兩件事、各自開關（user 定的）。
  - **用 Claude Code 原生的 `statusLine` 呼叫子指令**（user 定的，跟 ccstatusline 一樣）。否決 `$.ui.status`：2026-10-15 實測，它被畫成亮黃色、前面加上 `⚠ <外掛名>: `，ESC 會變成 `�`，只有一列、換行也變 `�`。
  - **打開時換掉使用者原本的狀態列**：原本有的話先跳 confirm（寫出原本的指令），原本的整段 JSON 記在 config 的 `statusline.before`（沒有就是 `null`），關掉或 `uninstall claude` 時原樣放回。`before` 在 config.json 裡存成字串（`config.Kept`）：存成 JSON 的話，`WriteConfig` 會把它重新排版，重開 TUI 或 `uninstall claude` 從檔案讀回來時縮排就變了（0.1.1 的寫法，0.1.2 修掉；0.1.1 存的物件照樣讀得到，壓成一行）；跟主題一樣只動最外層那個 key、其他位元組不變（`internal/claude/statusline.go`，測試守著換過去再換回來一模一樣）。使用者之後自己換掉的就不動；TUI 開啟時發現 config 是開、但 `statusLine` 已經不是我們的，就把 config 改成關。terminu-go 的 `Confirm` 沒有取消的 callback，所以 radio 先退回 off，按了確定才設成 on。
  - **值由 mod 算、指令只排**（見運作方式的「狀態列」）：drawer 的值大多 stdin 沒有（effort、git、tasks…），指令自己重算又會跟 drawer 不一致；ccstatusline 每次重掃整份 transcript，ccdrawu 不必。
  - **Model 用 id**：mod 只有 `$.session.model()`（id，跟 `/model` 一樣），型別檔裡沒有顯示名稱；fallback 也改用 stdin 的 `model.id`，兩條路才一致（user 選的）。
  - **on when narrow**（user 要的第三個選項）：終端機一窄，sidebar 就收起來或跑到輸入框上方，inline 又太高，這時改看狀態列。條件是終端機窄、而且 drawer 不在對話旁邊：mod 記下 `ui.render` 最後一次的 `placement`，每秒用 `$.ui.panes()` 看 pane 開著而且放上去了，兩者都成立才是 dock，寫進 session 檔的 `docked`；`ccdrawu statusline` 在 config 的 `statusline.when` 是 `narrow`、而且 `docked` 為真或 `COLUMNS` 達 144（API 寫的：沒人要求的 pane 從 144 欄開始 dock，要求過的 110）時印空行。原本叫 on when inline、只看 `docked`：user 回報 fullscreen 時沒消失。一是 `/reload-plugins` 之後 mod 的變數歸零，要等 pane 再畫一次才知道在 dock（寵物睡著、數字不動時要等好一陣子），這段時間 `docked` 是 false；二是寬的時候把 drawer 關掉也會跑出來。user 要的其實是「窄的時候才出現」，所以加上寬度，寬的時候一律空白。不顯示時印的是只有一個 `\x1b[0m`（`hidden`）的一行：2026-10-15 在 Claude Code 2.1.300 實測，只要設了 `statusLine` 就有那一列，指令沒印東西、只印換行或空白、exit 1，都是空白的一列；只印一個 SGR reset（有沒有換行都一樣）那一列就整個收起來。這是量出來的行為，不是文件寫的，Claude Code 改了就會退回空白的一列。設定是所有 session 共用的（`$.config.set` 只管 `/config` 的列、寫的是全域設定），不能每個 session 自己切。沒有項目可顯示時也印 `hidden`。實測縮窄 2 秒內出現、拉寬 1 秒內變空、`/ccdrawu` 關掉也會出現（refreshInterval 1）。
  - **Refresh 預設 10 秒**（跟 ccstatusline 一樣），1～60 秒由使用者設；狀態列打開時改了就直接寫進 `settings.json`。
  - **Color**：`highlighted` 照 drawer 的顏色；`dimmed` 整列不帶任何色碼，交給 Claude Code 畫成 dim（2026-10-15 實測它包成 `ESC[2m ESC[37m`），跟介面上其他暗字同色，不另挑一個灰。
  - **short label**：每一項多一個最多 4 字的 `short`（ctx、5h、7d…），Format 選 short 時用；catalog 沒有（舊版 mod）就退回 label。預設開的是 Model、Context、Session、Weekly、Cost、Git。
  - TUI 的 `[1]` 多一組 STATUSLINE（排在 DRAWER 與 SETTINGS 中間），頁面依序是 Statusline 開關、Items、Order、Format、Color、Separator、Refresh；底下的 status bar 改說「the statusline follows within N seconds」。
- **drawer 的開關（2026-10-15）。** user 不想每次開 session 都先冒出來再手動關掉，sidebar、inline 兩頁最上面各有一個 Drawer off／on（跟 Statusline 頁的排法一樣），寫成 config 的 `drawer.sidebarOff`、`drawer.inlineOff`（預設開，開的時候不寫）。0.1.1 原本是兩頁共用一個 `drawer.off`，user 回報這是 bug：要的是寬的時候有 sidebar、窄的時候不要又高又擠的 inline（改看狀態列的 on when inline）；`off` 照樣讀得到，當成兩邊都關。
  - **放哪裡由引擎決定**，2026-10-15 在 Claude Code 2.1.300、fullscreen 下量的：寬的時候沒人要求的 open 放在 dock；dock 著的 pane 縮窄會**變成 inline**（user 碰到的就是這個）；窄的時候沒人要求的 open 不畫、等變寬（`isPlaced` false）；`/ccdrawu` 這種要求的 open 窄的時候放 inline；inline 什麼都不畫還是會留一個上下兩列的框。所以「某個位置關掉」只能把 pane 關掉，不能畫空的。
  - **做法**（`keepPlace`）：pane 每次畫到新的位置就檢查，那個位置關了就關掉 pane；是 inline 而且 fullscreen、sidebar 又開著，就馬上再沒人要求地 open 一次，引擎會讓它等到夠寬再放進 dock。sidebar 關掉時不重開（會再放進 dock）。所以 fullscreen 下 inline 不會自己出現，只有 `/ccdrawu` 或非 fullscreen 才會。`/ccdrawu` 開的那一次照引擎放的位置留著，位置一變就照設定走。實測：sidebar 開、inline 關，寬→dock、窄→收起來（狀態列出現）、再寬→回 dock。
  - 兩邊都關時 session 開始不開 pane；TUI 切換時開著的 session 一秒內跟著開、關或搬（只在讀到的值變了的那一次動作，第一次讀只記下來，reload 不會把手動開的關掉）。session.start 先 `await` 讀完 config 才決定要不要開。
  - `/ccdrawu` 遇到開著但還沒畫出來（等變寬）的 pane 當作關著：再要求一次 open 就畫出來。原本要按兩次，第一次是把等著的 pane 關掉。
- **`[2]` 的欄位標題（2026-10-15）。** user 要每頁都有：drawer 與 statusline 的頁是 SETTING／VALUE，preferences 是 KEY／VALUE（terminu-go 的 `Values.Header`）。user 要標題用 Overlay0，但 tdp v0.6.11 規定標題是 Blue，跟 label 同色，表單上看不出是標題：寫成 CD10 交給 terminu，定成 tdp v0.6.12（一列一個值的 panel 表頭用 Overlay0，清單維持 Blue），terminu-go v0.1.11 照做，ccdrawu 從 v0.1.7 直接升上去，沒改程式。
- **也給 Windows 用，建置照 kbu（2026-10-14）。** Claude Code 三個平台都有，家族其他 app 是 unix-first，只有 kbu 出 Windows 版。goreleaser 多出 windows 的 amd64、arm64（zip）、發到 Scoop（`vulcanshen/scoop-bucket`），`install.ps1`／`uninstall.ps1` 照 kbu（裝到 `%LOCALAPPDATA%\ccdrawu`、加進使用者的 PATH）；release 的測試也在 windows-latest 跑（那邊沒有 cgo，不加 `-race`）。位置照 E3：Windows 上 `~` 就是 `%USERPROFILE%`，不用 `%APPDATA%`（tdp v0.6.4，ccdrawu 提的 CD4，user 選的）。mod 先讀 `USERPROFILE` 再讀 `HOME`，跟 Go 的 `os.UserHomeDir` 同一個順序，兩邊才會算出同一個資料夾。
- **先裝 ccdrawu，再 `ccdrawu install claude`（2026-10-14）。** 入口是這支程式，外掛由它裝（user 定的）：`claude plugin install ccdrawu --marketplace <來源> --json`（user scope；市集還沒加會先加；已經裝了就是沒變化），來源預設 `vulcanshen/ccdrawu`，`CCDRAWU__MARKETPLACE` 改指別處（開發時指本機資料夾）。`uninstall claude` 先 `plugin uninstall <id>` 再 `marketplace remove <市集名>`（市集只有 ccdrawu），設定與資料留著、告訴使用者在哪；狀態列是 ccdrawu 的就換回原本的（主題也是），外掛已經不在 Claude Code 裡時也一樣（0.1.1 在這裡直接結束，留下指著 ccdrawu 的 `statusLine`，0.1.2 修掉）。移除腳本刪執行檔前先跑一次 `uninstall claude`，`statusLine` 才不會指著刪掉的執行檔。`claude` 這個目標先佔著位置，之後要裝到別的地方再加。
- **不顯示 vim 的 INSERT／NORMAL（2026-10-14）。** mod API 拿不到 vim 的即時狀態，唯一的來源是 Claude Code 跑 `statusLine` 指令時給的 JSON：先做過外掛附的 sh 轉接（要 sh 與 jq），再改成 `ccdrawu bridge` 子指令（Windows 也能用）。但 `statusLine` 只能填一個指令，佔用它就會跟使用者自己的狀態列（例如 ccstatusline）互相蓋掉；而 Claude Code 輸入框下方本來就顯示 vim 模式。所以整列（原本叫 Editor，後來叫 Vim）連同轉接一起拿掉。後來狀態列做成使用者自己打開、先問過才換掉的功能（下一條），vim 模式仍然不顯示。
- **不畫底色（2026-10-14）。** 停靠的 pane 外框（上方 × 那列、左邊分隔欄、下方）由引擎的容器畫，底色是主題的 `composerSidebarBackground`（`dark-ansi` 是 `ansi:blackBright`，`dark` 是 `rgb(38, 38, 38)`）。mod 只能畫 body，蓋不到外框；只蓋 body 會變成「外框一個色、內容一個色」。所以底色交給主題，由 TUI 的 theme／background 寫自訂主題（下一條）。這個 token 不在官方文件的清單裡，是從 Claude Code 程式找到的，改名時主題會忽略它，退回灰底。
- **主題由 TUI 寫進 Claude Code（2026-10-14）。** user 選的做法：選一個內建主題當底，加上側邊欄底色。`internal/claude` 寫 `<Claude 資料夾>/themes/ccdrawu.json`（`CLAUDE_CONFIG_DIR`，否則 `~/.claude`），把 `settings.json` 的 `theme` 換成 `custom:ccdrawu`，換之前的值記在 ccdrawu 的 config.json（`theme.before`），選 off 或 `uninstall claude` 時換回去；使用者之後在 `/theme` 自己換過的就不動。改 `settings.json` 只改最外層 `theme` 那一個值，其他位元組原樣（`json.Decoder` 的 offset，測試守著「換過去再換回來，檔案一模一樣」），用暫存檔 rename、保留權限。否決：交給 mod 用 `$.config.set` 換——2026-10-14 實測 theme 那一列是 `choice`，只收 `auto` 與 6 個內建主題，`custom:…` 會被拒絕。內建主題清單（`claude.Themes`）是從 Claude Code 2.1.299 的執行檔裡找出來的。2026-10-14 實測：TUI 改了 `settings.json` 的 `theme` 與主題檔，開著的 session 立刻換色，不用重開。
- **寵物 pet（2026-10-15）。** Claude Code 的 `/buddy` 已經移除，user 要 ccdrawu 自己養一隻：叫 pet（不叫 buddy、不抽卡），預設開、是貓，放在 sidebar 最上面（台詞 1 列＋走道），inline 佔最左一欄（寬 14，不畫台詞與球）。圖是自己畫的：vscode-pets 的圖是 GIF、授權屬於各個畫師，只參考玩法。
  - **畫法**：每格 12×8 px，用 `▀▄` 半格字元畫成 12 欄×4 列；透明的像素是空白 Box，透出 drawer 的底色。5 種動物（`SPECIES`：cat dog rabbit duck penguin）× 7 格（walk1 walk2 sit sleep lie swipe1 swipe2）。記號：`#` 毛、`s` 斑紋／耳朵、`w` 胸腹、`x` 賓士貓多的白、`e` 眼、`n` 鼻、`p` 粉紅、`b` 喙與鴨／企鵝的腳、`q` 四肢（貓狗兔的四隻腳、鴨的翅膀、企鵝的鰭）。每組花色都明訂 `q` 的顏色，所有姿勢一開始就一起上色（user 要的；測試守著每格都有 `q`、`q` 跟 `#` 不同色、每個記號在每組花色都有顏色）。花色在 session.start 隨機抽，不給設定（user 定的）。
  - **否決**：ASCII（user 算過格數，像素佔的差不多、好看得多）；另外畫的跳格（看起來扁，改成 walk2 整隻抬高 1→2→1 px）；爬牆（另畫的攀爬格讓動物變小、用揮爪格抬高又很怪，三種畫法都否決，整個拿掉）；只有揮爪時手才換色（四肢一開始就要不同色）；從毛色自動算四肢色（每組花色明訂）；Raster＋blit（背景色不保證跟 drawer 一樣）。
  - **行為**是純函式（`hooks/pets.ts` 的 `petEvent`、`petTick`，亂數由參數注入，測試直接跑），`$.clock.every(200)` 走一步；沒變化就回同一個物件，不寫 state、不重畫。事件：`turn.step`（主對話）→ 工作中，閒逛、偶爾休息、跳、揮爪、跑；`turn.complete` 依 `reason`：answer 坐下說完成的話、aborted 安靜坐下、error／refusal 跳一下說擔心的話；`classic.PostToolUseFailure`（不是 interrupt）跳；`session.compact` 說一句；閒著 2 分鐘趴、10 分鐘睡（z 用文字 `z`→`zz`→`zzZ`，user 說不用像素畫）。台詞固定，不叫模型、不花 token（user 定的）。`/ccdrawu pat`、`/ccdrawu ball` 從 `command.run` 的 `e.args` 來，其他參數照舊是開關 drawer。
  - **球**從寵物遠的那一側滾進來（1.2 秒減速、彈 2,1,0,1,0 px），停在中段；寵物跑到球的近側 5 格（`BALL_REACH`，畫面上剛好碰到球），等球停了揮爪 1.6 秒，球消失、說一句；20 秒沒拿到就自己消失。球是 `position="absolute"` 的 Box，疊在走道上不佔位置，位置跟寵物一樣依寬度比例換算（`along()`）。
  - **框固定寬高、每列 `height={1}`**：只有空白 Box 的列會被排版壓成 0 高，睡覺（頭頂空 2 px）或跳的時候框會變高變矮，下面的清單就跟著跳。
  - **舊 state 要補齊**：`/reload-plugins` 後 `$.state` 保留舊值，session.start 用 `{ ...PET_START, ...old }` 補上新加的欄位，否則讀到 undefined 會算出 NaN。
  - **物種在 TUI 設**：preferences 的 Pet（radio，畫在原地），放在 Bar 與 Plugin 之間，寫進 config.json 的 `pet.species`。Go 的 `config.PetSpecies` 要跟 `SPECIES` 同順序；mod 不認得的物種畫成貓。
  - **不給取名（2026-10-15）**：做過 Pet name（最多 16 字），user 拿掉：花色每個 session 隨機，每次開都不是同一隻，給名字很怪；留給之後的版本再看。台詞那一列只有台詞。
- **不用 Fragment。** 這個 surface 把 `<>…</>` 當直排容器排版，放在 row 裡的「·」和 effort 會被疊成上下兩行。同一列的元素都直接放進 `flexDirection="row"` 的 Box。
- **bar 寬度要扣 paddingX。** 外層 Box 左右各 1 格 padding，`bodyColumns` 沒扣掉的話，bar 會超出寬度、折到下一行，看起來像兩條 bar。
- **bar 用 Nerd Font 的方塊（2026-10-15）。** 填滿的是 `nf-oct-square_fill`（U+F445），空的是 `nf-oct-square`（U+F51D），user 挑的。原本用 `▰▱`：很多程式字型沒有這兩個字（Maple Mono、JetBrains Mono、Fira Code 都沒有），終端機只能跟別的字型借，借到哪一套看那個終端機開啟後先出現過哪些缺的字（Ghostty 先找已經借過的字型）。2026-10-15 user 的 Ghostty 開機後第一個缺的字就是 bar，借到 Apple Symbols，字比一格寬，整串黏成一塊；前一天先出現過 Claude Code 的轉圈符號 `✻`，借到的是 Menlo，畫出來是正常的小平行四邊形。程式、字型、終端機都沒變，只差在出現順序。Nerd Font 的字在字型自己裡面，不用借；Maple Mono NF CN 與 JetBrains、Hack 的 Mono 版，方塊寬 10、一格 12，左右分開，高度只佔一列的中段，上下兩列也分開。代價是 drawer 也要 Nerd Font（README 的需求寫了）。終端機不會告訴程式它用什麼字型，mod API 也沒有字型或終端機能力的資訊（2.1.300 的型別檔查過），查電腦裝了哪些字型也不準（裝了不一定在用；沒用 Nerd Font 也可能從 Symbols Nerd Font 借到），所以不自動偵測：preferences 的 **Bar** 讓使用者選 squares（預設，user 定的）或 line（`━`，框線字元，等寬字型都有），每個選項旁邊畫著自己的 bar，畫不出來的那個一看就知道；寫進 config.json 的 `bar`，mod 不認得的值畫方塊。另外，非 Mono 版的 Nerd Font（例如 MesloLGS NF）方塊比一格寬（-0.8～13.65），有的終端機上可能又黏在一起。README 的範例用 `■□` 示意，因為 GitHub 網頁沒有 Nerd Font。否決：`█░`（整格高，上下兩列的 bar 會黏在一起）；請使用者把 `▰▱` 指定給 Menlo（只有 macOS 有，ccdrawu 也支援 Windows、Linux）。
- **key 欄寬 9。** key 最多 8 個字，加一格空白；只有 Directory 是 9 個字（user 定的名字，跟 Path／Working directory 對應），所以欄寬照目前顯示的最長 key 算：沒顯示 Directory 時跟原本一樣，顯示時才寬一格。測試守著其他 key 都在 8 個字以內。
- **會擋住引擎動作的 hook 都有 `.catch`。** `agent.spawn`、`session.compact`、`classic.*` 事件都是 gating hook；側邊欄出錯時 `.catch(($, e, next) => next(e))` 讓引擎照常進行（`next` 已呼叫過就沿用那次結果，不會重跑）。

## 已知的牆

- **權限模式有延遲。** 沒有「權限模式改變」的事件。Claude 工作中按 `shift+tab`，等目前這一步的工具跑完才更新；閒置時要等下一個 prompt。mod API 沒有讀目前模式的方法；`statusLine` 指令收到的 JSON 也沒有：2026-10-14 在 Claude Code 2.1.299 把它錄下來看過，按 `shift+tab` 時狀態列馬上重跑，但前後只差 `cost.total_duration_ms`，沒切換時它也會這樣重跑，連「重跑了」都不能當成模式變了的訊號。
- **tmux 256 色。** Claude Code 偵測到 `$TMUX` 就把顏色限制在 256 色，hex 色會被換成最接近的 256 色（`#11111b` 會變成第 17 號深藍）。要全彩得設 `CLAUDE_CODE_TMUX_TRUECOLOR=1`。
- **mod 拿不到模型的顯示名稱。** `$.session.model()` 只給 id（`claude-opus-5-5`），所以 drawer 與狀態列都顯示 id。
- **組織管理的環境。** `session.compact`、`skill.prompt` 與 `classic.*` 事件在組織管理的 Claude Code 裡會略過使用者安裝的外掛，Mode、Compact 和 Skills 會是空的，compact 後的 Context 也要等下一輪結束才更新，寵物不會在工具失敗時跳、compact 後說話。
- **面板放哪裡由 Claude Code 決定。** `Pane` 的 `placement`（`dock`／`inline`）是唯讀的，hook 改了會被拒絕；`$.ui.open` 只能要大小（inline 的 `rows`、dock 的 `columns`），沒有位置。fullscreen（alternate screen）而且終端機夠寬（自動開的 144 欄、使用者開過的 110 欄）才 dock，其他都 inline。所以 TUI 不提供「位置」設定，`[2]` 的 sidebar／inline 是「放在那裡時顯示哪些」；能控制的只有開不開、開多大。2026-10-15 user 的 session 曾被重開成 default 模式（process 環境有 `CLAUDE_CODE_TUI_JUST_SWITCHED=default`、沒有 `CLAUDE_CODE_NO_FLICKER=1`，settings.json 有 `"tui": "default"`），`/tui fullscreen` 卻回「Already using the fullscreen renderer.」，drawer 一直在 inline；結束 Claude Code 再重開就回到 sidebar。怎麼切到 default 的沒查到（那段對話沒有 `/tui default`，壓縮過的程式也看不出 `/tui` 的判斷）。README 的疑難排解寫了「重開」。
- **Windows 的終端機（2026-10-15 user 的同事試裝）。** Windows Terminal：TUI、inline、sidebar 排版都對；沒有 Nerd Font 時缺的字畫成 `�`，仍佔一格；powerline 與框線字元由 Windows Terminal 自己畫，不靠字型。Termius 的本機終端：bar 的方塊在終端機佔兩格（從截圖量的：13 個方塊加右邊一段深色剛好 26 格碰到右邊界，剩下 6 個方塊和百分比折到下一行最左邊），sidebar 整個錯位；TUI 也留殘影，radio 標記（U+F0000 以上的 Material Design 圖示）畫成 `�`；sidebar 的淺色底只塗在有字的格子上。terminu-go 量圖示寬度（CPR）只在 unix 上做，Windows 預設一格，可以用 `CCDRAWU__ICON_WIDTH`／`TERMINU__ICON_WIDTH` 指定，沒在 Termius 試過。不另外處理，README 寫改用 Windows Terminal。同事後來在 Windows Terminal 換成 Maple Mono NF CN，只剩沒選中的 radio 是 `�`：Maple 7.0 缺 U+F043D（nf-md-radiobox_blank），是 ccdrawu 與 terminu-go v0.1.7 用到的 39 個圖示裡唯一缺的（`fc-list ":family=…:charset=…"` 逐一查；zsh 裡變數要寫 `${f}`，`$f:charset` 的 `:c` 會被當成修飾符）。JetBrains Mono、Hack、FiraCode 的 Nerd Font 都不缺，同事換 JetBrains Mono Nerd Font 後就正常。user 定：字型問題，不是規範問題，不請 tdp 換圖示（Maple 用的人少）。
- **淺色主題。** sidebar 的底色是 Claude Code 主題的 `composerSidebarBackground`（內建淺色主題是 `rgb(235, 235, 235)` 或 `rgb(245, 245, 245)`，`light-ansi` 是 `ansi:white`），配色寫死 Catppuccin Mocha，在淺色底上看不清（同事的 Claude Code 是淺色主題）。mod 的顏色可以用主題鍵（`text`、`subtle`、`success`…，跟著使用者的主題變），之後要支援淺色主題可以從這裡改。

## 未做

- 配色可設定（目前寫死 Catppuccin Mocha）
- README 的截圖或 demo
- 寵物取名（做過又拿掉，見設計決定的寵物 pet；之後的版本再看）
- CHANGELOG：release workflow 從它取 release notes，第一次發布前要有
- Windows 還沒驗到的：rename 蓋過 mod 正在讀的檔會不會失敗（TUI 存設定靠 rename；同事試裝時 TUI 存檔成功，但沒刻意在 mod 讀檔時存）、`.ps1` 的語法（不知道同事用哪種方式裝的，這台 Mac 沒有 PowerShell）。mod 的 `$.fs` 寫 Windows 路徑已驗到：TUI 列得出 mod 寫的 catalog

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

### 錄 gif

- `make gif` 錄 README 的兩張：`docs/demo.gif`（Claude Code 裡的 drawer 與輸入框下方的狀態列：問 `~/my-app` 一句話、寵物在 Claude 工作時走動、做完坐下說話、`/ccdrawu pat`、`/ccdrawu ball`）與 `docs/settings.gif`（設定畫面：Items 勾兩項、Order 用 `D` 搬 Cost、statusline 的 Format 換成 short、preferences 把寵物換成狗）。tape 都在 `.local/demos/`（不進版控）。
- demo 的狀態列：`setup.sh demo` 在沙盒的 config 打開 `statusline.on`，另寫一份 `~/demo-settings.json`（`statusLine` 指向 repo 剛 build 的 `./ccdrawu statusline`），tape 用 `claude --settings` 帶進去。這樣錄出來的狀態列不受真正的 `~/.claude/settings.json` 和 `PATH` 上是哪一版 ccdrawu 影響。
- settings.gif 不按 Statusline 的開關、Refresh、Claude theme、Plugin：它們寫的是 Claude Code 的設定，沙盒裡 `~/.claude` 是接到真的那一份。改了 TUI 的游標路徑就先在 tmux 用 `CLAUDE_CONFIG_DIR` 指到別處、照 tape 的按鍵走一遍，確認每次 `Enter` 落在哪裡再錄。
- 沙盒照 sshu：`.local/demos/setup.sh` 每次重做 `.local/demos/home`（全新的 ccdrawu 設定、從 `~/.ccdrawu` 複製的 catalog、重建的 `my-app`：一個 commit 加一處沒 commit 的改動），tape 再 `export HOME` 過去，畫面上的路徑就是 `~/my-app`、`~/.config/ccdrawu/config.json`，真正的設定不會被碰。Claude Code 的設定、外掛與登入用 symlink 接真的：`~/.claude`、`~/.claude.json`，還有 `~/Library`（登入在 keychain，macOS 從 `$HOME/Library` 找；沒接會是 Not logged in，而且沙盒裡會長出一個真的 `Library`，setup.sh 會換回 symlink）。
- tape 要先 unset `XDG_CONFIG_HOME`／`XDG_DATA_HOME`／`XDG_STATE_HOME`（user 的環境有設，ccdrawu 照 E3 先讀它，就會讀到真正的設定），以及從 Claude Code 裡跑 vhs 時繼承來的 `TMUX*`、`TERM_PROGRAM*`、`CLAUDECODE`、`CLAUDE_CODE_*`（留 `CLAUDE_CODE_NO_FLICKER`），否則畫面上有 tmux focus-events 與「Transcript saving is off」的提示。
- `export HOME` 之後要 `cd $HOME/my-app`：這台的 bash 5.3 換了 `HOME` 之後 `~` 還是原本的家目錄。vhs 的 `Type` 字串不收 `\"`。
- demo 會開真的 Claude Code（要 fullscreen、ccdrawu 外掛），送一個 prompt，花一點 token；問句要指定英文回答，否則照 user 的 CLAUDE.md 回繁中。`~/my-app` 第一次要先信任：`vhs .local/demos/trust.tape`（信任對話框預設是 No，要 `↓` 再 `Enter`）。
- splash 不錄進 gif：它是彩蛋，`V` 才開，tdp S2 不寫進 README。
- 寵物的動作是隨機的，錄完要看；vhs 偶爾會漏幾秒畫面（一張 gif 停格好幾秒），重錄即可。

### Social preview

- `docs/social-preview.png` 是 GitHub 的 social preview：icon 在上、下面 Tiny5 字樣 `ccdrawu`，跟家族其他專案同一個構圖。`/social-preview ccdrawu` 產生（`.claude/commands/` 與 `.claude/agents/` 從 locku 搬來），輸出在 `/tmp`，看過再複製進 `docs/`（2026-10-15）。
- GitHub 沒有上傳 social preview 的 API：repo 的 Settings → General → Social preview 手動上傳。

### 檢查

- `make check`：gofmt、`go vet`、`go test` 加上 `make mod-check`（下面兩個 plugin 指令），commit 前跑。有 `go.work` 時另外跑 `GOWORK=off make check`，確認對 go.mod 裡的 terminu-go 版本也過。
- `claude plugin validate .`：讀 manifest、市集檔和 hooks module，列出 hook、呼叫的 API 和引擎會拒絕的地方。
- `claude plugin test .`：跑 `tests/*.test.ts`。
- 型別檢查：Claude Code 載入外掛後會在 `.claude-plugin/types/` 放型別檔、在根目錄放 `tsconfig.json`（兩者都在 `.gitignore`），之後 `tsc -p .` 即可。
