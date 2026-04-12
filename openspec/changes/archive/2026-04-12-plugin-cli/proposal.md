## Why

現状、プラグインの追加・削除・一覧確認はファイルシステムの操作と設定ファイルの手書き編集で行う必要があり、運用者とプラグイン開発者にとって手間がかかる。`xynon plugin` サブコマンドとして CLI を提供することで、プラグインのライフサイクル管理をシンプルなコマンド操作で完結させる。

## What Changes

- `xynon plugin list` — プラグインディレクトリに存在する `.wasm` ファイルと、設定ファイル上のチェーン順序・有効/無効状態を一覧表示するサブコマンドを追加する。
- `xynon plugin add <path>` — 指定した WASM ファイルをプラグインディレクトリにコピーし、設定ファイルのチェーンに末尾追加するサブコマンドを追加する。
- `xynon plugin remove <name>` — 設定ファイルのチェーンから指定プラグインを除外するサブコマンドを追加する（WASM ファイルは削除しない）。
- `xynon plugin build <src-dir>` — 指定ディレクトリの TinyGo プラグインをビルドして WASM を生成するサブコマンドを追加する。
- 既存の `cmd/xynon/main.go` にサブコマンドディスパッチを追加する。上位の動作（プロキシ起動）は変更しない。

## Capabilities

### New Capabilities
- `plugin-cli`: プラグインのリスト・追加・削除・ビルドを行う CLI サブコマンド群。

### Modified Capabilities
- `proxy-config`: `plugin add` / `plugin remove` が設定ファイルを読み書きするため、設定ファイルの書き込みサポートを要件として追加する。

## Impact

- `cmd/xynon/main.go` にサブコマンドディスパッチを追加する。
- `internal/plugincli/` パッケージを新規作成し、各サブコマンドのロジックを配置する。
- `internal/config/config.go` に設定ファイルの書き戻し機能を追加する。
- TinyGo のビルド呼び出しは外部プロセス (`os/exec`) 経由で行い、ホストビルドへの依存を増やさない。
- 新規依存ライブラリは追加しない（標準ライブラリのみで実装する）。
