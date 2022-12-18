.PHONY: all test clean zip mac

### バージョンの定義
VERSION     := "v1.0.0"
COMMIT      := $(shell git rev-parse --short HEAD)

### コマンドの定義
GO          = go
GO_BUILD    = $(GO) build
GO_TEST     = $(GO) test -v
GO_LDFLAGS  = -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)"
ZIP          = zip

### ターゲットパラメータ
DIST = dist
SRC = ./main.go
TARGETS     = $(DIST)/twnarrator.exe $(DIST)/twnarrator.app
GO_PKGROOT  = ./...

### PHONY ターゲットのビルドルール
all: $(TARGETS)
clean:
	rm -rf $(TARGETS) $(DIST)/*.zip
mac: $(DIST)/twnarrator.app
zip: $(TARGETS)
	cd dist && $(ZIP) twnarrator.zip twnarrator.exe wnarrator.app


### 実行ファイルのビルドルール
$(DIST)/twnarrator.exe: $(SRC)
	env GO111MODULE=on GOOS=windows GOARCH=amd64 $(GO_BUILD) $(GO_LDFLAGS) -o $@
$(DIST)/twnarrator.app: $(SRC)
	env GO111MODULE=on GOOS=darwin GOARCH=amd64 $(GO_BUILD) $(GO_LDFLAGS) -o $@

