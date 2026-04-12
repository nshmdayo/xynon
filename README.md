# Xynon

Go 製の HTTP フォワードプロキシ。リクエスト/レスポンス処理を **TinyGo 製 WASM プラグイン** で拡張でき、プロキシを再起動せずに **ホットリロード** でプラグインを差し替えられます。

## 構成

```
cmd/xynon/           ← プロキシの CLI エントリポイント
internal/config/     ← 設定ローダ
internal/plugin/     ← WASM ランタイム・ABI・ローダ
internal/plugin/abi/ ← ABI 定数・ドキュメント
internal/proxy/      ← プロキシハンドラ・チェーンレジストリ・ホットリロード
examples/plugins/    ← サンプルプラグイン
examples/config.yaml ← サンプル設定
```

## ビルド

### プロキシ本体

```bash
CGO_ENABLED=0 go build -o bin/xynon ./cmd/xynon
```

### プラグイン (TinyGo が必要)

TinyGo 0.40.1 は Go 1.19–1.25 をサポートしています。Go 1.23 を用意し、以下のように GOROOT を指定してビルドします。

```bash
GO123=/path/to/go1.23   # mise 例: ~/.local/share/mise/installs/go/1.23.12
cd /tmp                  # プロジェクト go.mod の干渉を避けるため /tmp で実行

GOROOT="$GO123" GOWORK=off PATH="$GO123/bin:$PATH" \
  tinygo build \
    -o /path/to/xynon/examples/plugins/add-header/add-header.wasm \
    -target wasip1 \
    /path/to/xynon/examples/plugins/add-header/main.go
```

または Makefile ターゲットを使用します（`make` が利用可能な場合）:

```bash
make plugins
```

## 起動

```bash
./bin/xynon -config examples/config.yaml
```

### フラグ

| フラグ | デフォルト | 説明 |
|--------|-----------|------|
| `-config` | `config.yaml` | 設定ファイルパス |
| `-no-hot-reload` | `false` | ホットリロードを無効化し起動時チェーン固定 |

## プラグイン管理 CLI

プロキシを起動せずにプラグインを管理するサブコマンド群です。

### plugin list — 一覧表示

```bash
./bin/xynon plugin list -config examples/config.yaml
```

プラグインディレクトリの `.wasm` ファイルとチェーン登録状態（順序番号）を表示します。

### plugin add — プラグインを追加

```bash
./bin/xynon plugin add -config examples/config.yaml ./path/to/my-plugin.wasm
```

指定した WASM ファイルをプラグインディレクトリにコピーし、設定ファイルのチェーン末尾に追加します。同名プラグインが既に登録済みの場合はファイルのみ上書きします。

### plugin remove — プラグインをチェーンから除外

```bash
./bin/xynon plugin remove -config examples/config.yaml my-plugin
```

設定ファイルのチェーンから指定名のプラグインを削除します。WASM ファイル自体は削除されません。

### plugin build — TinyGo プラグインをビルド

```bash
./bin/xynon plugin build -config examples/config.yaml ./examples/plugins/add-header
```

指定ディレクトリの TinyGo ソースをビルドし、設定ファイルのプラグインディレクトリに `<ディレクトリ名>.wasm` を出力します。

| 環境変数 | 説明 |
|----------|------|
| `TINYGO` | tinygo バイナリのパス（未設定時は PATH から解決） |
| `XYNON_TINYGO_GOROOT` | TinyGo に渡す GOROOT（Go 1.26+ 環境で TinyGo 0.40.1 を使う場合は Go 1.23 の GOROOT を指定） |

## 設定ファイル (YAML)

```yaml
listen: ":8080"

plugins:
  dir: "./examples/plugins"
  chain:
    - name: add-header    # examples/plugins/add-header/add-header.wasm を使用
```

| フィールド | 必須 | 説明 |
|-----------|------|------|
| `listen` | ✓ | リスンアドレス（例: `:8080`） |
| `plugins.dir` | ✓ | WASM ファイルが置かれるディレクトリ |
| `plugins.chain[]` | - | 有効化するプラグインと実行順序 |

## ホットリロードの動作確認

1. プロキシを起動する:
   ```bash
   ./bin/xynon -config examples/config.yaml
   ```

2. 別ターミナルでリクエストを送信する:
   ```bash
   curl -x http://localhost:8080 http://httpbin.org/headers
   ```
   レスポンスヘッダに `X-Xynon: true` が含まれることを確認します。

3. プラグインの WASM ファイルを上書きする（別バージョンのプラグインに差し替え）:
   ```bash
   cp new-plugin.wasm examples/plugins/add-header/add-header.wasm
   ```
   プロキシのログに `hot-reload complete` が表示され、次のリクエストから新バージョンが適用されます。

4. プロキシは `Ctrl+C` (SIGINT) または SIGTERM でシャットダウンします。

## プラグイン開発 (ABI v1)

プラグインは `internal/plugin/abi/abi.go` に記載された ABI を実装します。

### 必須エクスポート関数

```go
//export xynon_abi_version
func xynon_abi_version() int32 { return 1 }

//export on_request
func on_request(handle int32) int32 { /* 0=continue, 1=short-circuit, 2=error */ }

//export on_response
func on_response(handle int32) int32 { /* 0=continue, 1=short-circuit, 2=error */ }
```

### ホスト関数 (env 名前空間)

| 関数 | シグネチャ | 説明 |
|------|-----------|------|
| `get_header` | `(handle, namePtr, nameLen, outPtr, outCap) -> i32` | ヘッダ値を読み取る |
| `set_header` | `(handle, namePtr, nameLen, valPtr, valLen) -> i32` | ヘッダを設定する |
| `set_status` | `(handle, status) -> i32` | レスポンスステータスコードを設定する |
| `short_circuit` | `(handle, status) -> i32` | 指定ステータスで即時応答（on_request のみ有効） |
| `xynon_log` | `(handle, msgPtr, msgLen) -> i32` | ログ出力 |

ハンドルは各フック呼び出しの間のみ有効です。フック終了後に使用するとエラーが返ります。

## ABI バージョンポリシー

- ABI バージョンは `internal/plugin/abi/abi.go` の `CurrentVersion` 定数で管理します。
- ホストは自身がサポートするバージョンのプラグインのみロードします。非対応バージョンはロード時に拒否されます。
- **破壊的変更**（関数シグネチャ変更・削除）を行う場合は `CurrentVersion` をインクリメントします。
- **後方互換な追加**（新しいホスト関数の追加）は同一バージョン内で行えます。プラグインは不要なホスト関数をインポートしなくても動作します。
