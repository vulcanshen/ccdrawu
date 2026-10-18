# ccdrawu — build / run / test / package（家族共用的 target 在 family.mk）
#
# 跑 `make`（或 `make help`）列出所有指令。

BINARY := ccdrawu
CHECK := mod-check
# demo 開 Claude Code、送一個 prompt；settings 是設定 TUI。展示用的設定在 .local/demos/，不碰你的設定
TAPES := .local/demos/demo.tape .local/demos/settings.tape
include family.mk

##@ ccdrawu 自己的

# dev build 的 install claude 不指 CCDRAWU__MARKETPLACE 會裝 GitHub main 的外掛，
# 把這個資料夾裝的換掉；這個 target 一律指著這裡。Claude Code 把市集資料夾整個複製進外掛快取，
# git 不收的 .local 也一起（demos 有 Claude Code 的執行檔，一份 1.8G），所以先把 git 收的檔案
# （含還沒 commit、還沒 add 的）照原樣放進 .local/plugin，從那裡裝
.PHONY: plugin
plugin: build ## 把這個資料夾 git 收的檔案當市集，裝或更新 Claude Code 裡的外掛（真正的 Claude Code 設定）；開著的 session 跑 /reload-plugins
	rm -rf .local/plugin
	mkdir -p .local/plugin
	git ls-files -co --exclude-standard | while IFS= read -r f; do [ ! -e "$$f" ] || printf '%s\n' "$$f"; done | tar -cf - -T - | tar -xf - -C .local/plugin
	CCDRAWU__MARKETPLACE=$(CURDIR)/.local/plugin ./$(BINARY) install claude

.PHONY: mod-check
mod-check: ## 檢查 mod：claude plugin validate + claude plugin test（tests/*.test.ts）；check 最後跑它
	claude plugin validate .
	claude plugin test .
