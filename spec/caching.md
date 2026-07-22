# Feature: Caching WASM Plugin

## 1. Summary
頻繁にリクエストされるGETリクエストのレスポンスをプロキシ側でキャッシュし、バックエンドの負荷を軽減するとともにクライアントへのレスポンス速度を向上させるWASMプラグインの実装。

## 2. Architecture & Component Type
- **Component**: WASM Plugin (`examples/plugins/caching`)
- **Integration**: Proxy Engine (`internal/proxy`)

## 3. Behavior & Requirements
- **Request Information via Headers**:
  WASMプラグインはホストからパスやメソッドを直接取得するABIを持たないため、Xynonプロキシはプラグインの実行前に以下の擬似ヘッダーを付与する。
  - `X-Xynon-Req-Uri`: リクエストのURI（パスとクエリパラメータを含む）
  - `X-Xynon-Req-Method`: リクエストのメソッド
  - `X-Xynon-Res-Body`: バックエンドからのレスポンスボディ（Base64エンコード）

- **キャッシュキー**:
  - `X-Xynon-Req-Method` が `GET` 以外の場合はキャッシュをバイパスする。
  - `X-Xynon-Req-Uri` をキャッシュキーとして使用する。

- **キャッシュロジック (LRU)**:
  - プラグイン内部でオンメモリのLRUキャッシュを管理する。
  - TTL (Time To Live) をサポートし、期限切れのキャッシュはパージまたは無効化する。
  - `Cache-Control` ヘッダーを解釈し、`no-cache` や `no-store` が指定されている場合はキャッシュへの保存およびキャッシュからの読み出しをバイパスする。

- **OnRequest (キャッシュヒット時の処理)**:
  - キャッシュに有効なデータが存在する場合、`X-Xynon-Res-Body` に Base64エンコードされたボディをセットし、レスポンスヘッダーも復元した上で、`ActionShortCircuit` を返す。
  - プロキシ側は、プラグインが `ActionShortCircuit` を返した際、`X-Xynon-Res-Body` ヘッダーが存在すればそれをデコードしてクライアントへ返す。

- **OnResponse (キャッシュ保存処理)**:
  - `GET` リクエストかつ `Cache-Control` で許可されている場合、`X-Xynon-Res-Body` ヘッダーからレスポンスボディを取得し、レスポンスヘッダーとステータスコードと共にLRUキャッシュに保存する。

## 4. Configuration
`config.yaml` 経由での設定パースをサポートする。
```yaml
plugins:
  - name: caching
    type: wasm
    path: ./examples/plugins/caching/caching.wasm
    config:
      max_size: 100    # LRUキャッシュの最大エントリ数
      default_ttl: 60  # デフォルトのTTL (秒)
```
