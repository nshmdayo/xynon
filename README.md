# nproxy

Go言語で実装されたHTTP/HTTPSプロキシサーバーです。シンプルなフォワードプロキシから本格的なMITM（Man-in-the-Middle）プロキシまで対応しています。

## 機能

- **基本プロキシ機能**: HTTP リクエストの転送
- **MITM プロキシ機能**: HTTPS トラフィックの傍受・改ざん
- **証明書生成**: 動的なサーバー証明書の生成
- **リクエスト・レスポンス改ざん**: ヘッダーやコンテンツの書き換え
- **詳細ログ**: リクエスト・レスポンスの詳細ログ出力
- **セキュリティヘッダー追加**: レスポンスへのセキュリティヘッダー自動追加

## 使用方法

### 基本的なプロキシサーバーとして起動

```bash
# Goで直接実行
go run app/main.go

# または Makefileを使用
make run
```

### MITM プロキシサーバーとして起動

```bash
# MITMプロキシを起動（ログ出力のみ）
go run app/main.go -mitm -addr :8080

# MITMプロキシを起動（リクエスト・レスポンス改ざん有効）
go run app/main.go -mitm -modify -v -addr :8080

# または Makefileを使用
make run-mitm
make run-mitm-modify
```

### コマンドラインオプション

- `-addr`: サーバーのアドレス（デフォルト: `:8080`）
- `-mitm`: MITMプロキシとして起動
- `-modify`: リクエスト・レスポンスの改ざんを有効にする
- `-v`: 詳細ログを出力

### Dockerで起動

```bash
# 基本プロキシ
make start

# MITMプロキシ
make mitm

# MITMプロキシ（改ざん有効）
make mitm-modify
```

## MITM プロキシの使用

MITMプロキシを使用する場合は、以下の手順を実行してください：

1. **プロキシを起動**
   ```bash
   make run-mitm
   ```

2. **CA証明書をインストール**
   - プロキシ起動時に `./certs/ca.crt` に CA証明書が生成されます
   - この証明書をブラウザまたはシステムの信頼する証明書ストアにインストールしてください

3. **ブラウザのプロキシ設定**
   - HTTPプロキシ: `localhost:8080`
   - HTTPSプロキシ: `localhost:8080`

### CA証明書のインストール方法

#### macOS
```bash
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ./certs/ca.crt
```

#### Linux (Ubuntu/Debian)
```bash
sudo cp ./certs/ca.crt /usr/local/share/ca-certificates/nproxy-ca.crt
sudo update-ca-certificates
```

#### Windows
PowerShellで管理者権限で実行：
```powershell
Import-Certificate -FilePath ".\certs\ca.crt" -CertStoreLocation "Cert:\LocalMachine\Root"
```

## MITM機能の詳細

### リクエスト改ざん例

- `X-MITM-Proxy: true` ヘッダーの追加
- User-Agentの変更
- APIリクエストへの特別なヘッダー追加

### レスポンス改ざん例

- セキュリティヘッダーの自動追加
  - `X-Content-Type-Options: nosniff`
  - `X-Frame-Options: DENY`
  - `X-XSS-Protection: 1; mode=block`
- HTMLコンテンツの識別とマーキング
- カスタムヘッダーの追加

## テスト

```bash
# 全てのテストを実行
make test

# 詳細ログ付きでテストを実行
make test-verbose

# 特定のテストのみ実行
go test ./app/proxy/ -run TestMITMProxy
```

## セキュリティ注意事項

⚠️ **重要**: MITMプロキシは教育・デバッグ目的でのみ使用してください。

- 他人のネットワークトラフィックを無断で傍受することは違法です
- 本ツールの使用による損害について、開発者は一切の責任を負いません
- 生成されるCA証明書は適切に管理し、不要になったら削除してください

## ライセンス

このプロジェクトはMITライセンスの下で公開されています。詳細は [LICENSE](LICENSE) ファイルをご覧ください。

## 貢献

バグ報告や機能追加の提案は Issue または Pull Request でお願いします。