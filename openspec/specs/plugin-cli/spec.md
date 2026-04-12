### Requirement: プラグイン一覧の表示

`xynon plugin list` は設定ファイルで指定されたプラグインディレクトリ内の `.wasm` ファイルを走査し、プラグイン名・ファイルパス・チェーンへの登録状態を一覧表示 SHALL する。

#### Scenario: 登録済みプラグインが存在する
- **WHEN** `xynon plugin list -config <cfg>` を実行する
- **THEN** 設定のチェーンに含まれるプラグイン名と WASM ファイルパスが一覧で出力される
- **AND** チェーン内の順序番号が各行に表示される

#### Scenario: プラグインディレクトリが空
- **WHEN** プラグインディレクトリに `.wasm` ファイルが存在しない
- **THEN** 「プラグインが見つかりません」に相当するメッセージを表示して終了する

### Requirement: プラグインの追加

`xynon plugin add <path>` は指定した WASM ファイルをプラグインディレクトリにコピーし、設定ファイルのチェーン末尾に追加 SHALL する。同名プラグインが既に登録済みの場合はファイルのみ上書きし、チェーンへの重複追加は行わない。

#### Scenario: 新規プラグインを追加する
- **WHEN** `xynon plugin add ./my-plugin.wasm -config <cfg>` を実行する
- **THEN** `my-plugin.wasm` がプラグインディレクトリにコピーされる
- **AND** 設定ファイルのチェーンに `my-plugin` が末尾追加される

#### Scenario: 同名プラグインを再追加する
- **WHEN** チェーンに既に登録済みのプラグインと同名の WASM ファイルを指定して `add` を実行する
- **THEN** WASM ファイルは上書きコピーされる
- **AND** チェーンへの重複追加は行われない

#### Scenario: 存在しないファイルを指定する
- **WHEN** 存在しないパスを `add` に渡す
- **THEN** エラーメッセージを表示して終了する（設定ファイルは変更されない）

### Requirement: プラグインの除外

`xynon plugin remove <name>` は設定ファイルのチェーンから指定名のプラグインエントリを削除 SHALL する。WASM ファイル自体は削除しない。

#### Scenario: 登録済みプラグインを除外する
- **WHEN** `xynon plugin remove add-header -config <cfg>` を実行する
- **THEN** 設定ファイルのチェーンから `add-header` が削除される
- **AND** WASM ファイルはプラグインディレクトリに残る

#### Scenario: 未登録のプラグイン名を指定する
- **WHEN** チェーンに存在しない名前を `remove` に渡す
- **THEN** エラーメッセージを表示して終了する（設定ファイルは変更されない）

### Requirement: プラグインのビルド

`xynon plugin build <src-dir>` は指定ディレクトリの TinyGo ソースをビルドし、設定ファイルのプラグインディレクトリに WASM ファイルを出力 SHALL する。TinyGo のパスは `PATH` から解決し、環境変数 `TINYGO` で上書きできる。

#### Scenario: 有効なソースをビルドする
- **WHEN** TinyGo がインストール済みの環境で `xynon plugin build ./my-plugin` を実行する
- **THEN** `<plugins.dir>/my-plugin.wasm` が生成される

#### Scenario: TinyGo が見つからない
- **WHEN** `tinygo` が PATH になく `TINYGO` 環境変数も未設定の状態でビルドを実行する
- **THEN** TinyGo が見つからない旨と推奨インストール手順を含むエラーを表示して終了する
