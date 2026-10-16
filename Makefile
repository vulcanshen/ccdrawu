# ccdrawu — build / run / test / package
#
# 跑 `make`（或 `make help`）列出所有指令。
# 在 macOS / Linux 上開發，CGO_ENABLED=0 靜態編譯；release 另有 Windows（goreleaser，跟 kbu 一樣）。

BINARY   := ccdrawu
PKG      := ./cmd/ccdrawu
DIST_DIR := dist

GOOS   := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w

.DEFAULT_GOAL := help

##@ 編譯（build）

.PHONY: build
build: ## 編譯本地執行檔 → ./ccdrawu（CGO_ENABLED=0 靜態、-trimpath、strip）
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) $(PKG)
	@echo "built ./$(BINARY)  ($(VERSION) $(GOOS)/$(GOARCH))"

.PHONY: install
install: ## go install → $$GOBIN/PATH
	CGO_ENABLED=0 go install -trimpath -ldflags "$(LDFLAGS)" $(PKG)

.PHONY: uninstall
uninstall: ## 移除已安裝的 ccdrawu（從 $$GOBIN，否則 $$GOPATH/bin）
	@dir="$$(go env GOBIN)"; [ -n "$$dir" ] || dir="$$(go env GOPATH)/bin"; \
	if [ -f "$$dir/$(BINARY)" ]; then rm -f "$$dir/$(BINARY)" && echo "removed $$dir/$(BINARY)"; else echo "not installed ($$dir/$(BINARY))"; fi

.PHONY: tidy
tidy: ## go mod tidy
	go mod tidy

##@ 發布（release）

.PHONY: release-check
release-check: ## goreleaser check：驗證 .goreleaser.yaml（要有 GitHub remote）
	goreleaser check

.PHONY: snapshot
snapshot: ## goreleaser 本機 snapshot 打包（不發布、不推 tap / bucket）→ dist/，含 Windows 的 zip
	goreleaser release --snapshot --clean

##@ 執行（run）

.PHONY: run
run: ## 本地跑設定 TUI（讀真正的 config）
	go run $(PKG)

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

##@ 測試 / 檢查（test）

.PHONY: test
test: ## 跑所有測試
	go test ./...

.PHONY: vet
vet: ## go vet ./...
	go vet ./...

.PHONY: fmt
fmt: ## gofmt -w（就地格式化 cmd/ internal/）
	gofmt -w cmd internal

.PHONY: fmt-check
fmt-check: ## gofmt -l（列出未格式化的檔；有輸出即失敗，CI 用）
	@out=$$(gofmt -l cmd internal); if [ -n "$$out" ]; then echo "not gofmt'ed:"; echo "$$out"; exit 1; fi

.PHONY: mod-check
mod-check: ## 檢查 mod：claude plugin validate + claude plugin test（tests/*.test.ts）
	claude plugin validate .
	claude plugin test .

.PHONY: check
check: fmt-check vet test mod-check ## fmt-check + vet + test + mod-check 一次跑（commit 前）

##@ 打包（package）

.PHONY: package
package: build ## 打包 → dist/ccdrawu_<ver>_<os>_<arch>.tar.gz
	@mkdir -p $(DIST_DIR)
	tar -czf $(DIST_DIR)/$(BINARY)_$(VERSION)_$(GOOS)_$(GOARCH).tar.gz $(BINARY)
	@echo "packaged $(DIST_DIR)/$(BINARY)_$(VERSION)_$(GOOS)_$(GOARCH).tar.gz"

##@ 展示（demo）

# vhs 0.12.0 有時兩秒就結束、不出檔也不報錯（webu 踩過）；0.12.1 正常。
# 要指定別的版本：make gif VHS=/path/to/vhs
VHS ?= vhs

.PHONY: gif
gif: build ## 錄 docs/demo.gif 與 docs/settings.gif（需 vhs、JetBrainsMono Nerd Font；demo 開 Claude Code、送一個 prompt；tape 與展示用設定在 .local/demos/，不碰你的設定）
	$(VHS) .local/demos/demo.tape
	$(VHS) .local/demos/settings.tape

##@ 其他

.PHONY: clean
clean: ## 移除 ./ccdrawu 與 dist/
	rm -f $(BINARY)
	rm -rf $(DIST_DIR)

.PHONY: help
help: ## 顯示這份說明
	@awk 'BEGIN {FS = ":.*?## "} \
		/^##@/ {printf "\n\033[1m%s\033[0m\n", substr($$0, 5); next} \
		/^[a-zA-Z0-9_-]+:.*?## / {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
