# Feature: Health Check and Dynamic Load Balancing

## 1. Summary
Xynonプロキシエンジンにおいて、複数あるバックエンド（アップストリーム）へのトラフィックを適切に分散させるための「動的ロードバランシング機能」と、各バックエンドの死活監視を行う「アクティブ・パッシブヘルスチェック機能」を導入する。これにより、一部のバックエンドがダウンした場合でもトラフィックを正常なサーバーへルーティングし、システム全体の可用性を向上させる。

## 2. Architecture & Component Type
Core Xynon Feature (Proxy Engine) & CLI Tooling

## 3. Behavior & Requirements
- **ロードバランシングアルゴリズム**: 
  - `round_robin` (ラウンドロビン)
  - `least_connections` (最小接続数)
  - `ip_hash` (クライアントIPベースのハッシュ): クライアントの正規IP（X-Forwarded-Forなどの信頼できるプロキシヘッダーを考慮して決定）をソースとする。ハッシュ先のバックエンドがUnhealthyな場合は、Healthyなサーバーから再選択を行い、利用可能なHealthyなサーバーが0台のときのみ503を返す。
  の3種類をサポートし、設定から切り替え可能とする。
- **ヘルスチェック方式**:
  - **状態遷移マシン**: 各バックエンドは `Healthy`, `Unhealthy`, `Recovery` の決定論的ステータスを持つ。フラッピング防止のため、アクティブとパッシブの失敗カウンターは連動・共有管理し、一定時間内の連続失敗回数が `max_fails` を超過した場合に `Unhealthy` へ遷移する。`Unhealthy` 状態からの回復は、アクティブヘルスチェックが所定回数連続成功した場合に行われ、その間は `Recovery` 状態（トラフィックは流さないがチェックは継続）とする。アクティブ側の `interval` と `max_fails`、パッシブ側の `fail_timeout` は相互に連携してタイムウィンドウを形成する。
  - **アクティブヘルスチェック**: 定期的に指定されたパス（例: `/health`）に対してリクエストを送信し、期待するステータスコード（例: `200`）が返るかを監視する。
  - **パッシブヘルスチェック**: プロキシとしてトラフィックをルーティングする際の実際のリクエスト結果を監視し、タイムアウトや5xx系のエラーが連続した場合に異常と判定する。
  - これら両方を組み合わせて動作させることが可能。
- **ステータスの可視化**:
  - プロキシの管理用エンドポイント（例: `/_admin/upstreams`）を提供し、現在の各バックエンドのヘルスステータスをJSON形式で取得できるようにする。このエンドポイントはデフォルトで非公開ネットワークにバインドされ、認証・認可を必須とする。
  - 同様に、CLIツール（例: `xynon admin status` もしくは `xynon upstream list` など）からもこのエンドポイントを叩いて状態を表示できるようにする。CLIは設定等から適切に認証情報を付与してアクセスする。
- **Edge Cases**:
  - **全バックエンドダウン時**: すべてのバックエンドが `Unhealthy` と判定され、利用可能なサーバーが0台になった場合は、クライアントからのリクエストに対して即座に `503 Service Unavailable` を返す。

## 4. Configuration
`config.yaml` 内のアップストリーム定義において、以下のように設定を追加する。

```yaml
upstreams:
  - name: my-backend-cluster
    algorithm: round_robin # round_robin | least_connections | ip_hash
    servers:
      - url: http://backend1.example.com
      - url: http://backend2.example.com
    health_check:
      active:
        enabled: true
        path: "/health"
        expected_status: 200
        interval: "10s"
        timeout: "2s"
        max_fails: 3
      passive:
        enabled: true
        max_fails: 5
        fail_timeout: "30s" # 指定期間内にmax_fails回失敗したらUnhealthy
```
