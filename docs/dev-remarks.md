# ccdrawu 開發者備忘

開發 ccdrawu 時要提醒自己、以及與 AI 協作時記下的決策。使用者要知道的在 README，這裡只放開發者需要的。設定 TUI 遵循
[terminu design](https://github.com/vulcanshen/terminu)（tdp）的 principle，照 tdp v1.1.0；drawer 本身是 Claude Code 的 mod，
畫在 Claude Code 裡，tdp 的畫面規定不套到它上面。

---

## 運作方式

ccdrawu 是一個 Claude Code mod（function hooks 外掛）：`hooks/register.tsx` 在 `session.start` 開一個 pane（config 把 drawer 關掉時不開），`ui.render` 依 `$.state` 裡的值畫它。狀態列另外由 `ccdrawu statusline` 畫，見下面的「狀態列」。每個值由下表的來源更新，寫入 `$.state` 後引擎只重畫讀到它的地方。

| 列 | 來源 | 更新時機 |
|---|---|---|
| Context、Session、Weekly、Cost | `$.session.usage()` | `session.measure`（每輪結束、rate limit 變動）；compact 完成後（`classic.PostCompact`，`/compact` 不算一輪，沒有 measure）直接清空 Context 的 tokens／percent，不重讀：這時 `$.session.usage()` 還是 compact 前的值，約一秒後才清空，之後沒有事件再讀（2.1.300 實測；Claude Code 給狀態列的 stdin 當下就是 `used_percentage: null`） |
| Model | `$.session.model()` | session 開始、每輪結束 |
| Effort | `turn.step` 的 `e.effort`，依等級上色（`effortColor`：low 綠到 max 紅）；session 剛開始還沒送過請求時，先用 settings 的 `modelSettings[model].effortLevel` 或 `effortLevel` | 每次主對話送出請求 |
| Mode | 舊式 hook 事件（`classic.SessionStart`、`classic.UserPromptSubmit`、`classic.PostToolBatch`、`classic.Stop`）帶的 `permission_mode`，只認主對話 | 上述四個事件；PostToolBatch 是每一步的工具跑完後一次（`classic.PreToolUse` 在 mod 裡只有工具的內容、沒有 permission_mode，PostToolUse 每個工具一次、還帶整份輸出） |
| Uptime | `usage().startedAt` 與 `$.session.turns()` | 每 30 秒、每輪結束 |
| Cache | `turn.step` 結果的 `usage`：`cache_read / (input + cache_read + cache_creation)` | 每次主對話送出請求 |
| Compact | `session.compact` 成功（非 `precompute`、非 skip、非子 agent）時加一 | 每次壓縮 |
| Tasks | subagent：`$.agent.list()` 裡 `pending`／`running`／`waiting` 的數量（只算這個 session 的；workflow 的 agent 不在裡面）。其他：`classic.Stop` 的 `background_tasks` 按種類計數，不算 subagent——等自己背景工作的 subagent 不在這份清單裡，它開的 Monitor 算成 shell（2026-10-14 實測）。顯示用 `｜` 分隔（user 定的），它佔兩格，寬度用 `cellWidth` 算 | subagent 每 5 秒、`agent.spawn`、每輪結束；其他在每次回覆結束 |
| Git、Path、Directory | `git rev-parse --abbrev-ref HEAD`、`git diff HEAD --shortstat`；`$.session.cwd()`（Directory 是它的最後一層，`directoryName`）。太長時 Path 留尾巴（`cutStart`）；Git 的 `+N -N` 在側邊欄放在 branch 下一列（`View.below`，跟 resets 一樣），branch 太長就完整折行（原本跟 branch 同列，長 branch 折行時被拆在兩列中間；user 要完整呈現、不截斷），inline 與狀態列接在 branch 後面同一行 | 每 5 秒 |
| Claude | `$.session.version()`、npm registry 的 `@anthropic-ai/claude-code/latest` | session 開始、每 6 小時 |
| ccdrawu | 外掛自己的 `.claude-plugin/plugin.json`（`$.plugin` 只有 name 和 root，沒有版本） | session 開始 |

下面這些是 2026-10-14 從 ccstatusline 搬的（user 挑的），預設 dock、inline 都關，免得側邊欄一下子變長：

| 列 | 來源 | 更新時機 |
|---|---|---|
| Style | `$.config.list()` 裡 `outputStyle` 那一列 | session 開始、每輪結束 |
| ID | `$.session.id()` 的前 8 字 | session 開始 |
| Window | `usage().context` 的 `tokens`／`window` | 跟 Context 一樣 |
| Rate | 花費 ÷ 開了幾小時，滿一分鐘才算（太早一個請求就讓數字亂跳） | 跟 Cost、Uptime 一樣 |
| Tokens、Speed | `turn.step` 結果的 `usage` 累計（主對話；`SessionCost` 只有 usd，沒有累計 token，所以只能從 mod 載入後自己加，reload 或接續舊 session 會少算）；速度是 output tokens ÷ 從送出到收完的秒數（含等第一個字） | 每次主對話送出請求 |
| Skills | `skill.prompt` 的次數與最後的 `skill` | 每次跑 skill |
| Sync、Files、Commit | 一個 `git status --porcelain=v2 --branch`：`branch.ab`、XY 碼、`?`、`branch.oid` | 每 5 秒 |
| Repo | `git remote get-url origin` 解析成 owner/repo（`$.session.repo().name` 只有內部 repo 才有）；`--git-dir` 跟 `--git-common-dir` 不同就是 linked worktree，取 toplevel 的資料夾名；工作目錄沒變就不重查 | 工作目錄改變時 |
| Status | `https://status.claude.com/api/v2/status.json` 的 `status.indicator`／`description`；只在勾了 Status 時查（預設關，沒勾的人不該有這個連線；2026-10-15 user 定的） | 讀到設定時剛打開就查一次，之後每 5 分鐘 |

`/clear` 之後 process 用新的 session id 繼續跑，`$.state` 換成新 session 的、全部回到初始值（config、寵物、版本都在裡面），`session.start` 卻不會再來（引擎型別這樣寫，2.1.302 實測：Model、Session、Cost、版本變 `-`，drawer 回到預設的項目，寵物睡著）。所以 session 開始時讀的那些（`startSession`）在 `classic.SessionStart` 的 `source: 'clear'` 再跑一次；config 檔沒變也整份重讀（先清掉 `configStamp`）。計時器、註冊過的指令、module 的變數都還在，不用重來。新 session 的寵物重抽花色，照 README 說的「每個 session」。

### 狀態列

- Claude Code 的 `statusLine` 設成 `{"type":"command","command":"ccdrawu statusline","refreshInterval":N}`：Claude Code 在對話有進展時、以及沒有動靜時每 N 秒跑一次，stdin 給一份 JSON（`session_id`、`model`、`context_window`、`cost`…），stdout 的第一行就是狀態列。
- mod 在 config 的 `statusline.on` 為真時，每秒把每一項的值照 drawer 上方那種一列的寫法（`inlineValue(view(…))`，含顏色）算好，寫進快取目錄的 `sessions/<session id>.json`（`hooks/statusline.ts`，內容沒變就不寫）。session id 每次重問：`/clear` 之後 session 換了新的 id，Claude Code 給指令的也是新的那個。pet 不寫（狀態列只有一列）；Tasks 改用 `, ` 分隔（狀態列本身用 `|`）。
- `ccdrawu statusline`（`internal/statusline`）只負責排：讀 config 的項目與順序、Format（label／short／value）、Color、Separator，從 stdin 只拿 `session_id` 找檔案；檔案還沒有（外掛沒跑、或還沒寫）時，用 stdin 有的 Model（`model.id`，跟 drawer 一樣是 id）、Context、Cost 頂著。寬度是 `COLUMNS` 減 4：Claude Code 左右各留兩格，比這個長的它會自己從中間截斷、用值的顏色畫 `…`，所以我們先截、`…` 用分隔線的顏色。
- 跑完順便刪掉一週沒寫過的 session 檔：mod 的 `$.fs` 沒有刪檔，只能由指令這邊清。

## 設計決定

- **改名 ccdrawu，加入 terminu 家族（2026-10-14）。** 原名 statusdrawer、指令 `/sdrawer`。設定介面改用 terminu-go 做成獨立執行檔，由它安裝、設定 mod，入口本身就是 terminu app，所以照家族慣例取 `*u` 的名字。`cc` 點出是給 Claude Code 用的，又不把「Claude」放進產品名；repo 不加 `claude-code-` 前綴，跟家族其他 repo 一致。執行檔、外掛、斜線指令都叫 ccdrawu。
- **指令叫 `/ccdrawu`，別名交給使用者（2026-10-14）。** 要開源，所以選不容易跟內建指令或其他外掛撞名的名字。嫌長的人自己設別名，例如 `ccd`。`command.run` 因此是不帶 matcher 的 hook，每個斜線指令都會經過，所以加了 `.catch`：出錯時不是自己的指令就照常放行。
- **別名存在 config.yaml 的 `alias`，不用 `userConfig`（2026-10-14）。** 跟項目開關放同一個檔、由 TUI 設定，開著的 session 1 秒內讀到新名字就 `$.command.register`（`hooks/commands.ts`）。mod API 沒有移除指令的方法，所以一個名字註冊過就留到 session 結束：改掉或清空的別名在那之前照樣能用、也還列在選單裡。否決：TUI 改 Claude Code 的 `settings.json` 的 `pluginConfigs`（跟 `/config` 那一列同步）——要動別人的設定檔，開著的 session 也不確定會不會跟著換。
- **設定 TUI 是側欄＋明細（2026-10-14）。** `[1] PREFERENCES` 分組：DRAWER（sidebar、inline）、SETTINGS（preferences：所有設定放在同一頁——Alias、Claude theme、Drawer background、Plugin，user 要的，跟 locku 的 Settings › preference 一樣）；`[2]` 是游標那一項，即時換內容與 label（tdp v0.6.2 `components/layout/panel` 的「分組」「明細」，terminu-go v0.1.2 的 `Row.Group`、`List.OnMove`）。側欄寬 24，跟 locku、webu 一樣。在 `[1]` 按 `Enter` 進 `[2]`（tdp K3 的明細例外，tdp v0.6.3 開的；terminu-go v0.1.3 的 `Panel.Details`、`Action.ToDetails`），`Tab`、`2` 也行；每一項都有明細，所以都進得去。別名的檢查（`config.CleanAlias`）照 mod 的 `commandNames`，mod 不收的名字在輸入框就擋下。
- **sidebar、inline 的 `[2]` 每組一列，`Enter` 開一個全部項目的 checkbox popup（2026-10-14）。** 原本每一組是畫在 panel 上的 checkbox，照 tdp `components/input/checkbox` 在 panel 上要按 `Enter` 翻（`Space` 開 Space menu），user 覺得不好用；照規範「五個以內畫在原地、超過五個開 popup」改成混著用，同一頁又會有兩種操作。user 定的做法：`[2]` 每組一列、列出勾了哪些（用 drawer 上的 key，跟側邊欄看到的同一個字），panel operation `e` Edit 開一個 popup 列出全部項目（一定超過五個，照規範就是 popup），`Space` 勾、`Enter` 一次套用、`Esc` 不改。popup 每列是「組名、key、名稱」對齊成三欄（名稱跟 key 一樣就省掉），`/` 篩選打組名也篩得到。這兩頁的 Space menu 只有這一項。之後 user 要每一列的 `Enter` 也開同一個 popup：每組的值本來就只在這個 popup 裡改，跟 Alias 那種「值不畫在原地的欄位，`Enter` 開它的 input popup」是同一套（terminu-go 的 `Field.Edit`），不違反規範；`e` 留在 Space menu，但不寫進 hint，免得 hint 變成 `Enter:edit e:edit`。從哪一列打開，cursor 就停在那一組的第一項（`e` 從第一項開始）：terminu-go v0.1.5 的 `Checks.Cursor`，ccdrawu 提的，規範沒缺、只補 lib。
- **`[2]` 改成 Items、Order 兩列（2026-10-15）。** user 要能調整 drawer 的顯示順序，`[2]` 不再每組一列：兩列的值照 tdp components/input 的顯示方式，用 `, ` 接起來、太長由 lib 截斷加 `…`：Items 是勾了的 key（照 catalog 的順序），`Enter`（或 `e`）開原本那個全部項目的 checkbox popup，從第一項開始（`Checks.Cursor` 因此不再用到）；Order 列出勾了的 key，照 drawer 畫的順序，不含 pet（mod 把 pet 另外畫在最上面或最左邊，在陣列哪裡都一樣）。排順序的 popup 叫 reorder（user 定的名字：sort 是照規則排出來的順序，reorder 是使用者自己改的），是 user 要的獨立元件，等於把 panel 的 Drag mode 放進 dialog；terminu-go 沒有，請它新增。terminu 寫進 tdp v0.6.6 的 components/input/reorder（user 在 terminu 的 session 定的）：`D` 拿起（不用 `Space`：tdp K5 規定 popup 上 `Space` 不作用、只有 checkbox 例外），`j|k` 跟上下一項交換、到頭尾繞回，拿著時 `Enter` 放下、`Esc` 放回原位；沒拿東西時 `Enter` 寫回整個順序，`Esc` 在改過時先問要不要放棄；不做 `/` 篩選；不叫 sort，因為 tdp 已有依欄位排序的「排序選擇器」。terminu-go v0.1.6 的 `App.Reorder`（2026-10-15 升上）。popup 列出勾了的項目（不含 pet），跟 checkbox popup 一樣寫成組名、key、名稱三欄；寫回時（`Config.Reorder`）把新順序放進那些項目在陣列裡原本的位置，pet 與 catalog 不認得的 id 留在原位；勾了不到兩項時 Order 停用。順序就是 config.yaml 的 `dock`／`inline` 陣列：mod 照陣列順序畫（`enabledItems`），TUI 打勾不再排回 catalog 的順序（`Config.Set`），新勾的放在 catalog 裡它前面最近一個有勾的後面，沒排過的清單因此維持 catalog 的順序。否決：照 kbu 把順序清單放在 panel 上、用 panel 的 Drag mode 拖（user 要的是 dialog）。
- **狀態列（2026-10-15）。** 有些人就是要狀態列，user 要 ccdrawu 也能當狀態列，挑項目、順序與寫法（狀態列很擠，原本要解決的就是這個問題，所以有只寫值的 Format）。跟 drawer 是兩件事、各自開關（user 定的）。
  - **用 Claude Code 原生的 `statusLine` 呼叫子指令**（user 定的，跟 ccstatusline 一樣）。否決 `$.ui.status`：2026-10-15 實測，它被畫成亮黃色、前面加上 `⚠ <外掛名>: `，ESC 會變成 `�`，只有一列、換行也變 `�`。
  - **打開時換掉使用者原本的狀態列**：原本有的話先跳 confirm（寫出原本的指令），原本的整段 JSON 記在 config 的 `statusline.before`（沒有就是 `null`），關掉或 `uninstall claude` 時原樣放回。`before` 在設定檔裡存成字串（`config.Kept`）：存成 JSON 的話，`WriteConfig` 會把它重新排版，重開 TUI 或 `uninstall claude` 從檔案讀回來時縮排就變了（0.1.1 的寫法，0.1.2 修掉；0.1.1 存的物件照樣讀得到，壓成一行）；跟主題一樣只動最外層那個 key、其他位元組不變（`internal/claude/statusline.go`，測試守著換過去再換回來一模一樣）。使用者之後自己換掉的就不動；TUI 開啟時發現 config 是開、但 `statusLine` 已經不是我們的，就把 config 改成關。terminu-go 的 `Confirm` 沒有取消的 callback，所以 radio 先退回 off，按了確定才設成 on。
  - **值由 mod 算、指令只排**（見運作方式的「狀態列」）：drawer 的值大多 stdin 沒有（effort、git、tasks…），指令自己重算又會跟 drawer 不一致；ccstatusline 每次重掃整份 transcript，ccdrawu 不必。
  - **Model 用 id**：mod 只有 `$.session.model()`（id，跟 `/model` 一樣），型別檔裡沒有顯示名稱；fallback 也改用 stdin 的 `model.id`，兩條路才一致（user 選的）。
  - **on when narrow**（user 要的第三個選項）：終端機一窄，sidebar 就收起來或跑到輸入框上方，inline 又太高，這時改看狀態列。條件是終端機窄、而且 drawer 不在對話旁邊：mod 記下 `ui.render` 最後一次的 `placement`，每秒用 `$.ui.panes()` 看 pane 開著而且放上去了，兩者都成立才是 dock，寫進 session 檔的 `docked`；`ccdrawu statusline` 在 config 的 `statusline.when` 是 `narrow`、而且 `docked` 為真或 `COLUMNS` 達 144（API 寫的：沒人要求的 pane 從 144 欄開始 dock，要求過的 110）時印空行。原本叫 on when inline、只看 `docked`：user 回報 fullscreen 時沒消失。一是 `/reload-plugins` 之後 mod 的變數歸零，要等 pane 再畫一次才知道在 dock（寵物睡著、數字不動時要等好一陣子），這段時間 `docked` 是 false；二是寬的時候把 drawer 關掉也會跑出來。user 要的其實是「窄的時候才出現」，所以加上寬度，寬的時候一律空白。不顯示時印的是只有一個 `\x1b[0m`（`hidden`）的一行：2026-10-15 在 Claude Code 2.1.300 實測，只要設了 `statusLine` 就有那一列，指令沒印東西、只印換行或空白、exit 1，都是空白的一列；只印一個 SGR reset（有沒有換行都一樣）那一列就整個收起來。這是量出來的行為，不是文件寫的，Claude Code 改了就會退回空白的一列。設定是所有 session 共用的（`$.config.set` 只管 `/config` 的列、寫的是全域設定），不能每個 session 自己切。沒有項目可顯示時也印 `hidden`。實測縮窄 2 秒內出現、拉寬 1 秒內變空、`/ccdrawu` 關掉也會出現（refreshInterval 1）。
  - **Refresh 預設 10 秒**（跟 ccstatusline 一樣），1～60 秒由使用者設；狀態列打開時改了就直接寫進 `settings.json`。
  - **Color**：`highlighted` 照 drawer 的顏色；`dimmed` 整列不帶任何色碼，交給 Claude Code 畫成 dim（2026-10-15 實測它包成 `ESC[2m ESC[37m`），跟介面上其他暗字同色，不另挑一個灰。
  - **short label**：每一項多一個最多 4 字的 `short`（ctx、5h、7d…），Format 選 short 時用；catalog 沒有（舊版 mod）就退回 label。預設開的是 Model、Context、Session、Weekly、Cost、Git。
  - TUI 的 `[1]` 多一組 STATUSLINE（排在 DRAWER 與 SETTINGS 中間），頁面依序是 Statusline 開關、Items、Order、Format、Color、Separator、Refresh；底下的 status bar 改說「the statusline follows within N seconds」。
- **drawer 的開關（2026-10-15）。** user 不想每次開 session 都先冒出來再手動關掉，sidebar、inline 兩頁最上面各有一個 Drawer off／on（跟 Statusline 頁的排法一樣），寫成 config 的 `drawer.sidebarOff`、`drawer.inlineOff`（預設開，開的時候不寫）。0.1.1 原本是兩頁共用一個 `drawer.off`，user 回報這是 bug：要的是寬的時候有 sidebar、窄的時候不要又高又擠的 inline（改看狀態列的 on when inline）；`off` 照樣讀得到，當成兩邊都關。
  - **放哪裡由引擎決定**，2026-10-15 在 Claude Code 2.1.300、fullscreen 下量的：寬的時候沒人要求的 open 放在 dock；dock 著的 pane 縮窄會**變成 inline**（user 碰到的就是這個）；窄的時候沒人要求的 open 不畫、等變寬（`isPlaced` false）；`/ccdrawu` 這種要求的 open 窄的時候放 inline；inline 什麼都不畫還是會留一個上下兩列的框。所以「某個位置關掉」只能把 pane 關掉，不能畫空的。
  - **做法**（`keepPlace`）：pane 每次畫到新的位置就檢查，那個位置關了就關掉 pane；是 inline 而且 fullscreen、sidebar 又開著，就馬上再沒人要求地 open 一次，引擎會讓它等到夠寬再放進 dock。sidebar 關掉時不重開（會再放進 dock）。所以 fullscreen 下 inline 不會自己出現，只有 `/ccdrawu` 或非 fullscreen 才會。「在不在這個位置」由 `drawerWanted` 決定：有 `/ccdrawu` 設的手動狀態就照它，沒有才看設定（見下面的 `/ccdrawu`）。實測：sidebar 開、inline 關，寬→dock、窄→收起來（狀態列出現）、再寬→回 dock。
  - 兩邊都關時 session 開始不開 pane；TUI 切換時開著的 session 一秒內跟著開、關或搬（只在讀到的值變了的那一次動作，第一次讀只記下來，reload 不會把手動開的關掉）。session.start 先 `await` 讀完 config 才決定要不要開。
  - **窄的時候把 inline 設回 on，drawer 不會回來**（user 回報）：inline 關掉時 pane 已經關掉再沒人要求地 open，在等變寬；引擎不會把沒人要求的 pane 放進 inline（只有使用者 open 的、或從 dock 被縮窄的才會），而 TUI 改設定不算使用者要求，mod API 也沒有別的辦法（`focus` 明寫不算）。所以只做提示：inline 從關變開、一秒後 pane 開著卻沒放上去，就 toast 一句「Type /ccdrawu to show the drawer above the prompt」（`hintInline`；原本把原因也寫進去，user 說太難懂，原因留給 README）。寬的時候 pane 會進 dock，不跳。README 的限制寫了。
  - `/ccdrawu` 遇到開著但還沒畫出來（等變寬）的 pane 當作關著：再要求一次 open 就畫出來。原本要按兩次，第一次是把等著的 pane 關掉。
  - **`/ccdrawu` 是這個 session 的手動開關（2026-10-16，user 定）。** 原本只在打開的那個位置有效（`asked`），位置一變就照設定走；關掉則是整個 pane 關掉、哪裡都不回來。user 回報：inline 開著時用 `/ccdrawu` 關掉，拉寬後 sidebar 也不出現。討論了兩種：①`/ccdrawu` 只切當下這個位置，位置一變就回到設定；②`/ccdrawu` 設定整個 session 的開或關，蓋過設定。選②：①在「sidebar 關掉後縮窄、照設定要出現在輸入框上方」這條路做不到（引擎不把沒人要求的 pane 放進 inline），規則會有例外；②每一條路都做得到。手動狀態存在 `$.state` 的 `manual`（`'on' | 'off' | null`，reload 後還在、新 session 是 null），`drawerWanted(config, place, manual)` 有手動狀態就照它；session.start、`keepPlace` 都看它。TUI 改了 Drawer（`refreshConfig` 讀到開關變了）就清回 null，以最後一個動作為準（user 定）。沙盒實測：兩邊都開，窄時 `/ccdrawu` 關→拉寬、再縮窄都不出現→`/ccdrawu` 開→寬在 dock、窄在 inline；inline 關，窄時 `/ccdrawu` 開→寬 dock→窄仍在 inline→關了拉寬也不出現→config 改 inline 開→寬 dock、窄 inline。
- **`[2]` 的欄位標題（2026-10-15）。** user 要每頁都有：drawer 與 statusline 的頁是 SETTING／VALUE，preferences 是 KEY／VALUE（terminu-go 的 `Values.Header`）。user 要標題用 Overlay0，但 tdp v0.6.11 規定標題是 Blue，跟 label 同色，表單上看不出是標題：交給 terminu，定成 tdp v0.6.12（一列一個值的 panel 表頭用 Overlay0，清單維持 Blue），terminu-go v0.1.11 照做，ccdrawu 從 v0.1.7 直接升上去，沒改程式。
- **也給 Windows 用，建置照 kbu（2026-10-14）。** Claude Code 三個平台都有，家族其他 app 是 unix-first，只有 kbu 出 Windows 版。goreleaser 多出 windows 的 amd64、arm64（zip）、發到 Scoop（`vulcanshen/scoop-bucket`），`install.ps1`／`uninstall.ps1` 照 kbu（裝到 `%LOCALAPPDATA%\ccdrawu`、加進使用者的 PATH）；release 的測試也在 windows-latest 跑（那邊沒有 cgo，不加 `-race`）。位置照 tdp E3：Windows 上 `~` 就是 `%USERPROFILE%`，不用 `%APPDATA%`（tdp v0.6.4，ccdrawu 提的，user 選的）。mod 先讀 `USERPROFILE` 再讀 `HOME`，跟 Go 的 `os.UserHomeDir` 同一個順序，兩邊才會算出同一個資料夾。
- **先裝 ccdrawu，再 `ccdrawu install claude`（2026-10-14）。** 入口是這支程式，外掛由它裝（user 定的）：`claude plugin install ccdrawu --marketplace <來源> --json`（user scope；市集還沒加會先加）；已經裝了就改成更新：`claude plugin marketplace update ccdrawu`，再 `claude plugin update ccdrawu@ccdrawu`（0.1.2 起；原本是沒變化，`brew upgrade` 後外掛停在舊版，新 TUI 寫的設定舊 mod 不認得）。2026-10-16 看 user 的 `known_marketplaces.json`：官方 marketplace 的 `lastUpdated` 是當天，第三方（GitHub、git）的都停在幾週到幾個月前，Claude Code 不會自己更新它們。外掛固定在執行檔自己的 release tag（`SourceFrom`：`https://github.com/vulcanshen/ccdrawu.git#v<版本>`，dev build 取 main）：原本用 `vulcanshen/ccdrawu`，裝到的是 main 當下的內容，push 到 main 的 mod 改動使用者馬上拿到、跟執行檔不同步（0.1.1 發版後 main 上就已經是還沒發的 0.1.2 mod）。`owner/repo#ref` 也行，但它走 SSH clone，沒有 GitHub SSH key 的人會失敗，所以用 HTTPS 的 git URL。已裝的來源跟要的不一樣（舊版的 tag、main）就 `marketplace remove`（Claude Code 會連外掛一起移除）再 `install`：同名、不同來源的 marketplace 加不進去（它記在 settings.json 的 `extraKnownMarketplaces`）。README 寫明外掛只用 `ccdrawu install claude` 裝與更新，不要手動從 marketplace 裝。開發時用 `make plugin`（`CCDRAWU__MARKETPLACE` 指著 repo 資料夾），否則 dev build 的 `install claude` 會把外掛換成 GitHub main 的。用沙盒（`CLAUDE_CONFIG_DIR`）實際跑過：0.1.1 的執行檔全新安裝→外掛 0.1.1（`#v0.1.1`）；換 0.1.2 的執行檔→移到 `#v0.1.2`、外掛 0.1.2；再跑一次→`up_to_date`；原本從 main 裝的→移到 `#v0.1.2`，來源預設 `vulcanshen/ccdrawu`，`CCDRAWU__MARKETPLACE` 改指別處（開發時指本機資料夾）。`uninstall claude` 先 `plugin uninstall <id>` 再 `marketplace remove <市集名>`（市集只有 ccdrawu），設定與資料留著、告訴使用者在哪；狀態列是 ccdrawu 的就換回原本的（主題也是），外掛已經不在 Claude Code 裡時也一樣（0.1.1 在這裡直接結束，留下指著 ccdrawu 的 `statusLine`，0.1.2 修掉）。移除腳本刪執行檔前先跑一次 `uninstall claude`，`statusLine` 才不會指著刪掉的執行檔。`claude` 這個目標先佔著位置，之後要裝到別的地方再加。
- **`ccdrawu update`（2026-10-16，0.1.3）。** user 要能一個指令更新：照安裝方式更新執行檔，再讓外掛跟上。家族沒有先例、tdp 沒有規範。`internal/update`：先查 GitHub API 的 latest（跟安裝腳本同一個來源），不比目前新就不動執行檔；方式由執行檔的實際路徑（`os.Executable`＋`EvalSymlinks`）判斷（`How`）：在 `Cellar/ccdrawu` 底下是 Homebrew，跑 `brew upgrade vulcanshen/tap/ccdrawu`，新版是 prefix 的 `bin/ccdrawu`（`BrewBin`；Cellar 路徑是舊版的）；在 `scoop/apps/ccdrawu` 底下是 Scoop，只印出要跑的 `scoop update ccdrawu` 與 `ccdrawu install claude`：Scoop 更新前會檢查 app 有沒有在執行（它的 `test_running_process`），而這時執行中的正是 ccdrawu 自己（依 Scoop 原始碼的行為，沒在 Windows 實測）；其他（安裝腳本或手動下載）就自己換：下載 release 的 archive 與 `checksums.txt`、比 sha256、取出執行檔、寫在同一個資料夾的暫存檔再 rename 蓋過去（`Replace`）；Windows 不能覆蓋執行中的 exe 但能改名，所以先把它改成 `.old`、新的再放上去，`.old` 下次 update 時刪（rename 流程有單元測試，在 CI 的 Windows 上也會跑，但沒在 Windows 上實際換過執行中的 exe）。資料夾不能寫入時在下載前就停、說明原因，不自己 sudo（user 定的）。dev build 不更新，請它在原處 `git pull && make install`。執行檔換好後，只有 Claude Code 已經有這個外掛時，才用**新的**執行檔跑 `install claude`（新版決定外掛釘在哪個 tag）；外掛被拿掉的不會被裝回去。實測（scratchpad 沙盒、`CLAUDE_CONFIG_DIR` 指到別處）：標成 0.1.1 的執行檔裝好 0.1.1 外掛→`update` 下載真正的 v0.1.2、驗證後換掉自己、外掛移到 0.1.2；已是最新→不動執行檔、外掛 up to date；Claude Code 沒有外掛→只提示；資料夾 555→下載前停下；dev build→提示；Homebrew 用假的 `brew` 與假的 Cellar 驗了判斷、參數、`BrewBin` 與用新版裝外掛（沒動 user 真正的 Homebrew）。0.1.2 以前沒有這個指令，README 的 Update 留了手動的步驟。
- **不顯示 vim 的 INSERT／NORMAL（2026-10-14）。** mod API 拿不到 vim 的即時狀態，唯一的來源是 Claude Code 跑 `statusLine` 指令時給的 JSON：先做過外掛附的 sh 轉接（要 sh 與 jq），再改成 `ccdrawu bridge` 子指令（Windows 也能用）。但 `statusLine` 只能填一個指令，佔用它就會跟使用者自己的狀態列（例如 ccstatusline）互相蓋掉；而 Claude Code 輸入框下方本來就顯示 vim 模式。所以整列（原本叫 Editor，後來叫 Vim）連同轉接一起拿掉。後來狀態列做成使用者自己打開、先問過才換掉的功能（下一條），vim 模式仍然不顯示。
- **不畫底色（2026-10-14）。** 停靠的 pane 外框（上方 × 那列、左邊分隔欄、下方）由引擎的容器畫，底色是主題的 `composerSidebarBackground`（`dark-ansi` 是 `ansi:blackBright`，`dark` 是 `rgb(38, 38, 38)`）。mod 只能畫 body，蓋不到外框；只蓋 body 會變成「外框一個色、內容一個色」。所以底色交給主題，由 TUI 的 theme／background 寫自訂主題（下一條）。這個 token 不在官方文件的清單裡，是從 Claude Code 程式找到的，改名時主題會忽略它，退回灰底。
- **主題由 TUI 寫進 Claude Code（2026-10-14）。** user 選的做法：選一個內建主題當底，加上側邊欄底色。`internal/claude` 寫 `<Claude 資料夾>/themes/ccdrawu.json`（`CLAUDE_CONFIG_DIR`，否則 `~/.claude`），把 `settings.json` 的 `theme` 換成 `custom:ccdrawu`，換之前的值記在 ccdrawu 的 config.yaml（`theme.before`），選 off 或 `uninstall claude` 時換回去；使用者之後在 `/theme` 自己換過的就不動。改 `settings.json` 只改最外層 `theme` 那一個值，其他位元組原樣（`json.Decoder` 的 offset，測試守著「換過去再換回來，檔案一模一樣」），用暫存檔 rename、保留權限。否決：交給 mod 用 `$.config.set` 換——2026-10-14 實測 theme 那一列是 `choice`，只收 `auto` 與 6 個內建主題，`custom:…` 會被拒絕。內建主題清單（`claude.Themes`）是從 Claude Code 2.1.299 的執行檔裡找出來的。2026-10-14 實測：TUI 改了 `settings.json` 的 `theme` 與主題檔，開著的 session 立刻換色，不用重開。
- **寵物 pet（2026-10-15）。** Claude Code 的 `/buddy` 已經移除，user 要 ccdrawu 自己養一隻：叫 pet（不叫 buddy、不抽卡），預設開、是貓，放在 sidebar 最上面（台詞 1 列＋走道），inline 佔最左一欄（寬 14，不畫台詞與球）。圖是自己畫的：vscode-pets 的圖是 GIF、授權屬於各個畫師，只參考玩法。
  - **畫法**：每格 12×8 px，用 `▀▄` 半格字元畫成 12 欄×4 列；透明的像素是空白 Box，透出 drawer 的底色。5 種動物（`SPECIES`：cat dog rabbit duck penguin）× 7 格（walk1 walk2 sit sleep lie swipe1 swipe2）。記號：`#` 毛、`s` 斑紋／耳朵、`w` 胸腹、`x` 賓士貓多的白、`e` 眼、`n` 鼻、`p` 粉紅、`b` 喙與鴨／企鵝的腳、`q` 四肢（貓狗兔的四隻腳、鴨的翅膀、企鵝的鰭）。每組花色都明訂 `q` 的顏色，所有姿勢一開始就一起上色（user 要的；測試守著每格都有 `q`、`q` 跟 `#` 不同色、每個記號在每組花色都有顏色）。花色在 session.start 隨機抽，不給設定（user 定的）。
  - **否決**：ASCII（user 算過格數，像素佔的差不多、好看得多）；另外畫的跳格（看起來扁，改成 walk2 整隻抬高 1→2→1 px）；爬牆（另畫的攀爬格讓動物變小、用揮爪格抬高又很怪，三種畫法都否決，整個拿掉）；只有揮爪時手才換色（四肢一開始就要不同色）；從毛色自動算四肢色（每組花色明訂）；Raster＋blit（背景色不保證跟 drawer 一樣）。
  - **行為**是純函式（`hooks/pets.ts` 的 `petEvent`、`petTick`，亂數由參數注入，測試直接跑），`$.clock.every(200)` 走一步；沒變化就回同一個物件，不寫 state、不重畫。事件：`turn.step`（主對話）→ 工作中，閒逛、偶爾休息、跳、揮爪、跑；`turn.complete` 依 `reason`：answer 坐下說完成的話、aborted 安靜坐下、error／refusal 跳一下說擔心的話；`classic.PostToolUseFailure`（不是 interrupt）跳；`session.compact` 說一句；閒著 2 分鐘趴、10 分鐘睡（z 用文字 `z`→`zz`→`zzZ`，user 說不用像素畫）。台詞固定，不叫模型、不花 token（user 定的）。`/ccdrawu pat`、`/ccdrawu ball` 從 `command.run` 的 `e.args` 來，其他參數照舊是開關 drawer。
  - **球**從寵物遠的那一側滾進來（1.2 秒減速、彈 2,1,0,1,0 px），停在中段；寵物跑到球的近側 5 格（`BALL_REACH`，畫面上剛好碰到球），等球停了揮爪 1.6 秒，球消失、說一句；20 秒沒拿到就自己消失。球是 `position="absolute"` 的 Box，疊在走道上不佔位置，位置跟寵物一樣依寬度比例換算（`along()`）。
  - **框固定寬高、每列 `height={1}`**：只有空白 Box 的列會被排版壓成 0 高，睡覺（頭頂空 2 px）或跳的時候框會變高變矮，下面的清單就跟著跳。
  - **舊 state 要補齊**：`/reload-plugins` 後 `$.state` 保留舊值，session.start 用 `{ ...PET_START, ...old }` 補上新加的欄位，否則讀到 undefined 會算出 NaN。
  - **小隻 kits（2026-10-16，user 的點子）**：main agent 等 subagent 時主寵物坐著不動，背景的工作卻還在跑。現在每個跑著的 subagent 會有一隻同種的小隻，在主寵物後面走。user 定的範圍：
    - 只算 subagent。來源跟 Tasks 一樣：`$.agent.list()` 裡 `pending`、`running`、`waiting` 的，每 5 秒、`agent.spawn`、每輪結束時更新。背景的 shell、monitor 只在回覆結束時才知道，開著的 dev server 會讓一隻永遠留著，所以不算。
    - 最多三隻。只畫在 sidebar，輸入框上方的寵物欄只有 14 欄寬，沒空間。
    - 主寵物照原本的規則動。main agent 也在工作時，主寵物和小隻一起走。
    - 小隻是 6×4 px（6 欄×2 列），兔子多兩 px 給耳朵（6×6，不然看起來像老鼠），只畫 walk1、walk2（`KITS`），站在 6 px 高的框底部。原本做 8×6，user 實際看了覺得太大：走道只有 30 欄，三隻 8 欄的小隻就佔掉 24 欄。
    - 顏色：用該物種真正的花色、正常上色，每隻進場時從主寵物和其他小隻都沒用的花色裡抽一個（`kitsFor` 的 `coat`），所以四隻（主寵物加三隻）一定不同色。和主寵物的區隔靠體型、在後面、花色不同。狗、兔、鴨各補一組花色，企鵝補兩組（主寵物也會抽到），每種動物都至少 4 組，測試守著。經過四版：
      - 用主寵物的花色調暗 0.7：user 一開始就說要跟主寵物有區隔，不然畫面太花。
      - 一律灰階，亮度壓在 64–192（單純調暗的話，深色花色在深色底上幾乎看不見）：兩隻疊在一起會「變怪獸」，分不出是兩隻。
      - 灰色混入 Catppuccin 的藍、橘、綠，每隻一個色調：user 看了說暗色系「跟殭屍一樣」，在後面也要認真上色。
      - 真正的花色，每隻都跟主寵物、跟彼此不同（現在的做法）。每一版都先用 SVG 轉 PNG 預覽：30 欄的走道上，主寵物在中間、三隻小隻在兩邊和後面。
    - 狀態在 `$.state` 的 `kits` 陣列（`Kit`），跟主寵物分開。`kitsFor` 依 subagent 的 id 補進新的小隻（從走道兩端之一進來），也把結束的標成 `leaving`；`kitsTick` 跟主寵物共用 200ms 的 tick 走一步，到了目的地偶爾停一下再換下一個點，`leaving` 的走到近的那一端就移除。都是純函式，有測試。
    - 畫法：小隻是走道上 `position="absolute"` 的 Box，在 children 裡排在主寵物之前，所以主寵物畫在它們上面。主寵物透明的格子（空白 Box）不會蓋掉後面的小隻。這點 2026-10-16 在 tmux 沙盒裡用 3 個背景 subagent（各跑 `sleep 45`）實測過：`capture-pane -e` 逐格看顏色，重疊處是主寵物的彩色，主寵物透明的格子看得到後面灰色的小隻；subagent 結束後小隻都消失了。小隻的 `kits` 只在 sidebar 那條路徑讀，所以小隻走動不會讓輸入框上方的版面重畫。
  - **物種在 TUI 設**：preferences 的 Pet（radio，畫在原地），放在 Bar 與 Plugin 之間，寫進 config.yaml 的 `pet.species`。Go 的 `config.PetSpecies` 要跟 `SPECIES` 同順序；mod 不認得的物種畫成貓。
  - **不給取名（2026-10-15）**：做過 Pet name（最多 16 字），user 拿掉：花色每個 session 隨機，每次開都不是同一隻，給名字很怪；留給之後的版本再看。台詞那一列只有台詞。
- **不用 Fragment。** 這個 surface 把 `<>…</>` 當直排容器排版，放在 row 裡的「·」和 effort 會被疊成上下兩行。同一列的元素都直接放進 `flexDirection="row"` 的 Box。
- **bar 寬度要扣 paddingX。** 外層 Box 左右各 1 格 padding，`bodyColumns` 沒扣掉的話，bar 會超出寬度、折到下一行，看起來像兩條 bar。
- **bar 用 Nerd Font 的方塊（2026-10-15）。** 填滿的是 `nf-oct-square_fill`（U+F445），空的是 `nf-oct-square`（U+F51D），user 挑的。原本用 `▰▱`：很多程式字型沒有這兩個字（Maple Mono、JetBrains Mono、Fira Code 都沒有），終端機只能跟別的字型借，借到哪一套看那個終端機開啟後先出現過哪些缺的字（Ghostty 先找已經借過的字型）。2026-10-15 user 的 Ghostty 開機後第一個缺的字就是 bar，借到 Apple Symbols，字比一格寬，整串黏成一塊；前一天先出現過 Claude Code 的轉圈符號 `✻`，借到的是 Menlo，畫出來是正常的小平行四邊形。程式、字型、終端機都沒變，只差在出現順序。Nerd Font 的字在字型自己裡面，不用借；Maple Mono NF CN 與 JetBrains、Hack 的 Mono 版，方塊寬 10、一格 12，左右分開，高度只佔一列的中段，上下兩列也分開。代價是 drawer 也要 Nerd Font（README 的需求寫了）。終端機不會告訴程式它用什麼字型，mod API 也沒有字型或終端機能力的資訊（2.1.300 的型別檔查過），查電腦裝了哪些字型也不準（裝了不一定在用；沒用 Nerd Font 也可能從 Symbols Nerd Font 借到），所以不自動偵測：preferences 的 **Bar** 讓使用者選 squares（預設，user 定的）或 line（`━`，框線字元，等寬字型都有），每個選項旁邊畫著自己的 bar，畫不出來的那個一看就知道；寫進 config.yaml 的 `bar`，mod 不認得的值畫方塊。另外，非 Mono 版的 Nerd Font（例如 MesloLGS NF）方塊比一格寬（-0.8～13.65），有的終端機上可能又黏在一起。README 的範例用 `■□` 示意，因為 GitHub 網頁沒有 Nerd Font。否決：`█░`（整格高，上下兩列的 bar 會黏在一起）；請使用者把 `▰▱` 指定給 Menlo（只有 macOS 有，ccdrawu 也支援 Windows、Linux）。
- **key 欄寬 9。** key 最多 8 個字，加一格空白；只有 Directory 是 9 個字（user 定的名字，跟 Path／Working directory 對應），所以欄寬照目前顯示的最長 key 算：沒顯示 Directory 時跟原本一樣，顯示時才寬一格。測試守著其他 key 都在 8 個字以內。
- **設定了的項目一定列出，沒有值時顯示 `-`（2026-10-16，user 定）。** 原本沒有值的項目整列不畫（`view` 回 undefined）。新 session 的 Mode 要到第一個 prompt 才拿到 permission mode，那時這列才冒出來，下面的列被往下推。user 定的規則：設定裡有的項目一律列出，值拿到以前顯示 `-`。現在 `view` 只有 pet（pane 自己畫）和 catalog 沒有的 id 會回 undefined；sidebar、輸入框上方、狀態列都照這個規則，狀態列包括 `ccdrawu statusline` 的 `Line`，session 檔還沒寫出來時的 fallback 也一樣。Model、Uptime 拿到值以前原本顯示 `…`，也改成 `-`。否決：`…`，user 覺得意思不清楚。不用 `N/A`：字面意思是「不適用」，但大多數情況其實是「還沒拿到」。Session、Weekly、Spend 的 resets 列在 sidebar 上一樣先留著，顯示 `resets -`（user 定）。用量上限要等第一個 API 回應才有，型別檔寫 rateLimits 是「上一個 API 回應報的」。拿到上限但沒有重置時間時也顯示 `-`；輸入框上方和狀態列只有一列，不用預留位置，所以那裡不顯示 `· -`。Git 的增減行數還是拿到值才出現。
- **拿掉 Account（2026-10-16，user 定）。** Account 原本讀 `$.session.identity()` 的 `email`。2.1.300 的型別檔寫明，identity 只回答組織 managed settings 裝的 mod，其他 mod 收到 `{ deny: 原因 }`；`SessionIdentity` 是 union，不是 reject。一開始以為會 reject，判斷錯了，所以 user 的 session 一直顯示 `-`。ccdrawu 只用 `ccdrawu install claude` 自己裝，所以每個人都拿不到。中間試過兩種做法：拿不到時顯示 `api unavailable`；在 TUI 的項目選單把它停用（tdp v0.6.13 定了停用選項的樣子，terminu-go 在 v0.1.12 加 `Checks.Disabled`）。結果卡住：照 v0.6.13，勾著的停用選項 cursor 碰不到，已經勾的人取消不了。user 認為拿不到值的項目不該出現，所以整個拿掉，等 API 開放再加回來。config 裡還列著 `account` 的照舊讀得進來，mod（`enabledItems`）與狀態列都略過 catalog 沒有的 id。
- **會擋住引擎動作的 hook 都有 `.catch`。** `agent.spawn`、`session.compact`、`classic.*` 事件都是 gating hook；側邊欄出錯時 `.catch(($, e, next) => next(e))` 讓引擎照常進行（`next` 已呼叫過就沿用那次結果，不會重跑）。
- **跟上 tdp v0.7.4 與 terminu-go v0.1.13（2026-10-18）。** 照 terminu 的交代整理，用 tdp 的 checklist 對過一遍：
  - repo 照 tdp project 的範本（`family.mk`、release、goreleaser、install／uninstall、`internal/version`、`.gitignore`），出 Windows 版的段落打開。
  - **Claude theme 改成 select**：off 加 6 個內建主題共 7 個，tdp components/input/radio 規定五個以內才畫在原地，超過用 select。
    值寫選中的那一個（`off` 也是選項）。
  - **Refresh 改成 number**（tdp components/input/number：秒數用 number，說明列寫範圍 `1-60`，只收得進數字）；原本是 text 加檢查。
  - **Drawer background 寫成色塊加 hex**（tdp components/input/color-picker 的「值」，`terminu.ColorText`）。
  - **值出錯時 label 轉 Red**（tdp components/layout/panel，`Field.Err`）：手改 config.yaml 留下 mod 不收的 alias、不是 hex 的
    background、超過 60 的 refresh、有控制字元的 separator，以及 Plugin 查不出來時。值裡的控制字元，lib 畫成紅色的 `\x1b`，不會進終端機
    （terminu-go v0.2.0 之前 ccdrawu 自己轉）。
  - **存不進去就改回去**（`commit`、`revert`）：寫檔或套用主題失敗時，`d.cfg` 與畫在原地的 radio（`d.radios`，每個記著怎麼從
    `d.cfg` 算出選哪一個）改回檔案裡的樣子，畫面不會假裝存好。從 input 改的走 lib 的 `Commit`（terminu-go v0.2.0）：
    失敗時 input 留著、值還在，error popup 蓋在它上面；radio 這種畫在原地的照舊開 error popup。
  - **error popup 的標題寫哪件事失敗，不放名稱**：lib 把整個標題轉大寫，`Claude Code` 會變成 `CLAUDE CODE`（tdp components 的
    「大寫」：名稱照原樣）。Install failed、Uninstall failed、Can't read the settings。Alias 的錯誤點名 Alias（tdp components/input/README）。
  - **`[2]` 的 Enter 是 item operation**（tdp M3、tdp K6）：terminu-go v0.2.0 起 `Values` 自己把 Enter 列進 Space menu 與 `?`；
    之前 ccdrawu 自己登記一個轉給 `Values.Key` 的動作，已經拿掉。
  - **Refresh 帶單位**：panel 上寫 `10 seconds`，input 的值下面一列灰字寫 `seconds`（`NumberInput.Unit`）。
  - **畫面測試**：`TestL4EveryRowFullWidth`（`internal/ui/screen_test.go`）在三種 icon 與 powerline 寬度下，開 ccdrawu 的每一種
    popup，在 200x50 與最低尺寸 40x12 各開一次，200x50 開的再拉到比它窄、矮的尺寸，每一列都要剛好等於終端機寬（tdp L4、tdp E5、tdp E6）。
    popup 的動畫是真的時間，整個跑完約 30 秒。
- **跟上 terminu-go v0.2.0（tdp v0.8.1，2026-10-19），之後是 v0.2.1（tdp v0.8.2）、v0.2.2（tdp v0.8.3），ccdrawu 都不用改。** 等 lib 的定案都做完：input 的 `Commit`、`Values` 的空狀態
  （沒有 catalog 的頁用 Overlay0 的事實與提示，`Failure` 只留給讀不了 catalog）、`NumberInput.Unit`、Success toast（套用
  主題、打開或關掉狀態列、安裝或移除外掛；搬家照 tdp E3 用 Info）、最低尺寸；值沒變不寫回、Enter 列進 Space menu、
  控制字元改由 lib 做，ccdrawu 的暫時做法拿掉。reorder 的 `h|l`、`?` 只列動得了的鍵、color picker 的空值、捲動位置、
  窄的時候的收法都是 lib 做的，ccdrawu 不用改。
- **最低尺寸 40×12（2026-10-19，user 定，tdp L1）。** tdp L1 給每個 app 一個門檻，預設 80×40，比它小時只顯示 app 名稱與
  `needs a terminal of at least W×H`；app 可以調低。ccdrawu 調到 40×12：
  - 它是 Claude Code 旁邊的小工具，常開在 tmux 分割出來的窄窗格（YouTube 那幾章就是），80×40 會讓它在那裡打不開。
  - 畫面只有一個側欄加一頁值，60 欄以下本來就只畫一個 panel（`NarrowBelow`）；40 欄還放得下最長的 label（DRAWER BACKGROUND，17 格）、
    兩格、值的開頭，12 列放得下最上面一列、footer、框線與一頁的前幾欄，其他照常捲。
  - popup 照 tdp F7 的寬度與高度，在最低尺寸 40×12 本身開過每一種（`TestL4EveryRowFullWidth`）。
  - 設在 `App.MinWidth`、`App.MinHeight`（terminu-go v0.2.0）；`TestBelowTheLeastSize` 量門檻以下到 1×1：滿寬、不當掉、畫出名稱，
    拉大後回到原本的畫面。

- **設定檔是 config.yaml，catalog 與 sessions 是快取（2026-10-19，tdp E3，tdp v0.8.0）。**
  - **config.yaml**（tdp E3：設定目錄裡的 `config.yaml`）：Go 照舊用 JSON 的欄位與規則（`omitempty`、沒設的清單寫
    `null`、`Kept` 存成字串），寫的時候 JSON 經過 gopkg.in/yaml.v3 的 Node 轉成 block 的 YAML（`config.WriteConfig`），讀的時候
    反過來（`jsonFromYAML`），欄位順序跟著 struct。mod 沒有 YAML 的套件，`hooks/yaml.ts` 是一個小的讀取器：ccdrawu 寫的樣子
    （block、需要時才加引號、多列的 `|-`）加上手寫常見的寫法（flow 的 `[..]`、`{..}`，所以 JSON 也讀得了、註解、兩種引號、
    `|`、`>`）；anchor、tag、第二份文件不讀，跟讀不懂一樣當作還沒寫好、下一秒再讀。
  - **catalog 與 sessions 是快取**（tdp E3：刪了 app 會自己重建的是快取；sessions 是推得出來的，catalog 也照同一條
    規則）：mod 每次開始都重寫 catalog，sessions 每秒寫，刪了都會回來，所以搬出 `~/.ccdrawu`、改用 `CCDRAWU__CACHE`
    （`~/.ccdrawu` 後來只放原值，見下面的「原值是資料」）。位置照 Go 的 `os.UserCacheDir`（macOS `~/Library/Caches`、Linux 絕對路徑的 `$XDG_CACHE_HOME` 或
    `~/.cache`；Windows 見下面）；mod 沒有平台資訊，有 `LOCALAPPDATA` 就是 Windows，其他跑一次 `uname -s`，`darwin`
    就是 macOS，兩邊才會算出同一個資料夾。
  - **搬法**（tdp E3 的「換了位置時」）：`config.Move` 在 TUI 開始時跑，新位置沒有時把 `config.json` 寫成
    `config.yaml`、把 `~/.ccdrawu` 裡的東西搬進快取目錄，跳一個 Info toast；搬不動開 error popup 寫兩個路徑。在那之前（或搬不動時）
    都照舊讀舊的：Go 的 `ReadConfig`、`Locations.Catalog`，`ccdrawu statusline` 找不到新的 session 檔時也看舊的，mod 沒有
    config.yaml 時讀 config.json。`WriteConfig` 第一次寫出 config.yaml 時刪掉旁邊的 config.json（`install`／`uninstall claude` 寫設定時
    也算搬完）。
  - **新位置已經有時**（tdp E3-16）：config.json 與 config.yaml 都在時，config.json 留著不刪，讀 config.yaml，
    每次啟動都開 error popup 寫兩個路徑，請使用者自己合併、刪掉 config.json；不能默默留著。其他照常搬（originals 從
    config.yaml 搬、`~/.ccdrawu` 的快取）。
  - **錯誤一個一列**：error popup 裡的 `\n` 會照控制字元畫成紅色的 `\n`（tdp components 的「控制字元」），所以 app 自己的字
    不夾換行。`Move` 的每個錯誤是自己說完的句子（哪裡到哪裡、為什麼、ccdrawu 讀哪一邊、要使用者做什麼），用 `errors.Join`
    接起來，畫面那邊 `rows` 拆開、一個錯誤一列交給 `Error(title, rows...)`（tdp v0.9.15 之前是用空白接成一段）。
  - **Windows 的快取放在 `%LOCALAPPDATA%\ccdrawu\cache`**（user 定的）：照 os.UserCacheDir 是
    `%LOCALAPPDATA%\ccdrawu`，那正是 install.ps1 放 ccdrawu.exe 的資料夾，刪快取會刪到程式；改放它底下的 `cache`，
    uninstall.ps1 刪整個資料夾時一起走。
  - **`Move` 逐項搬**：mod 在 session 開始時就寫新的 catalog，常常比使用者第一次開 TUI 還早，那時新位置已經在，照「整個目錄搬」
    就搬不了，`~/.ccdrawu` 會一直留著。所以一項一項看：新位置沒有的搬過去，已經有的（mod 寫的新 catalog）把舊的刪掉，那是舊的
    快取；搬完把空的 `~/.ccdrawu` 拿掉。

- **跟上 terminu-go v0.4.0（tdp v0.9.0，2026-10-20），之後是 tdp v0.9.1（同一天）。**
  - **App Log**（tdp dialog/note、Rules F5）：lib 記每個 toast 與 error popup，`!` 打開，footer 常駐 `!:log`（有沒看過的
    錯誤、警告時是 `!:N errors`、`!:N warning`）。ccdrawu 另外記自己做的事：每次存檔（`Saved <路徑>`）、Refresh 寫進
    Claude Code 的設定、啟動時發現 Claude Code 換了自己的 status line 而把 statusline 關掉、`claude plugin list` 看不出外掛
    在不在的原因（WARN）、丟掉讀不了的 catalog（WARN）。
  - **讀不了時**（tdp E3）：config.yaml（或還沒搬的 config.json，`config.Source`）讀不了時不碰它，畫面用預設開，`Start` 時
    `app.MoveAside`：`Enter` 改名成 `.bak`（已經有就 `.bak.1`……，lib 做的）後像第一次啟動一樣重來（tdp v0.9.2 改過，見下面
    v0.4.3 那一段）；`Esc` 讓 `Run` 回傳錯誤，main 印到 stderr。catalog 是快取：讀不了就刪掉、記進 App Log，mod 下次開始時重寫；頁面上的紅字 `Failure` 拿掉了。
    `ccdrawu statusline` 與 mod 也讀 config.yaml，但它們不是 app 的啟動、在 Claude Code 裡沒有畫面可以問：讀不了時
    statusline 把原因印到 stderr（Claude Code 那一列空著）、mod 照預設，等使用者開 TUI 時再問要不要搬開。
  - **上色的值用 lib 的 span**：Drawer background 的色塊是 `Styled` 回傳的 `ColorSpans`，值是 `HexText`；最上面一列的
    `Status` 回傳 `[]Span`。顏色碼不再塞進字串。
  - **計時器用 `app.Tick`**：每 2 秒重讀 catalog 的 `reload`，tdptest 的 `Advance` 才推得動。
  - **沒有 `input_tab_width`**（tdp v0.9.1 E3-11）：只有 App Log 的 app 不算用到這個 key，照預設的 4。ccdrawu 沒有自己的
    textarea、viewer、log panel、檔案預覽。v0.9.0 時加過，v0.9.1 改窄適用後拿掉，沒發版過。
  - **tdptest**：`internal/ui/tdp_test.go` 的 `TestTDP` 在 tdptest 的沙盒裡（`CCDRAWU__CONFIG`、`CCDRAWU__CACHE`、
    `CLAUDE_CONFIG_DIR` 指進去，`claude plugin list` 給 `claudeFake`）照第一次開的樣子走一遍；`TestTDPTricky` 把容易出事的值
    （顏色碼、OSC 52、換行、寬字，取自 `tdptest.Tricky`）手寫進 alias、separator、drawer background 再各走一遍。
  - **原值是資料，放 `~/.ccdrawu/originals.yaml`（user 定，tdp v0.9.1 E3）。** ccdrawu 換掉使用者自己的 Claude Code
    theme 與 statusLine 之前記下的原值，關掉時靠它放回去；刪了找不回來，tdp v0.9.1 E3 把「app 替使用者記下的原值」寫成資料
    （例子就是 ccdrawu）。
    - `config.Originals`（`Theme`、`Statusline`），`Locations.Data` 是 `CCDRAWU__DATA` 或 `~/.ccdrawu`（原本的 `OldData`），
      檔案是 `Locations.Originals()`。只有 ccdrawu 的 theme 或狀態列開著時才有東西：`WriteOriginals` 寫空的時候刪掉檔案，
      資料目錄空了也一起拿掉，沒用過的人不會多一個 `~/.ccdrawu`。
    - **搬法**（tdp E3「換了位置時」）：`Move` 在 config.json 搬完之後，把 config 的 `theme.before`、`statusline.before`
      逐項搬進 originals.yaml，toast 寫 `Moved theme.before and statusline.before in <config.yaml> to <originals.yaml>`。
      originals.yaml 已經有一樣的值，只從 config 拿掉；有不一樣的，留在 config、error popup 寫兩個路徑。先寫 originals 再寫
      config：中間斷掉時是兩邊都有，下次當作一樣的拿掉，不會兩邊都沒有。`~/.ccdrawu` 裡其他的東西（0.1.4 的 catalog、
      sessions）照舊搬去快取，originals.yaml 不動。
    - **還沒搬之前讀舊的**（tdp E3-15）：`ReadOriginals(path, cfg)` 讀 originals.yaml，config 還留著的那一項以 config 的為準
      （搬不動、或兩邊不一樣時都是 config 的）。TUI 改到原值時（`commit`）先寫 originals.yaml，再把 config 的舊位置清掉。
      `ccdrawu uninstall claude`（`putBack`）也照這樣讀、寫。
    - **讀不了時**：originals.yaml 是 app 自己的資料，跟 config.yaml 一樣在啟動時 `app.MoveAside`；搬開後當作沒有原值。
    - `ccdrawu statusline` 與 mod 都不讀原值，不用改。
    - 測試的 `Locations` 一定要給 `Data`：沒給時 `Originals()` 是相對路徑，會把 originals.yaml 寫進 package 的資料夾。

- **跟上 terminu-go v0.4.3（tdp v0.9.2，2026-10-21）。**
  - **讀不了時，`Enter` 像第一次啟動一樣重來**（tdp E3-18）。啟動要讀的東西集中在 `drawer.open`：先看 config 與 originals
    讀不讀得了，有一個讀不了就什麼都不搬（不跑 `config.Move`）、畫面用讀得到的開、一次只問一個；`Enter` 搬開後的 `reopen`
    再跑一次 `open`，搬舊位置、讀檔、檢查 Claude Code 是不是換了自己的 status line，全照第一次啟動。以前 `Enter` 只做
    `revert`，跟第一次啟動有三處不一樣：config.yaml 壞掉而旁邊有 config.json 時，config.json 留在舊位置被讀、沒有搬進
    config.yaml；originals.yaml 壞掉時，搬開後不會檢查 status line 被換掉；而且先 `Move` 再問，config.json 與 config.yaml 並存時的 error popup 會在問題
    底下說「ccdrawu 讀 config.yaml」，搬開以後就不對了。第一次啟動不會建任何檔（config.yaml 等第一次存檔才有），`Enter` 後也
    一樣。沙盒裡實際跑過：沒有檔的第一次啟動、config.yaml 壞掉、壞掉而旁邊有 config.json、originals.yaml 壞掉而 Claude
    Code 換了 status line，四種的畫面與留下的檔都對得上。
  - **`~/.ccdrawu` 裡 `originals.yaml` 開頭的都不搬**：`Move` 把資料目錄裡 originals.yaml 以外的東西當成 0.1.4 的快取搬走，
    搬開的 `originals.yaml.bak` 也會被搬進快取資料夾；現在 `originals.yaml.bak`、`.bak.1`……留在原處。
  - **App Log 每一筆大寫開頭，跟 toast 一樣**（tdp note-78）。toast 一字不改記進 App Log，所以 toast 也算。改了三處：
    `claude plugin list` 看不出外掛在不在（`Can't tell from claude plugin list whether Claude Code has the plugin: …`，指令名
    不能改大寫，所以改寫成句子）、丟掉讀不了的 catalog（`Threw away <路徑> for the plugin to write anew, as it can't be read:
    …`，原本是路徑開頭）、外掛裝好或移除後的 toast：那是 Claude Code 自己的話，`claude plugin update` 會回
    `ccdrawu is already at the latest version (…)`，開頭是外掛名，前面加 `Claude Code: `。其他 toast 都是 `Claude Code …`、
    `Moved …`、`Saved …` 開頭；error popup 記進去是 `<標題>: <內容>`，標題是大寫。MoveAside 自己記的那一筆
    （`Moved … to ….bak`）是 lib 寫的，v0.4.3 改成大寫。

- **跟上 terminu-go v0.4.4（tdp v0.9.3，2026-10-21）。**
  - MOVE ASIDE 的問句改成 `Move it to <name>.bak and start fresh?`，lib 畫的，ccdrawu 沒有測試比對問句。
  - **不再用的檔**（tdp E3-23：不刪、不跳 toast，檔還在時每次啟動在 App Log 記一筆，lib 的 `App.NoLongerUsed`）：ccdrawu
    沒有，所以沒叫。查過 0.1.0～0.1.4 寫過的檔：`~/.config/ccdrawu/config.json` 搬進 config.yaml 時 `WriteConfig` 就把它刪了，
    跟 config.yaml 並存時是 E3-16（每次啟動 error popup），不是不再用；`~/.ccdrawu` 的 catalog、sessions 是快取，不算；
    `~/.claude/themes/ccdrawu.json` 還在用；`~/.claude/ccdrawu/config.json` 是 0.1.0 之前的、沒發版過；Windows 上
    `ccdrawu update` 留下的 `ccdrawu.exe.old` 是執行檔，不是設定、狀態或資料，下次 update 會刪。之後改版不再用某個檔時，
    在 `Start` 叫 `a.NoLongerUsed(路徑)`，CHANGELOG 寫一句。
  - lib 自己送開關動畫與 toast 的計時，ccdrawu 不用改；自己的 command 照常回傳。

- **跟上 terminu-go v0.4.22（tdp v0.9.4～v0.9.16，2026-10-25）。** 這輪 terminu 沒發交代，照 tdp 的 CHANGELOG 逐版對。
  - **app 要做的**：
    - **多列的字一列一列給**（tdp v0.9.15，components-30～32）：`Move` 搬不了的一個錯誤一列（見上面「錯誤一個一列」），
      搬了的幾項在 toast 裡也一項一列（原本用 `; ` 接）。外面程式說的好幾行（`claude plugin ...` 的 stderr、看不懂的
      回答）不整段塞進短的字：Install／Uninstall 失敗的 error popup、Plugin 那一列的值只放第一個有字的那一行
      （`firstLine`；claude 是 Node 程式，原因在第一行、後面是 stack），整段另外記進 App Log（`Claude Code said: …`）。
    - **`make dev`**（tdp v0.9.4 的專案範本）：`family.mk` 照範本換過，編好再跑 `./ccdrawu`，讀的是真的設定與資料。
  - **lib 畫的，ccdrawu 不用改**：capsule 的顏色與斜角接頭（v0.9.4、v0.9.5）、模式名稱的開合動畫（v0.9.9、v0.9.10）、
    模式裡的 Space menu（v0.9.7）、有 popup 時第一個 `q`／`Ctrl-C` 先問 `Quit ccdrawu?`（v0.9.8，ccdrawu 沒有自己的
    quit flow，沙盒裡在 Space menu 上按 `q` 看過，`Esc` 回到 menu）、`Visual>select`（v0.9.13，App Log 的選取）、
    `h|l:tab`（v0.9.14，ccdrawu 沒有分頁）。Items 的 checkbox popup 第一列是 lib 加的 `Select all`、選項往內兩格
    （v0.9.11、v0.9.12，沙盒裡看過）；ccdrawu 沒有自己代表「全部」的選項。
  - **不適用**：compare 模式（v0.9.6，ccdrawu 沒有要比對的東西）；被指到的 panel（v0.9.16，`List.Points`）：[1] 的每一列
    對應的都是同一個 [2]，沒有「指出是哪一個」的資訊，所以不設。

- **跟上 terminu-go v0.5.0～v1.0.0（tdp v0.10.0～v1.0.0，2026-10-26～27）。** 每一版照 terminu、terminu-go 的交代對過，
  程式只改了引用與版號。tdp v1.0.0 是穩定版，user 在每個 app 上手動測過。
  - **引用 tdp 寫段落編號**（tdp v0.10.0）：註解與測試寫 `tdp radio.1`、`tdp panel.12` 這種段落 ID，整個元件寫檔名
    （`tdp reorder`），規則照舊寫 `tdp K4`；`terminu design E3` 改成 `tdp E3`。範本開頭的路徑、checklist 編號
    （E3-18、note-78、components-31/32）照 tdp v0.10.2 可以留著，下次改到那段程式時再換成出處（`tdp E3`、`tdp note.7`、
    `tdp components.5`）。
  - **`ccdrawu statusline` 交給 app**（tdp v0.10.0 P5）：見下面「tdp 不管的地方」。
  - **tdptest 也查 repo**（terminu-go v0.5.3）：`Check` 的子測試 `repo` 比對 tdp project.2 的範本，再查 project.3 的檔案、
    project.5 的 Makefile、project.6 的 CHANGELOG；ccdrawu 照原樣就過了。
  - **標準尺寸 60×32**（tdp v0.11.0 L1，原本 80×40）：ccdrawu 的最低尺寸 40×12、`NarrowBelow` 60 都在上限內，`Check` 改在
    60×32 跑。比最低尺寸小時，lib 多畫一列 `Press [Ctrl-C] twice to quit`。60～65 欄時 preferences 頁的 [2] 只剩 36 欄，
    Bar 的樣本與 `the theme's own` 會被截斷，66 欄以上都放得下；user 定維持現狀：側欄 24 跟 locku、webu 一樣，不改窄、
    也不換 Preferences 這個詞（側欄改成 18 欄的話，`[1] PREFERENCES` 放不下）。
  - **App Log 記的範圍**（tdp v0.11.0～v0.11.3 note.7）：使用者執行的 operation（Rules P3）都記，只看不改的也記；切換畫面由
    lib 記。ccdrawu 改設定時都記 `Saved <路徑>`，toast、error popup 由 lib 記；在 [1] 按 `Enter` 進 [2] 是換 focus，panel 的
    篩選、比較模式、按 `Esc` 取消都不記。ccdrawu 只有一個畫面，沒有 viewer，也沒有自己的篩選，所以不用改。
  - **用不到的**：`PopupActions.ErrorSubject`（v0.5.2）、PTY panel 啟動前的空狀態（panel.15，ccdrawu 沒有 PTY）、提權
    `Elevate`（tdp terminal.7）、清單旗標 `List.Flags`（tdp panel.7、panel.11）。

- **去敏：terminu-go v1.1.0-beta.1（tdp v1.1.0-beta.1，2026-10-28）。** user 把個人資訊從公開的 repo 拿掉，照 terminu 交代的做：
  - **Claude Code 的檔不進 git**（tdp project.3、project-37）：`.gitignore` 照 tdp v1.1.0-beta.1 的範本重抄（`CLAUDE.md`、
    `.claude/` 底下的都不收，只留 harness 的 `.claude/rules/terminu.md`），ccdrawu 自己的兩行（Claude Code 載入外掛時寫的
    `.claude-plugin/types/`、`tsconfig.json`）接在後面。social preview 的指令與 agent 用 `git rm --cached` 取消追蹤，檔案留在
    本機。tdptest 的 `Check` 會查，有追蹤就失敗。
  - **不指向不公開的筆記**：本檔與註解原本寫的盲點、定案編號，指的是 GitHub 上看不到的紀錄。編號都拿掉了：要用的內容本來就
    寫在旁邊，拿編號當指稱的地方改成直接寫是哪一件事。tdp 的 checklist 編號（E3-16、note-78 這類）是公開的，照舊。
  - **載入外掛不寫這台機器的路徑**：只跑一次的寫法改成在 repo 的根目錄跑 `claude --plugin-dir .`。`hooks/format.ts` 與
    `tests/ccdrawu.test.ts` 裡的 `~/sideproj/ccdrawu`、`/Users/me` 是範例值，留著。

- **第一個 beta：v1.0.1-beta.1（terminu-go v1.1.0-beta.2，tdp v1.1.0-beta.2，2026-10-28）。** user 定的：stable v1.1.0 之前，
  家族每個 app 都先發一個 beta，把發版流程完整跑過一遍；畫面與操作都沒變，所以版號只動 patch。
  - **範本**（tdp project.7、project-38、project-39）：`.goreleaser.yaml` 加 `release.prerelease: auto`，tap 與 scoop bucket 加
    `skip_upload: auto`，帶 `-beta.N` 的 tag 在 GitHub 上是 pre-release，brew、scoop、install 腳本（都拿 latest）留在 stable；
    `release.yml` 把 release notes 寫到 `$RUNNER_TEMP`、拿掉 `--skip=validate`。原本寫在 checkout 裡，工作區變 dirty，v1.0.0 的
    執行檔用 `go version -m` 看是 `+dirty`。只改範本允許的地方：`ccdrawu`、`# app:` 那幾行、Windows（bucket 的
    `skip_upload` 跟著 Windows 那段打開）。
  - **`ccdrawu update` 照 semver 比版號**（`update.Newer`）：原本照 `.` 切開逐段轉數字，`1.0.1-beta.1` 的 `1-beta` 轉不出來當 0，
    讀成 1.0.0.1。這次的 beta 剛好比對得對，但等 v1.0.1 發了，`1.0.2-beta.1` 會被當成比 1.0.1 舊，`update` 會把它「更新」回
    1.0.1。現在先比數字，同樣時正式版比 pre-release 新，pre-release 照 semver 一段一段比（數字照大小、排在字的前面，字照位元組，
    其他都一樣時段多的新），`+` 後面的 build metadata 不算。裝正式版的人不會被提示 beta：`LatestVersion` 讀 GitHub 的
    `releases/latest`，它從不回 pre-release。
  - **外掛的版號**：`plugin.json` 寫 `1.0.1-beta.1`，`claude plugin validate` 收；`ccdrawu install claude` 裝 `#v1.0.1-beta.1`
    這個 tag 上的外掛。drawer 的 Claude Code 版本列比的是 npm 的 `latest`，跟 ccdrawu 的 beta 無關。

- **v1.0.1-beta.2（terminu-go v1.1.0-beta.3，tdp v1.1.0-beta.3，2026-10-28）。** beta.1 全部通過（CI、pre-release、tap 與
  bucket 留在 1.0.0、沒有 `+dirty`）。kbu、sshu 的 beta 抓到 v2 以後的 module 路徑沒帶 `/vN`，tdp 把 goreleaser 範本的
  ldflags 改成 `-X {{ .ModulePath }}/internal/version.Version={{.Version}}`（module 路徑取自 go.mod），每個 app 都用最後的
  範本再發一次 beta。ccdrawu 還在 v1，module 路徑不用改；`.goreleaser.yaml` 的 ldflags 與上面的註解照範本抄，跟範本一字不差。

- **v1.0.1（terminu-go v1.1.0，tdp v1.1.0，2026-10-28）。** 7 個 app 的 beta.2 都把發版流程完整跑過、全部通過，tdp 與
  terminu-go 發了 v1.1.0。跟 beta.3 比，tdp 只多 `project/version.go` 範本的註解：ldflags 的路徑寫成 `<module>`（go.mod 的
  module，v2 以後帶 `/vN`），`internal/version/version.go` 照抄。CHANGELOG 的 v1.0.1 那段照 project.6 寫 v1.0.0 以來的全部
  改動，beta 的併進來；兩個 beta 的段落留著。這次會推共用的 tap 與 bucket，由 terminu 一個 app 一個 app 發。

## 已否決，不要重提

理由寫在上面的設計決定，這裡只列出來。

- 狀態列用 `$.ui.status`：畫成亮黃色、加上外掛名，ESC 與換行變 `�`。用 Claude Code 的 `statusLine` 呼叫 `ccdrawu statusline`。
- 順序清單放在 panel 上、用 panel 的 Drag mode：user 要的是 dialog，做成 reorder popup。
- 沒有值時顯示 `…` 或 `N/A`：用 `-`。
- 顯示 vim 的 INSERT／NORMAL：要佔用 `statusLine`，Claude Code 輸入框下方本來就有。
- drawer 自己畫底色：蓋不到引擎畫的外框，交給主題。
- 寵物：ASCII、另外畫的跳格、爬牆、只有揮爪時才換色的手、從毛色算四肢色、Raster＋blit；取名（花色每個 session 換，之後的版本再看）。
- bar 用 `▰▱`（很多程式字型沒有）、`█░`（上下兩列黏在一起）、請使用者把 `▰▱` 指定給 Menlo（只有 macOS）。
- Account 項目：自己裝的外掛拿不到 `$.session.identity()`；停用選項的做法也一起拿掉，等 API 開放再加回來。

## 已知的牆與未做

### 已知的牆

- **權限模式有延遲。** 沒有「權限模式改變」的事件。Claude 工作中按 `shift+tab`，等目前這一步的工具跑完才更新；閒置時要等下一個 prompt。mod API 沒有讀目前模式的方法；`statusLine` 指令收到的 JSON 也沒有：2026-10-14 在 Claude Code 2.1.299 把它錄下來看過，按 `shift+tab` 時狀態列馬上重跑，但前後只差 `cost.total_duration_ms`，沒切換時它也會這樣重跑，連「重跑了」都不能當成模式變了的訊號。
- **tmux 256 色。** Claude Code 偵測到 `$TMUX` 就把顏色限制在 256 色，hex 色會被換成最接近的 256 色（`#11111b` 會變成第 17 號深藍）。要全彩得設 `CLAUDE_CODE_TMUX_TRUECOLOR=1`。
- **mod 拿不到模型的顯示名稱。** `$.session.model()` 只給 id（`claude-opus-5-5`），所以 drawer 與狀態列都顯示 id。
- **組織管理的環境。** `session.compact`、`skill.prompt` 與 `classic.*` 事件在組織管理的 Claude Code 裡會略過使用者安裝的外掛，Mode、Compact 和 Skills 會是空的，compact 後的 Context 也要等下一輪結束才更新，寵物不會在工具失敗時跳、compact 後說話。
- **面板放哪裡由 Claude Code 決定。** `Pane` 的 `placement`（`dock`／`inline`）是唯讀的，hook 改了會被拒絕；`$.ui.open` 只能要大小（inline 的 `rows`、dock 的 `columns`），沒有位置。fullscreen（alternate screen）而且終端機夠寬（自動開的 144 欄、使用者開過的 110 欄）才 dock，其他都 inline。所以 TUI 不提供「位置」設定，`[2]` 的 sidebar／inline 是「放在那裡時顯示哪些」；能控制的只有開不開、開多大。2026-10-15 user 的 session 曾被重開成 default 模式（process 環境有 `CLAUDE_CODE_TUI_JUST_SWITCHED=default`、沒有 `CLAUDE_CODE_NO_FLICKER=1`，settings.json 有 `"tui": "default"`），`/tui fullscreen` 卻回「Already using the fullscreen renderer.」，drawer 一直在 inline；結束 Claude Code 再重開就回到 sidebar。怎麼切到 default 的沒查到（那段對話沒有 `/tui default`，壓縮過的程式也看不出 `/tui` 的判斷）。README 的疑難排解寫了「重開」。
- **Windows 的終端機（2026-10-15 user 的同事試裝）。** Windows Terminal：TUI、inline、sidebar 排版都對；沒有 Nerd Font 時缺的字畫成 `�`，仍佔一格；powerline 與框線字元由 Windows Terminal 自己畫，不靠字型。Termius 的本機終端：bar 的方塊在終端機佔兩格（從截圖量的：13 個方塊加右邊一段深色剛好 26 格碰到右邊界，剩下 6 個方塊和百分比折到下一行最左邊），sidebar 整個錯位；TUI 也留殘影，radio 標記（U+F0000 以上的 Material Design 圖示）畫成 `�`；sidebar 的淺色底只塗在有字的格子上。terminu-go 量圖示寬度（CPR）只在 unix 上做，Windows 預設一格，可以用 `CCDRAWU__ICON_WIDTH`／`TERMINU__ICON_WIDTH` 指定，沒在 Termius 試過。不另外處理，README 寫改用 Windows Terminal。同事後來在 Windows Terminal 換成 Maple Mono NF CN，只剩沒選中的 radio 是 `�`：Maple 7.0 缺 U+F043D（nf-md-radiobox_blank），是 ccdrawu 與 terminu-go v0.1.7 用到的 39 個圖示裡唯一缺的（`fc-list ":family=…:charset=…"` 逐一查；zsh 裡變數要寫 `${f}`，`$f:charset` 的 `:c` 會被當成修飾符）。JetBrains Mono、Hack、FiraCode 的 Nerd Font 都不缺，同事換 JetBrains Mono Nerd Font 後就正常。user 定：字型問題，不是規範問題，不請 tdp 換圖示（Maple 用的人少）。
- **淺色主題。** sidebar 的底色是 Claude Code 主題的 `composerSidebarBackground`（內建淺色主題是 `rgb(235, 235, 235)` 或 `rgb(245, 245, 245)`，`light-ansi` 是 `ansi:white`），配色寫死 Catppuccin Mocha，在淺色底上看不清（同事的 Claude Code 是淺色主題）。mod 的顏色可以用主題鍵（`text`、`subtle`、`success`…，跟著使用者的主題變），之後要支援淺色主題可以從這裡改。

### 未做

- 配色可設定（目前寫死 Catppuccin Mocha）
- 寵物取名（做過又拿掉，見設計決定的寵物 pet；之後的版本再看）
- Windows 還沒驗到的：rename 蓋過 mod 正在讀的檔會不會失敗（TUI 存設定靠 rename；同事試裝時 TUI 存檔成功，但沒刻意在 mod 讀檔時存）、`.ps1` 的語法（不知道同事用哪種方式裝的，這台 Mac 沒有 PowerShell）。mod 的 `$.fs` 寫 Windows 路徑已驗到：TUI 列得出 mod 寫的 catalog

## tdp 不管的地方

照 tdp v1.0.0 對過（2026-10-27）。tdp 的畫面規定只管 app 自己畫的 TUI，也就是 `ccdrawu` 這個設定程式（tdp P5）。drawer 是 mod 畫在
Claude Code 裡的 pane，狀態列是 `ccdrawu statusline` 印給 Claude Code 顯示的一列，都由 app 自己定：

- **`ccdrawu statusline` 用 `ansi.Truncate` 截斷。** 寬度是 Claude Code 給的 `COLUMNS`，不經過 terminu 的 icon 探測。tdp v0.9.16 以前
  這裡算偏離 tdp E5（所有量寬度的地方走同一個函式）；tdp v0.10.0 拿掉「偏離」，P5 把印給別的程式顯示的輸出交給 app，例子就是它。

以前算偏離、後來 tdp 定案的（tdp v0.8.0）：README 的兩張 gif（tdp E7 第 3 段：其他段可以放說明那一段的）、Drawer 的 off／on
radio（tdp components/input/radio：開關可以用兩個選項的 radio）、README 自己的段落（tdp E7）、表頭比 label 寬
（tdp components/layout/panel：label 欄寬取 label 與表頭第一欄名裡最長的）、sessions 的位置（現在是快取）。

## 設計文件導讀

ccdrawu 沒有另外的設計文件。

- 使用者看得到的：`README.md`、`README-zh_TW.md`（內容對齊）；每一版改了什麼：`CHANGELOG.md`。
- drawer 怎麼拿到每一個值、多久更新：本檔的「運作方式」；為什麼這樣做：「設計決定」。
- 設定 TUI 照的規範：terminu 的 `principle/`、`components/`、`project/`（tdp v1.1.0）。
- 程式：mod 在 `hooks/`（`register.tsx` 是入口，`items.ts` 是項目的 catalog，`pets.ts` 是寵物），設定 TUI 在 `internal/ui`。

## 建置與開發

目錄照 tdp project（`cmd/ccdrawu`、`internal/`、`docs/`、`family.mk`……）。ccdrawu 另有：

- `hooks/`：Claude Code 的 mod（TypeScript，hooks module 是 `hooks/register.tsx`，`hooks/hooks.json` 指向它）。
- `tests/`：mod 的測試（`claude plugin test .` 跑 `tests/*.test.ts`）。
- `types/`：mod 的 `$.state` 型別合約（`.claude-plugin/plugin.json` 的 `types`）。
- `.claude-plugin/`：外掛的 `plugin.json` 與市集的 `marketplace.json`；Claude Code 載入時在裡面寫 `types/`（git 不收）。
- `tsconfig.json`：Claude Code 載入外掛時寫的（git 不收），`tsc -p .` 用。
- `install.ps1`、`uninstall.ps1`：Windows 的安裝與移除（tdp project 沒有範本，照 kbu）。
- `.claude/`：git 不收，只留在維護者的機器上（tdp project.3），例如 social preview 的指令與 agent（見下面）。

### 載入

- 只跑一次：在 repo 的根目錄跑 `claude --plugin-dir .`
- 長期：`make plugin`，把這個資料夾 git 收的檔案當市集安裝：先把 git 收的檔案（`git ls-files -co --exclude-standard`，含還沒 commit、還沒 add 的）照原樣放進 `.local/plugin`，再 `CCDRAWU__MARKETPLACE=<repo>/.local/plugin ./ccdrawu install claude`。Claude Code 會把市集資料夾整個複製進外掛快取（`~/.claude/plugins/cache/ccdrawu/ccdrawu/<版本>`），不是直接讀資料夾。原本直接指著 repo，連 git 不收的 `.local` 也複製進去，一份 1.8G（demos 裡有 Claude Code 的執行檔，還有 `.local/draft` 的貼文草稿），2026-10-16 改成先挑檔案。改完要再跑一次 `make plugin`，開著的 session 再跑 `/reload-plugins`。不要直接跑 dev build 的 `ccdrawu install claude`：沒有 `CCDRAWU__MARKETPLACE` 時它會裝 GitHub main 的外掛，把這個資料夾裝的換掉。

### 設定程式（Go）

- 跟家族其他 app 同一套建置：`cmd/ccdrawu`（命令列）、`internal/config`（tdp E3 的位置、catalog 與設定檔）、`internal/ui`（TUI 與 splash）、`internal/claude`（`ccdrawu install`／`uninstall claude`）、`internal/version`（release 時由 goreleaser 的 ldflags 注入）。`make` 列出所有指令；`make run` 跑 TUI，`make build` 編出 `./ccdrawu`。
- Windows 版編不編得過：`GOOS=windows GOARCH=amd64 go build ./cmd/ccdrawu`（arm64 同），或 `make snapshot` 一次打六個平台的包到 `dist/`。`make release-check`（`goreleaser check`）要有 GitHub remote 才過。
- 本機同時改 terminu-go 時用 `go work init . ../terminu-go`（`go.work` 不進版控）。
- `internal/ui/icon.svg` 是 `docs/icon.svg` 的複本：splash 用 `go:embed` 讀 icon，讀不到上層目錄。換 icon 時兩個一起換，`TestSplashIconIsDocsIcon` 會抓不一致。

### 錄 gif

- `make gif` 錄 README 的兩張：`docs/demo.gif`（Claude Code 裡的 drawer 與輸入框下方的狀態列：問 `~/my-app` 一句話、寵物在 Claude 工作時走動、做完坐下說話、`/ccdrawu pat`、`/ccdrawu ball`）與 `docs/settings.gif`（設定畫面：Items 勾兩項、Order 用 `D` 搬 Cost、statusline 的 Format 換成 short、preferences 把寵物換成狗）。tape 都在 `.local/demos/`（不進版控）。
- demo 的狀態列：`setup.sh demo` 在沙盒的 config 打開 `statusline.on`，另寫一份 `~/demo-settings.json`（`statusLine` 指向 repo 剛 build 的 `./ccdrawu statusline`），tape 用 `claude --settings` 帶進去。這樣錄出來的狀態列不受真正的 `~/.claude/settings.json` 和 `PATH` 上是哪一版 ccdrawu 影響。
- demo 的 drawer：沙盒用 `youtube/setup.sh`（裡面先跑 `setup.sh demo`），`~/.claude` 是逐一連到真的、只有 `plugins` 空著的資料夾，tape 用 `--plugin-dir` 載這個 repo 的外掛（v1.0.0 起，2026-10-27）。drawer 的 ccdrawu 列讀的是載入的外掛的 `plugin.json`：接真的 `~/.claude` 時載的是裝好的那份（`make plugin` 的 `.local/plugin`），版號會是上次 `make plugin` 的；你裝的其他外掛也不會在畫面上留下訊息。發版時先改好 `plugin.json` 的版號再錄，tape 等的是 `ccdrawu 1.0` 這幾個字。
- settings.gif 不按 Statusline 的開關、Refresh、Claude theme、Plugin：它們寫的是 Claude Code 的設定，沙盒裡 `~/.claude` 是接到真的那一份。改了 TUI 的游標路徑就先在 tmux（獨立的 server，見「在真的 Claude Code 裡試」）用 `CLAUDE_CONFIG_DIR` 指到別處、照 tape 的按鍵走一遍，確認每次 `Enter` 落在哪裡再錄。
- 沙盒照 sshu：`.local/demos/setup.sh` 每次重做 `.local/demos/home`（全新的 ccdrawu 設定、從快取目錄（`~/Library/Caches/ccdrawu`）複製的 catalog、重建的 `my-app`：一個 commit 加一處沒 commit 的改動），tape 再 `export HOME` 過去，畫面上的路徑就是 `~/my-app`、`~/.config/ccdrawu/config.yaml`，真正的設定不會被碰。Claude Code 的設定、外掛與登入用 symlink 接真的：`~/.claude`、`~/.claude.json`，還有 `~/Library`（登入在 keychain，macOS 從 `$HOME/Library` 找；沒接會是 Not logged in，而且沙盒裡會長出一個真的 `Library`，setup.sh 會換回 symlink）。
- tape 要先 unset `XDG_CONFIG_HOME`／`XDG_DATA_HOME`／`XDG_STATE_HOME`（user 的環境有設，ccdrawu 照 tdp E3 先讀它，就會讀到真正的設定），以及從 Claude Code 裡跑 vhs 時繼承來的 `TMUX*`、`TERM_PROGRAM*`、`CLAUDECODE`、`CLAUDE_CODE_*`（留 `CLAUDE_CODE_NO_FLICKER`），否則畫面上有 tmux focus-events 與「Transcript saving is off」的提示。
- `export HOME` 之後要 `cd $HOME/my-app`：這台的 bash 5.3 換了 `HOME` 之後 `~` 還是原本的家目錄。vhs 的 `Type` 字串不收 `\"`。
- settings.gif 的 Items：清單第一列是 lib 加的 `Select all`，游標從它底下第一項開始；項目的順序跟著 catalog，catalog 加了項目就要重數 `Down` 幾下。
- demo 會開真的 Claude Code（要 fullscreen），送一個 prompt，花一點 token；問句要指定英文回答，否則照 user 的 CLAUDE.md 回繁中。`~/my-app` 第一次要先信任：`vhs .local/demos/trust.tape`（信任對話框預設是 No，要 `↓` 再 `Enter`）。
- splash 不錄進 gif：它是彩蛋，`V` 才開，tdp S2 不寫進 README。
- 寵物的動作是隨機的，錄完要看；vhs 偶爾會漏幾秒畫面（一張 gif 停格好幾秒），重錄即可。

### Social preview

- `docs/social-preview.png` 是 GitHub 的 social preview：icon 在上、下面 Tiny5 字樣 `ccdrawu`，跟家族其他專案同一個構圖。`/social-preview ccdrawu` 產生（`.claude/commands/` 與 `.claude/agents/`，從 locku 搬來，git 不收），輸出在 `/tmp`，看過再複製進 `docs/`（2026-10-15）。
- GitHub 沒有上傳 social preview 的 API：repo 的 Settings → General → Social preview 手動上傳。

### 檢查

- `make check`：gofmt、`go vet`、`go test -race`；有 `go.work` 時加 `check-lib`（關掉 `go.work`，照 go.mod 裡的 terminu-go 版本編譯、測試）；最後是 `make mod-check`（下面兩個 plugin 指令，Makefile 的 `CHECK`）。commit 前跑。家族共用的 target 在 `family.mk`（tdp project 的範本，照原樣複製、不在這裡改），`Makefile` 只設變數，ccdrawu 自己的 `plugin`、`mod-check` 放在最後一段。
- `claude plugin validate .`：讀 manifest、市集檔和 hooks module，列出 hook、呼叫的 API 和引擎會拒絕的地方。
- `claude plugin test .`：跑 `tests/*.test.ts`。
- 型別檢查：Claude Code 載入外掛後會在 `.claude-plugin/types/` 放型別檔、在根目錄放 `tsconfig.json`（兩者都在 `.gitignore`），之後 `tsc -p .` 即可。

### 在真的 Claude Code 裡試

- 沙盒：`CCDRAWU__CONFIG`、`CCDRAWU__CACHE` 指到暫存資料夾（config 寫好要試的設定），`claude --plugin-dir <repo> --settings '{"enabledPlugins":{"ccdrawu@ccdrawu":false},"statusLine":…}'`，裝好的那份關掉、狀態列只對這個 session。寬窄用 tmux 的 `resize-window` 換，畫面用 `capture-pane -p` 讀。
- **tmux 一律開獨立的 server**：`tmux -L ccdrawu-test new-session -d -s <名字> -x 200 -y 50 …`，之後的 `send-keys`、`capture-pane`、`resize-window`、`kill-session` 都帶 `-L ccdrawu-test`，用完 `tmux -L ccdrawu-test kill-server`；要完全乾淨的設定再加 `-f /dev/null`。user 的 tmux.conf 有全域的 `session-closed` hook，預設 server 上任何 session 關掉、剩下超過一個就彈出 session picker：在預設 server 開關暫存 session，user 的畫面就一直跳 popup（2026-10-16 另一個 session 回報、重現過）。獨立 server 也碰不到 user 的 locku 自動鎖定、hook 與狀態列設定。

## 發布

照 tdp project 的「發布」：`make check` 通過，CHANGELOG 的 `## Unreleased` 改成 `## vX.Y.Z — tdp vA.B.C`，commit，打 tag `vX.Y.Z`，
push commit 與 tag，`release.yml` 接著測試（含 windows-latest，不帶 `-race`）、goreleaser 發 GitHub release、推 Homebrew tap 與
Scoop bucket，再用 CHANGELOG 那一段蓋過 release 的內文。ccdrawu 多一步：

- **外掛的版本**：`.claude-plugin/plugin.json` 的 `version` 改成同一個版號，跟 CHANGELOG 同一個 commit。`ccdrawu install claude`
  裝的是跟執行檔同一版的 tag 上的外掛，drawer 的 ccdrawu 列讀的也是它。
- **beta**（tdp project.7）：tag `vX.Y.Z-beta.N`，CHANGELOG 標題 `## vX.Y.Z-beta.N — tdp vA.B.C`，步驟一樣；GitHub 上是
  pre-release，tap 與 bucket 不動。之後的正式版那一段要寫上一個正式版以來的全部改動，beta 的也算。
- `v0.1.4` 以前的標題是 `## [0.1.x] - 日期`（tdp v0.7.0 以前的寫法），照 tdp 不改。
