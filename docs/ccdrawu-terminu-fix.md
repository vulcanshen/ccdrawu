# ccdrawu 還沒符合 tdp 的地方

照 tdp v0.9.0 對過（2026-10-20），ccdrawu 用 terminu-go v0.4.0。下面的條目在等 terminu 回覆（括號裡是 ccdrawu 回報的 CD 編號）：
定案後照改，全部修完就刪掉這個檔（tdp E7）。偏離 tdp 的地方寫在 [`dev-remarks.md`](dev-remarks.md) 的「偏離 tdp」。

| 條目 | 在哪裡 | 現況 | 定案之後怎麼改 |
|---|---|---|---|
| tdp E3 的分類、清單 E3-13（CD55） | `config.yaml` 的 `theme.before`、`statusline.before` | 留在 `config.yaml`（暫定，程式裡標 `TODO(tdp CD55)`）。這兩個是 ccdrawu 換掉使用者自己的 Claude Code theme 與 statusLine 之前的原值，刪了就找不回來，照 E3-13 的字面是資料 | 算資料：搬到資料目錄，照 tdp E3「換了位置時」逐項搬；算設定：拿掉 `TODO(tdp CD55)` 與這一列 |
