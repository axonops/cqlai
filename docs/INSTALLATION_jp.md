# CQLAIインストールガイド

CQLAIは、さまざまなパッケージマネージャー、Docker、またはソースからビルドしてインストールできます。環境に最適な方法を選択してください。

## 目次
- [パッケージマネージャーインストール](#パッケージマネージャーインストール)
  - [APT (Debian/Ubuntu)](#apt-debianubuntu)
  - [YUM/DNF (RHEL/CentOS/Fedora)](#yumdnf-rhelcentosfedora)
- [Dockerインストール](#dockerインストール)
- [バイナリインストール](#バイナリインストール)
- [ソースからビルド](#ソースからビルド)
- [インストールの確認](#インストールの確認)
- [設定](#設定)

## パッケージマネージャーインストール

### APT (Debian/Ubuntu)

CQLAIリポジトリを追加してインストール:

```bash
# リポジトリキーを追加
curl -fsSL https://packages.axonops.com/apt/KEY.gpg | sudo gpg --dearmor -o /usr/share/keyrings/axonops-archive-keyring.gpg

# リポジトリを追加
echo "deb [signed-by=/usr/share/keyrings/axonops-archive-keyring.gpg] https://packages.axonops.com/apt stable main" | sudo tee /etc/apt/sources.list.d/cqlai.list

# パッケージリストを更新
sudo apt update

# CQLAIをインストール
sudo apt install cqlai
```

#### 特定バージョンのインストール
```bash
# 利用可能なバージョンをリスト
apt list -a cqlai

# 特定バージョンをインストール
sudo apt install cqlai=0.0.5
```

#### アップグレード
```bash
sudo apt update && sudo apt upgrade cqlai
```

### YUM/DNF (RHEL/CentOS/Fedora)

CQLAIリポジトリを追加してインストール:

```bash
# リポジトリを追加
sudo tee /etc/yum.repos.d/cqlai.repo <<EOF
[cqlai]
name=CQLAI Repository
baseurl=https://packages.axonops.com/rpm/stable/\$basearch
enabled=1
gpgcheck=1
gpgkey=https://packages.axonops.com/rpm/KEY.gpg
EOF

# CQLAIをインストール
sudo yum install cqlai
# または新しいシステムの場合
sudo dnf install cqlai
```

#### 特定バージョンのインストール
```bash
# 利用可能なバージョンをリスト
yum list available cqlai --showduplicates
# または
dnf list available cqlai --showduplicates

# 特定バージョンをインストール
sudo yum install cqlai-0.0.5
# または
sudo dnf install cqlai-0.0.5
```

#### アップグレード
```bash
sudo yum update cqlai
# または
sudo dnf upgrade cqlai
```

## Dockerインストール

CQLAIはコンテナ化されたデプロイメント向けのDockerイメージとして利用できます。

### 最新イメージをプル
```bash
docker pull registry.axonops.com/axonops-public/axonops-docker/cqlai:latest
```

### 特定バージョンをプル
```bash
docker pull registry.axonops.com/axonops-public/axonops-docker/cqlai:0.0.5
```

### DockerでCQLAIを実行

#### 対話モード
```bash
# ホストネットワーク上のCassandraに接続
docker run -it --rm --network host \
  registry.axonops.com/axonops-public/axonops-docker/cqlai:latest \
  -h localhost -p 9042 -u cassandra

# カスタム設定を使用
docker run -it --rm \
  -v $(pwd)/cqlai.json:/app/cqlai.json:ro \
  registry.axonops.com/axonops-public/axonops-docker/cqlai:latest
```

#### バッチモード
```bash
# 単一クエリを実行
docker run --rm --network host \
  registry.axonops.com/axonops-public/axonops-docker/cqlai:latest \
  -h localhost -u cassandra \
  -e "SELECT * FROM system.local;"

# ファイルからクエリを実行
docker run --rm \
  -v $(pwd)/queries.cql:/queries.cql:ro \
  --network host \
  registry.axonops.com/axonops-public/axonops-docker/cqlai:latest \
  -h localhost -u cassandra \
  -f /queries.cql
```

### Docker Composeの例

```yaml
version: '3.8'

services:
  cqlai:
    image: registry.axonops.com/axonops-public/axonops-docker/cqlai:latest
    container_name: cqlai
    network_mode: host
    volumes:
      - ./cqlai.json:/app/cqlai.json:ro
      - ./queries:/queries:ro
    environment:
      - CQLAI_HOST=localhost
      - CQLAI_PORT=9042
      - CQLAI_USERNAME=cassandra
    stdin_open: true
    tty: true
```

## バイナリインストール

GitHubリリースからプリビルドバイナリをダウンロード:

### Linux (amd64)
```bash
# 最新リリースをダウンロード
curl -L https://github.com/axonops/cqlai/releases/latest/download/cqlai-linux-amd64 -o cqlai

# 実行可能にする
chmod +x cqlai

# PATHに移動(オプション)
sudo mv cqlai /usr/local/bin/
```

### macOS (Intel)
```bash
# 最新リリースをダウンロード
curl -L https://github.com/axonops/cqlai/releases/latest/download/cqlai-darwin-amd64 -o cqlai

# 実行可能にする
chmod +x cqlai

# PATHに移動(オプション)
sudo mv cqlai /usr/local/bin/
```

### macOS (Apple Silicon)
```bash
# 最新リリースをダウンロード
curl -L https://github.com/axonops/cqlai/releases/latest/download/cqlai-darwin-arm64 -o cqlai

# 実行可能にする
chmod +x cqlai

# PATHに移動(オプション)
sudo mv cqlai /usr/local/bin/
```

### Windows
実行可能ファイルを次からダウンロード:
```
https://github.com/axonops/cqlai/releases/latest/download/cqlai-windows-amd64.exe
```

## ソースからビルド

### 前提条件
- Go 1.21以降
- Git

### ビルド手順
```bash
# リポジトリをクローン
git clone https://github.com/axonops/cqlai.git
cd cqlai

# 依存関係をダウンロード
go mod download

# バイナリをビルド
go build -o cqlai cmd/cqlai/main.go

# PATHにインストール(オプション)
sudo cp cqlai /usr/local/bin/
```

### 特定バージョンでビルド
```bash
# ビルド中にバージョンを設定
go build -ldflags "-X main.Version=0.0.5" -o cqlai cmd/cqlai/main.go
```

## インストールの確認

インストール後、CQLAIが動作していることを確認:

```bash
# バージョンを確認
cqlai --version

# Cassandraへの接続をテスト
cqlai -h localhost -p 9042 -u cassandra -e "SELECT release_version FROM system.local;"
```

## 設定

### クイックスタート
```bash
# デフォルト設定を作成
cqlai --generate-config > cqlai.json

# 設定を編集
nano cqlai.json
```

### 設定場所
CQLAIは次の順序で設定を探します:
1. `./cqlai.json`(現在のディレクトリ)
2. `~/.config/cqlai/cqlai.json`(ユーザー設定)
3. `/etc/cqlai/cqlai.json`(システム全体)

### 環境変数
環境変数を使用して設定することもできます:
```bash
export CQLAI_HOST=localhost
export CQLAI_PORT=9042
export CQLAI_USERNAME=cassandra
export CQLAI_PASSWORD=cassandra
export CQLAI_KEYSPACE=system
```

## トラブルシューティング

### 一般的な問題

#### Permission Denied
インストール時に権限エラーが発生した場合:
```bash
# sudo権限があることを確認
sudo apt install cqlai
```

#### Repository Not Found
パッケージリポジトリが見つからない場合:
```bash
# リポジトリキャッシュを更新
sudo apt update
# または
sudo yum makecache
```

#### Dockerネットワークの問題
DockerがCassandraに接続できない場合:
```bash
# ホストネットワークモードを使用
docker run --network host ...

# またはCassandraコンテナ名を指定
docker run --link cassandra:cassandra ...
```

## サポート

問題とサポートについて:
- GitHub Issues: https://github.com/axonops/cqlai/issues
- ドキュメント: https://github.com/axonops/cqlai/tree/main/docs
- リリースノート: https://github.com/axonops/cqlai/releases

## ライセンス

CQLAIはApache License 2.0の下で配布されています。詳細についてはLICENSEファイルを参照してください。
