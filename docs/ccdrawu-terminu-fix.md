# ccdrawu 還沒符合 tdp 的地方

照 tdp v0.7.4 與它的 checklist（`checklist/items-zh_TW.md`）對過，2026-10-18，ccdrawu 用 terminu-go v0.1.13。下面每一條都
回報給 terminu 了（CD 編號），等 lib 發版或規範定案再修；全部修完就刪掉這個檔（tdp E7）。偏離 tdp 的地方（有理由、不打算修的）
寫在 [`dev-remarks.md`](dev-remarks.md) 的「偏離 tdp」。

## 等 lib（terminu-go）

| 條目 | 在哪裡 | 現況 | 該怎麼改 |
|---|---|---|---|
| tdp M3、tdp K6（清單 M3-1、K6-5） | `[2]` 每一頁 | 每一列的 `Enter`（選、開 input、開 popup、開 confirm）不在 Space menu，`?` 裡也沒有；preferences 頁的 Space menu 只剩 Global operation | lib 的 `Values` 提供 `[Enter]` 的 item operation（CD19） |
| tdp K7、tdp K12（清單 K7-1、K12-4） | Order 的 reorder popup | `h\|l`、`←\|→` 不動，同一個 app 的其他一維清單都會動 | lib 的 reorder 收 `h\|l`（CD20） |
| tdp M6（清單 M6-6） | 沒有 catalog 時的 `[2]`、Install confirm | `?` 照常列出移動鍵、捲動鍵，其實沒東西可移、可捲 | lib 不列，或用停用色（CD21） |
| tdp E5（清單 E5-10） | 任何 popup | 終端機 5 列以下打開 popup 會 panic | lib 的 draw 不讓高度變負（CD22） |
| components/layout/screen 空狀態（清單 screen-29） | 沒有 catalog 時的 sidebar、inline、statusline 頁 | 只能用 `Failure` 畫，事實那一句是 Red | lib 給 `Values` 一個 Overlay0 的空狀態，ccdrawu 改用它（CD23，`TODO(lib CD23)`） |
| components/layout/popup 錯誤列、dialog/note（清單 popup-69、note-69） | Alias、Separator、Refresh、Claude theme、Drawer background 等 input 的動作失敗時 | popup 已經關了，error popup 蓋在 panel 上，打的值不見；ccdrawu 已經把設定改回檔案裡的樣子，畫面不會假裝存好 | lib 的 input 能回報失敗、留住 popup（CD24，`TODO(lib CD24)`） |
| tdp K3（清單 K3-8） | Alias、Separator、Refresh | lib 在值沒變時照樣呼叫 `Done`；ccdrawu 自己比對，值一樣就不寫 | lib 修好後拿掉 ccdrawu 的比對（CD18，`TODO(lib CD18)`） |
| components/input/README（值裡的控制字元） | Alias、Separator、Drawer background、Plugin 的值 | lib 原樣畫出控制字元；ccdrawu 自己寫成 `\x1b`（`shown()`） | lib 照規範畫（CD25，`TODO(lib CD25)`） |
| components/dialog/confirm、layout/popup hint | Install、Uninstall confirm，窄到出現捲動位置時 | hint 只剩 `Enter:install`，`Esc:cancel` 被擠掉 | lib 讓 confirm 的框放得下 hint 與捲動位置（CD49） |
| components/layout/panel 膠囊串 | `[1]`、`[2]` 的標題，很窄時 | 選中的膠囊比框寬時被硬切，沒有 `…`、右端的斜角不見 | 規範定截法，lib 照做（CD43） |
| components/input/README 太長加 `…` | 最上面一列的 statusbar，很窄時 | `…` 是終端機的預設字色 | lib 的 `Truncate` 給 `…` 顏色（CD44） |

## 等規範（tdp）

| 條目 | 在哪裡 | 現況 | 該怎麼改 |
|---|---|---|---|
| components/input/color-picker | Drawer background 沒設時打開 color picker | old 寫 `NONE`，色塊是終端機的預設前景色；new 從 `#000000` 起 | 規範定空值的畫法，lib 照做（CD12，`TODO(tdp CD12)`） |

Drawer 的 off／on radio（CD13）是 user 定的，寫在 dev-remarks 的偏離，不在這裡。
