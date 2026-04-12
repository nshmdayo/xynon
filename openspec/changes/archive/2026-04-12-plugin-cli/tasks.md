## 1. 設定ファイルへの書き戻し（proxy-config 拡張）

- [x] 1.1 `internal/config/config.go` に `Save(path string, cfg *Config) error` 関数を追加し、Config 構造体を YAML にシリアライズしてファイルに書き出す
- [x] 1.2 `Save` の単体テスト（書き戻し後に `Load` で正常読み込みできること、チェーン変更が反映されること）を追加する

## 2. サブコマンドフレームワーク

- [x] 2.1 `cmd/xynon/main.go` を `os.Args[1]` でトップレベルコマンドを判定する構造に変更し、`plugin` サブコマンドを `internal/plugincli` にディスパッチする（既存の `-config` フラグによるプロキシ起動は維持する）
- [x] 2.2 `internal/plugincli/` パッケージを作成し、`Run(args []string) error` を持つエントリポイントを実装する

## 3. plugin list

- [x] 3.1 `internal/plugincli/list.go` を実装する。プラグインディレクトリの `.wasm` ファイルを走査し、チェーン登録状態（順序番号）を含む一覧を標準出力に表示する
- [x] 3.2 `list` のテスト（プラグインあり・ディレクトリ空の各ケース）を追加する

## 4. plugin add

- [x] 4.1 `internal/plugincli/add.go` を実装する。WASM ファイルをプラグインディレクトリにコピーし、`config.Save` でチェーンに末尾追加する
- [x] 4.2 同名プラグインが既登録の場合はファイル上書きのみ行い、チェーンに重複追加しないことを確認するテストを追加する
- [x] 4.3 存在しないファイルパスを渡したときにエラーを返し設定を変更しないテストを追加する

## 5. plugin remove

- [x] 5.1 `internal/plugincli/remove.go` を実装する。チェーンから指定名のエントリを削除し `config.Save` で書き戻す（WASM ファイルは削除しない）
- [x] 5.2 未登録名を渡したときにエラーを返し設定を変更しないテストを追加する

## 6. plugin build

- [x] 6.1 `internal/plugincli/build.go` を実装する。`TINYGO` 環境変数または PATH から `tinygo` バイナリを解決し、`tinygo build -target wasip1 -o <plugins.dir>/<name>.wasm <src-dir>` を実行する
- [x] 6.2 TinyGo が見つからない場合に推奨インストール手順を含むエラーメッセージを返すテストを追加する
- [x] 6.3 `XYNON_TINYGO_GOROOT` 環境変数が設定されている場合に `GOROOT` として渡すことを確認するテストを追加する

## 7. ドキュメント

- [x] 7.1 `README.md` の「起動」セクションに `plugin` サブコマンドの使い方（list / add / remove / build）と各フラグを追記する
