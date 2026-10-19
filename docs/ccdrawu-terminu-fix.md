# ccdrawu 還沒符合 tdp 的地方

照 tdp v0.8.0 對過（2026-10-19），ccdrawu 用 terminu-go v0.1.13。下面每一條的規範都定案了（括號裡是 ccdrawu 回報的 CD 編號與
定案的 S 編號），都在等 terminu-go 的第三批發版：發版後升級 go.mod、照改，全部修完就刪掉這個檔（tdp E7）。偏離 tdp 的地方
寫在 [`dev-remarks.md`](dev-remarks.md) 的「偏離 tdp」。

| 條目 | 在哪裡 | 現況 | 等 lib 之後怎麼改 |
|---|---|---|---|
| tdp K7、tdp K12、components/input/reorder（CD20，定案 S21） | Order 的 reorder popup | `h\|l`、`←\|→` 不動 | lib 的 reorder 收 `h\|l`，Drag 時也一樣；ccdrawu 不用改 |
| tdp M6（CD21，定案 S20） | 沒有 catalog 時的 `[2]`、Install confirm | `?` 照常列出移動鍵、捲動鍵 | lib 照當下的狀態給；ccdrawu 不用改 |
| tdp E5（CD22） | 任何 popup | 終端機 5 列以下打開 popup 會 panic | 第一批 #1 修好；`TestL4EveryRowFullWidth` 加 5 列以下的尺寸 |
| tdp L1（定案 S53） | 整個 TUI | 沒有最小尺寸的門檻 | 設 `App.MinWidth = 40`、`App.MinHeight = 12`（terminu-go v0.2.0，user 定的，見 dev-remarks 的設計決定），測試加門檻以下到 1×1 |
| components/layout/screen 空狀態（CD23） | 沒有 catalog 時的 sidebar、inline、statusline 頁 | 用 `Failure` 畫，事實那一句是 Red | 改用 lib 給 `Values` 的空狀態（第二批 #8），拿掉 `TODO(lib CD23)` |
| components/layout/popup 錯誤列、dialog/note（CD24） | Alias、Separator、Refresh、Claude theme、Drawer background 的動作失敗時 | popup 已經關了，error popup 蓋在 panel 上；ccdrawu 把設定改回檔案裡的樣子 | 改用能回報失敗的 Done（第二批 #1），拿掉 `revert` 裡的 `TODO(lib CD24)` |
| components/input/README 值裡的控制字元（CD25，定案 S22） | Alias、Separator、Drawer background、Plugin 的值 | ccdrawu 自己寫成 `\x1b`，沒上色（`shown()`） | lib 畫成紅色的轉義後拿掉 `shown()` 與 `TODO(lib CD25)` |
| components/input/color-picker（CD12，定案 S40） | Drawer background 沒設時打開 color picker | old 寫 `NONE`、色塊是終端機的前景色 | lib 照改（old 寫灰色的 `none`、沒有色塊），拿掉 `TODO(lib CD12)` |
| components/input/number 的單位（CD51，定案 S47） | Refresh 的輸入框 | panel 上已經寫 `10 seconds`；輸入框沒有單位列 | NumberInput 給單位 `seconds`，拿掉 `TODO(lib CD51)` |
| components/dialog/toast 的 Success（CD27，定案 S8） | 套用主題、打開或關掉狀態列、安裝或移除外掛之後的 toast | 用 Info | 成功做完的改用 Success（Green）；搬家照 tdp E3 留 Info |
| components/layout/panel 一列一個值的捲動位置（CD40，定案 S11） | `[2]` 的 `N of M` | radio 的每個選項各算一個 | lib 照欄位算；ccdrawu 不用改 |
| components/layout/popup hint、panel 膠囊、screen 空狀態的「放不下時」（CD43、CD49，定案 S12） | 很窄時的 confirm、`[1]`／`[2]` 的標題 | hint 擠掉 `Esc`、膠囊被硬切 | lib 照改；ccdrawu 不用改 |
| components/input/README 太長加 `…`（CD44） | 最上面一列的 statusbar，很窄時 | `…` 是終端機的預設字色 | 第一批 #3 修好；ccdrawu 不用改 |

已經符合、但靠 ccdrawu 自己的暫時做法的，lib 發版後拿掉：`[2]` 的 Enter 列進 Space menu 與 `?`（`enterField`，`TODO(lib CD19)`，第二批 #7）；值沒變不寫回（三處 `TODO(lib CD18)`，第一批 #2）。
