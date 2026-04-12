## MODIFIED Requirements

### Requirement: 設定ファイルへの書き戻し

`plugin add` および `plugin remove` コマンドは設定ファイルを読み込み、チェーンを変更した後、同一パスに YAML として書き戻し SHALL する。書き戻し後のファイルは `config.Load` で正常にロードできる状態でなければならない。既存のコメントは保持されない場合がある。

#### Scenario: plugin add 後の設定ファイルを読み込める
- **WHEN** `xynon plugin add` 実行後に `config.Load` を呼び出す
- **THEN** エラーなくロードでき、追加したプラグインがチェーンに含まれる

#### Scenario: plugin remove 後の設定ファイルを読み込める
- **WHEN** `xynon plugin remove` 実行後に `config.Load` を呼び出す
- **THEN** エラーなくロードでき、削除したプラグインがチェーンに含まれない
