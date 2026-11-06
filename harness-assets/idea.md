# ccdrawu

由 analysis 從 README 整理（2026-11-06），不是使用者的原話。

Claude Code 的側邊欄與狀態列：用量上限、context、模型、花費、git 和 session 統計放在對話旁邊的面板裡，不再擠在輸入框下方。想放
在輸入框下方的話，也可以當成 Claude Code 的狀態列，照使用者挑的方式排。

給用 Claude Code 的人。側邊欄是 Claude Code 的外掛；`ccdrawu` 指令用來挑側邊欄與狀態列顯示的項目、安裝與更新外掛、移除。在
macOS、Linux、Windows 上用（Windows 在 Windows Terminal 試過）。

側邊欄最上面住著一隻陪伴的像素寵物：Claude 工作時走動、做完時坐下說話、閒著就睡；subagent 在跑時有同類的小隻；git 還沒
commit、還沒 push 的東西在牠的走道上堆成紙箱。

不能發生的事：

- 換成 ccdrawu 的主題與狀態列之後，使用者原本的找不回來：換之前先記下來，關掉或移除時放回去。
- 設定讀不了時動到使用者的檔。
- 寵物的台詞花 token。
- 除了查 Claude Code 的新版，與打開 Status 時查 status.claude.com 之外連網。
