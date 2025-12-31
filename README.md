# Go MQTT API

MQTT受信・データ永続化・API提供を行うモダンなGoアプリケーションのプロトタイプです。

## アーキテクチャ

Clean Architectureに基づき、以下のレイヤーで構成されています：

- **Domain層**: エンティティとリポジトリインターフェース定義
- **Usecase層**: ビジネスロジック
- **Infrastructure層**: DB/MQTT/Redisの具象実装
- **Interfaces層**: HTTPハンドラー

## 技術スタック

- **Language**: Go 1.23+
- **MQTT**: eclipse/paho.mqtt.golang
- **API Framework**: Labstack Echo v4
- **RDBMS**: PostgreSQL (GORM)
- **KVS**: Redis (go-redis/v9)
- **Config**: caarlos0/env (環境変数ベース)

## 機能

1. **MQTT受信**: 特定のトピック（デフォルト: `sensors/#`）をSubscribeし、受信したJSONデータを処理
2. **データ永続化**:
   - PostgreSQL: センサーデータの全履歴保存
   - Redis: 各デバイスの最新値のみを保存（高速取得用）
3. **REST API**:
   - `GET /devices/:id/latest`: Redisから最新値を取得
   - `GET /devices/:id/history`: PostgreSQLから履歴を取得（ページネーション対応）

## セットアップ

### 前提条件

- Go 1.23以上
- PostgreSQL
- Redis
- MQTTブローカー（Mosquitto等）

### インストール

1. リポジトリをクローン
```bash
git clone <repository-url>
cd go_mqtt_api
```

2. 依存パッケージのインストール
```bash
go mod download
```

3. 環境変数の設定
```bash
cp .env.example .env
# .envファイルを編集して設定を変更
```

### 環境変数

主要な環境変数は以下の通りです：

- `MQTT_BROKER_URL`: MQTTブローカーURL（デフォルト: `tcp://localhost:1883`）
- `MQTT_TOPIC`: Subscribeするトピック（デフォルト: `sensors/#`）
- `POSTGRES_DSN`: PostgreSQL接続文字列
- `REDIS_ADDR`: Redisアドレス（デフォルト: `localhost:6379`）
- `HTTP_PORT`: HTTPサーバーポート（デフォルト: `8080`）

詳細は`.env.example`を参照してください。

### データベースの準備

PostgreSQLでデータベースを作成：

```sql
CREATE DATABASE mqtt_api;
```

アプリケーション起動時に自動的にテーブルが作成されます。

### Docker Composeを使用したセットアップ（推奨）

動作確認用に、PostgreSQL、Redis、MQTTブローカー（Mosquitto）を含むdocker-compose.ymlを用意しています。

1. Docker Composeでサービスを起動：

```bash
docker-compose up -d
```

これにより、以下のサービスが起動します：
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- Mosquitto (MQTT): `localhost:1883`

2. 環境変数の設定（デフォルト値で動作します）：

```bash
cp env.example .env
```

デフォルト設定で動作するため、`.env`ファイルの編集は任意です。

3. アプリケーションの起動：

```bash
go run cmd/api/main.go
```

4. サービスの停止：

```bash
docker-compose down
```

データを保持したい場合は：

```bash
docker-compose down -v  # ボリュームも削除
```

## 実行

```bash
go run cmd/api/main.go
```

## APIエンドポイント

### 最新値の取得

```bash
GET /devices/:id/latest
```

**レスポンス例:**
```json
{
  "device_id": "sensor01",
  "value": 25.5,
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 履歴の取得

```bash
GET /devices/:id/history?limit=100&offset=0
```

**クエリパラメータ:**
- `limit`: 取得件数（デフォルト: 100、最大: 1000）
- `offset`: オフセット（デフォルト: 0）

**レスポンス例:**
```json
{
  "data": [
    {
      "id": 1,
      "device_id": "sensor01",
      "value": 25.5,
      "timestamp": "2024-01-01T00:00:00Z",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 1000,
  "limit": 100,
  "offset": 0
}
```

## MQTTメッセージ形式

アプリケーションは以下の形式のJSONメッセージを期待します：

```json
{
  "device": {
    "id": "sensor01"
  },
  "data": {
    "value": 25.5,
    "timestamp": "2024-01-01T00:00:00Z"
  }
}
```

## Graceful Shutdown

アプリケーションはSIGINT（Ctrl+C）またはSIGTERMシグナルを受信すると、Graceful Shutdownを実行します：

1. 新しいHTTPリクエストの受付を停止
2. 既存のリクエストの処理完了を待機
3. MQTTクライアントの切断
4. データベース接続のクローズ

## ディレクトリ構成

```
.
├── cmd/
│   └── api/
│       └── main.go          # エントリーポイント
├── config/
│   └── config.go             # 環境設定読み込み
├── internal/
│   ├── domain/               # ドメイン層
│   │   ├── sensor_data.go
│   │   └── repository.go
│   ├── usecase/              # ユースケース層
│   │   └── sensor_data_service.go
│   ├── infrastructure/       # インフラ層
│   │   ├── mqtt/
│   │   │   └── client.go
│   │   ├── postgres/
│   │   │   └── repository.go
│   │   └── redis/
│   │       └── repository.go
│   └── interfaces/           # インターフェース層
│       └── http/
│           └── handler.go
├── mosquitto/
│   ├── config/
│   │   └── mosquitto.conf    # Mosquitto設定ファイル
│   ├── data/                 # Mosquittoデータ
│   └── log/                  # Mosquittoログ
├── docker-compose.yml        # Docker Compose設定
├── env.example               # 環境変数サンプル
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## ライセンス

MIT

