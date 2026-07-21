# Health Check and Dynamic Load Balancing

## 概要 (Overview)
Xynonプロキシエンジンにおいて、複数バックエンドサーバーへのトラフィック分散（ロードバランシング）および正常性確認（ヘルスチェック）を行う機能の仕様を定義します。

## 設定 (Configuration)
`config.yaml` に `upstreams` セクションを追加し、複数のアップストリーム群を定義できるようにします。

```yaml
listen: ":8080"
upstreams:
  my_backend:
    algorithm: round_robin # round_robin, least_connections, ip_hash
    servers:
      - "http://127.0.0.1:8081"
      - "http://127.0.0.1:8082"
    health_check:
      path: "/health"
      interval: "5s"
      timeout: "2s"
      healthy_threshold: 2
      unhealthy_threshold: 2
```

## ロードバランシング (Load Balancing)
プロキシに届いたリクエストの `Host` ヘッダが `upstreams` のキー（例: `my_backend`）に一致する場合、リクエストを対象のサーバーへルーティングします。
- `round_robin`: 順番に割り当てます。
- `least_connections`: 現在のアクティブなリクエスト数が最も少ないサーバーに割り当てます。
- `ip_hash`: クライアントのIPアドレスのハッシュ値を用いて一貫した割り当てを行います。

利用可能なサーバーが0台になった場合は、直ちに `503 Service Unavailable` を返却します。

## ヘルスチェック (Health Check)
- **アクティブ**: 定期的に（`interval` ごとに）`path` へ HTTP GET リクエストを送信し、200番台の応答が得られるか確認します。
- **パッシブ**: プロキシ転送時に接続エラーやタイムアウト、または 5xx エラーが発生した場合に、失敗とカウントします。

## 管理用エンドポイント (Admin Endpoint)
`/_admin/upstreams` にアクセスすると、現在の各アップストリームのステータス（利用可能なサーバー、ヘルスチェックの状態など）を JSON 形式で返却します。

## CLI ツール (CLI Tooling)
`xynon upstreams status` などのコマンドを提供し、内部で管理用エンドポイント（デフォルト: `http://localhost:8080/_admin/upstreams`）を呼び出してステータスを表示します。
