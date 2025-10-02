# dockertest Examples

[dockertest](https://github.com/ory/dockertest)を使った統合テストのサンプルプロジェクト。

## 概要

このプロジェクトは、dockertestライブラリを使ってDockerコンテナを起動し、統合テストを実行する方法を示しています。

## テストケース

### 1. TestStdout
`hello-world`コンテナを起動し、標準出力に"Hello from Docker!"が含まれることを確認します。
- コンテナのログ取得方法のデモンストレーション

### 2. TestNginx
nginxコンテナを起動し、HTTP GETリクエストでステータス200を確認します。
- ポートバインディング
- `pool.Retry()`による起動待機
- 動的ポート割り当て（`resource.GetPort()`）

### 3. TestMultipleServices
2つのEchoアプリケーション（service-aとservice-b）を起動し、サービス間通信をテストします。
- service-a: service-bにリクエストし、結果を加工して返す
- service-b: シンプルなJSONレスポンスを返す
- Dockerfileからのビルドと実行（`BuildAndRunWithOptions`）
- サービス間通信のテスト

## プロジェクト構成

```
.
├── main_test.go           # テストコード
├── docker-compose.yml     # Docker Compose設定
├── service-a/
│   ├── main.go           # サービスAのEchoアプリ
│   ├── Dockerfile        # サービスAのDockerfile
│   ├── go.mod
│   └── go.sum
└── service-b/
    ├── main.go           # サービスBのEchoアプリ
    ├── Dockerfile        # サービスBのDockerfile
    ├── go.mod
    └── go.sum
```

## 実行方法

### ローカルで実行

```bash
go test -v ./...
```

### Docker Compose内で実行

```bash
docker compose up
```

Docker Compose内で実行する場合は、`DOCKER_HOST_ADDR`環境変数を使って`host.docker.internal`経由でDockerホストに接続します。

## ポイント

### 動的ポート割り当て
テストでは`resource.GetPort()`を使用して動的に割り当てられたポートを取得します。これにより、ポート競合を避けることができます。

### Docker Compose対応
環境変数`DOCKER_HOST_ADDR`をチェックすることで、ローカル実行とDocker Compose内での実行の両方に対応しています：

```go
dockerHost := os.Getenv("DOCKER_HOST_ADDR")
if dockerHost == "" {
    dockerHost = "localhost"
}
```

### リソースのクリーンアップ
`defer pool.Purge(resource)`を使用して、テスト終了時に確実にコンテナを削除します。

### マルチステージビルド
service-aとservice-bのDockerfileでは、マルチステージビルドを使用してイメージサイズを削減しています。

## 必要な環境

- Go 1.23.0以降
- Docker
- Docker Compose (Docker Compose内で実行する場合)

## 参考

- [dockertest](https://github.com/ory/dockertest)
- [dockertest examples](https://github.com/ory/dockertest/tree/v3/examples)