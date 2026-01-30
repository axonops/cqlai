<div align="center">
  <img src="./assets/cqlai-logo.svg" alt="CQLAI Logo" width="400">

  # CQLAI - モダンなCassandra® CQLシェル

  [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
  [![Go Version](https://img.shields.io/github/go-mod/go-version/axonops/cqlai)](https://golang.org/)
  [![GitHub Issues](https://img.shields.io/github/issues/axonops/cqlai)](https://github.com/axonops/cqlai/issues)
  [![GitHub Discussions](https://img.shields.io/github/discussions/axonops/cqlai)](https://github.com/axonops/cqlai/discussions)
  [![GitHub Stars](https://img.shields.io/github/stars/axonops/cqlai)](https://github.com/axonops/cqlai/stargazers)
</div>

**CQLAI**は、Goで構築された高速でポータブルなCassandra(CQL)対話型ターミナルです。高度なターミナルUI、クライアント側コマンドパース、生産性向上機能を備えた、`cqlsh`の最新で使いやすい代替ツールを提供します。

**AI機能は完全にオプションです** - CQLAIはAI設定やAPIキーなしでも、スタンドアロンのCQLシェルとして完璧に動作します。

<div align="center">
  <video src="https://github.com/user-attachments/assets/334bd302-3152-4f48-9d2d-ed617e8d86d3" controls width="100%" style="max-width: 800px;">
    Your browser does not support the video tag.
  </video>
</div>

<div align="center">

### 🎁 100% 無料 & オープンソース
**隠れたコストなし • プレミアム層なし • ライセンスキー不要**

完全な透明性を持つコミュニティ主導の開発

</div>

元々のcqlshコマンドは[Apache Cassandra](https://cassandra.apache.org/)プロジェクトでPythonで書かれており、システムにPythonのインストールが必要です。cqlaiは単一の実行可能バイナリにコンパイルされ、外部依存関係を必要としません。このプロジェクトは以下のプラットフォーム向けのバイナリを提供しています:

- Linux x86-64
- macOS x86-64
- Windows x86-64
- Linux aarch64
- macOS arm64


美しいターミナルUIには[Bubble Tea](https://github.com/charmbracelet/bubbletea)、[Bubbles](https://github.com/charmbracelet/bubbles)、[Lip Gloss](https://github.com/charmbracelet/lipgloss)が使用されています。最新のCassandra機能を実装しているcassandra gocqlドライバーチーム[gocql](https://github.com/apache/cassandra-gocql-driver)に感謝します。

## 📑 目次

- [📊 プロジェクトの状況](#-プロジェクトの状況)
- [✨ 機能](#-機能)
- [🔧 インストール](#-インストール)
- [📚 使用方法](#-使用方法)
- [⚙️ 利用可能なコマンド](#️-利用可能なコマンド)
- [🛠️ 設定](#️-設定)
  - [設定の優先順位](#設定の優先順位)
  - [CQLSHRC互換性](#cqlshrc互換性)
  - [CQLAI JSON設定](#cqlai-json設定)
  - [AIプロバイダー設定](#aiプロバイダー設定)
    - [OpenAI](#openai-gpt-4-gpt-35)
    - [Anthropic](#anthropic-claude-3)
    - [Google Gemini](#google-gemini)
    - [Synthetic](#synthetic-複数のオープンソースモデル)
    - [Ollama](#ollama-ローカルモデル)
    - [OpenRouter](#openrouter-複数のモデル)
    - [Mockプロバイダー](#mockプロバイダーテスト用)
- [🤖 AI駆動のクエリ生成](#-ai駆動のクエリ生成)
- [📦 Apache Parquetサポート](#-apache-parquetサポート)
- [⚠️ 既知の制限事項](#️-既知の制限事項)
- [🔨 開発](#-開発)
- [🏗️ 技術スタック](#️-技術スタック)
- [🙏 謝辞](#-謝辞)
- [💬 コミュニティ & サポート](#-コミュニティ--サポート)
- [📝 ライセンス](#-ライセンス)
- [⚖️ 法的通知](#️-法的通知)

---

## 📊 プロジェクトの状況

**CQLAIは本番環境対応**であり、Cassandraクラスタを使用した開発、テスト、本番環境で活発に使用されています。このツールは、拡張機能とパフォーマンスを備えた`cqlsh`の完全で安定した代替品を提供します。

### 動作機能
- すべてのコアCQL操作とクエリ
- 完全なメタコマンドサポート(`DESCRIBE`、`SHOW`、`CONSISTENCY`など)
- クライアント側コマンドパース(軽量、ANTLRに依存しない)
- `COPY TO/FROM`によるデータインポート/エクスポート(CSVおよびParquet形式)
- SSL/TLS接続と認証
- ユーザー定義型(UDT)と複雑なデータ型
- スクリプトと自動化のためのバッチモード
- 効率的なデータ交換のためのApache Parquet形式サポート
- CQLキーワード、テーブル、カラム、キースペースのタブ補完
- **オプション**: AI駆動のクエリ生成([OpenAI](https://openai.com/)、[Anthropic](https://www.anthropic.com/)、[Google Gemini](https://ai.google.dev/)、[Synthetic](https://synthetic.new/))

### 近日公開予定
- AI コンテキスト認識の強化
- Cassandra MCPサービス
- 追加のパフォーマンス最適化

**今すぐCQLAIをお試しください**。開発にご協力いただければ幸いです！フィードバックと貢献は、CassandraコミュニティにとってベストなCQLシェルを作るために非常に貴重です。[問題を報告](https://github.com/axonops/cqlai/issues)するか、[貢献](https://github.com/axonops/cqlai/pulls)してください。

---

## ✨ 機能

- **対話型CQLシェル:** Cassandraクラスタがサポートする任意のCQLクエリを実行できます。
- **リッチターミナルUI:**
    - オルタネートスクリーンバッファを使用したマルチレイヤー・フルスクリーンターミナルアプリケーション(ターミナル履歴を保持)。
    - 自動データロード機能付きの仮想化スクロール可能テーブルで、大規模クエリによるメモリオーバーロードを防止。
    - vimスタイルのキーボードショートカットによる高度なナビゲーションモード。
    - ホイールスクロールとテキスト選択を含む完全なマウスサポート。
    - 接続詳細、クエリレイテンシ、セッションステータス(一貫性、トレース)を表示するスティッキーフッター/ステータスバー。
    - 履歴、ヘルプ、コマンド補完のためのモーダルオーバーレイ。
- **Apache Parquetサポート:**
    - 分析と機械学習ワークフロー向けの高性能カラムナーデータ形式。
    - `COPY TO`コマンドでCassandraテーブルをParquetファイルにエクスポート。
    - 自動スキーマ推論でParquetファイルをCassandraにインポート。
    - Hiveスタイルのディレクトリ構造を持つパーティション分割データセット。
    - インテリジェントな時間ベースパーティショニングのためのTimeUUID / timestamp仮想カラム。
    - UDT、コレクション、ベクトルを含むすべてのCassandraデータ型のサポート。
- **オプションのAI駆動クエリ生成:**
    - AIプロバイダー([OpenAI](https://openai.com/)、[Anthropic](https://www.anthropic.com/)、[Google Gemini](https://ai.google.dev/)、[Synthetic](https://synthetic.new/))を使用した自然言語からCQLへの変換。
    - 自動コンテキスト付きのスキーマ認識クエリ生成。
    - 実行前の安全なプレビューと確認。
    - DDLおよびDMLを含む複雑な操作のサポート。
    - **APIキーの設定が必要** - コア機能には不要です。
- **設定:**
    - 現在のディレクトリまたは`~/.cqlai.json`の`cqlai.json`による簡単な設定。
    - 証明書認証付きSSL/TLS接続のサポート。
- **単一バイナリ:** 外部依存関係なしの単一静的バイナリとして配布。高速起動と小さなフットプリント。

## 🔧 インストール

`cqlai`はいくつかの方法でインストールできます。パッケージマネージャー(APT、YUM)やDockerを含む詳細な手順については、[インストールガイド](docs/INSTALLATION_jp.md)を参照してください。

### プリコンパイルバイナリ

お使いのOSとアーキテクチャに適したバイナリを[**Releases**](https://github.com/axonops/cqlai/releases)ページからダウンロードしてください。


### Goを使用する

```bash
go install github.com/axonops/cqlai/cmd/cqlai@latest
```

### ソースから

```bash
git clone https://github.com/axonops/cqlai.git
cd cqlai
go build -o cqlai cmd/cqlai/main.go
```

### Dockerを使用する

```bash
# イメージをビルド
docker build -t cqlai .

# コンテナを実行
docker run -it --rm --name cqlai-session cqlai --host your-cassandra-host
```

## 📚 使用方法

### 対話モード

Cassandraホストに接続:
```bash
# コマンドラインでパスワードを指定(推奨されません - psで表示されます)
cqlai --host 127.0.0.1 --port 9042 --username cassandra --password cassandra

# パスワードプロンプトを使用(安全 - パスワードは隠されます)
cqlai --host 127.0.0.1 --port 9042 -u cassandra
# Password: [hidden input]

# 環境変数を使用(スクリプト/コンテナ向けに安全)
export CQLAI_PASSWORD=cassandra
cqlai --host 127.0.0.1 -u cassandra
```

または設定ファイルを使用:
```bash
# サンプルから設定を作成
cp cqlai.json.example cqlai.json
# cqlai.jsonを編集して設定を変更し、実行:
cqlai
```

### コマンドラインオプション

```bash
cqlai [options]
```

#### 接続オプション
| オプション | 短縮 | 説明 |
|--------|-------|-------------|
| `--host <host>` | | Cassandraホスト(設定を上書き) |
| `--port <port>` | | Cassandraポート(設定を上書き) |
| `--keyspace <keyspace>` | `-k` | デフォルトのキースペース(設定を上書き) |
| `--username <username>` | `-u` | 認証用のユーザー名 |
| `--password <password>` | `-p` | 認証用のパスワード* |
| `--no-confirm` | | 破壊的コマンド(DROP、DELETE、TRUNCATE)の確認プロンプトを無効化 |
| `--connect-timeout <seconds>` | | 接続タイムアウト(デフォルト: 10) |
| `--request-timeout <seconds>` | | リクエストタイムアウト(デフォルト: 10) |
| `--debug` | | デバッグログを有効化 |

*\*注意: パスワードは3つの方法で提供できます:*
1. *`-p`でコマンドライン指定(推奨されません - プロセスリストに表示されます)*
2. *`-u`を`-p`なしで使用した場合の対話型プロンプト(推奨)*
3. *環境変数`CQLAI_PASSWORD`(自動化に適しています)*

#### バッチモードオプション
| オプション | 短縮 | 説明 |
|--------|-------|-------------|
| `--execute <statement>` | `-e` | CQLステートメントを実行して終了 |
| `--file <file>` | `-f` | ファイルからCQLを実行して終了 |
| `--format <format>` | | 出力形式: ascii, json, csv, table |
| `--no-header` | | カラムヘッダーを出力しない(CSV) |
| `--field-separator <sep>` | | CSVのフィールド区切り文字(デフォルト: ,) |
| `--page-size <n>` | | バッチあたりの行数(デフォルト: 100) |

#### 一般オプション
| オプション | 短縮 | 説明 |
|--------|-------|-------------|
| `--help` | `-h` | ヘルプメッセージを表示 |
| `--version` | `-v` | バージョンを表示して終了 |

### バッチモードの例

CQLステートメントを非対話的に実行(cqlshと互換性があります):

```bash
# 単一のステートメントを実行
cqlai -e "SELECT * FROM system_schema.keyspaces;"

# ファイルから実行
cqlai -f script.cql

# パイプ入力
echo "SELECT * FROM users;" | cqlai

# 出力形式を制御
cqlai -e "SELECT * FROM users;" --format json
cqlai -e "SELECT * FROM users;" --format csv --no-header

# ページサイズを制御
cqlai -e "SELECT * FROM large_table;" --page-size 50
```

### 基本コマンド

- **CQLを実行:** 任意のCQLステートメントを入力してEnterキーを押します。
- **メタコマンド:**
  ```sql
  DESCRIBE KEYSPACES;
  USE my_keyspace;
  DESCRIBE TABLES;
  CONSISTENCY QUORUM;
  TRACING ON;
  PAGING 50;
  EXPAND ON;  -- 垂直出力モード
  SOURCE 'script.cql';  -- CQLスクリプトを実行
  ```
- **AI駆動のクエリ生成:**
  ```sql
  .ai What keyspaces are there?
  .ai What columns does the users table have?
  .ai create a table for storing product inventory
  .ai delete orders older than 1 year from the orders table
  ```

### キーボードショートカット

#### ナビゲーション＆コントロール
| ショートカット | アクション | macOS代替 |
|----------|--------|-------------------|
| `↑`/`↓` | コマンド履歴をナビゲート | 同じ |
| `Ctrl+P`/`Ctrl+N` | コマンド履歴の前/次 | 同じ |
| `Alt+N` | 履歴の次の行に移動 | `Option+N` |
| `Tab` | コマンドとテーブル/キースペース名の自動補完 | 同じ |
| `Ctrl+C` | 入力をクリア / ページネーションをキャンセル / 操作をキャンセル(2回で終了) | `⌘+C` または `Ctrl+C` |
| `Ctrl+D` | アプリケーションを終了 | `⌘+D` または `Ctrl+D` |
| `Ctrl+R` | コマンド履歴を検索 | `⌘+R` または `Ctrl+R` |
| `Esc` | ナビゲーションモードを切り替え / ページネーションをキャンセル / モーダルを閉じる | 同じ |
| `Enter` | コマンドを実行 / 次のページをロード(ページネーション中) | 同じ |

#### テキスト編集
| ショートカット | アクション | macOS代替 |
|----------|--------|-------------------|
| `Ctrl+A` | 行の先頭にジャンプ | 同じ |
| `Ctrl+E` | 行の末尾にジャンプ | 同じ |
| `Ctrl+Left`/`Ctrl+Right` | 単語ごとにジャンプ(または20文字) | 同じ |
| `PgUp`/`PgDn` (入力時) | 長いクエリで左/右にページ移動 | `Fn+↑`/`Fn+↓` |
| `Ctrl+K` | カーソルから行末までカット | 同じ |
| `Ctrl+U` | 行頭からカーソルまでカット | 同じ |
| `Ctrl+W` | 単語を後方にカット | 同じ |
| `Alt+D` | 単語を前方に削除 | `Option+D` |
| `Ctrl+Y` | 以前にカットしたテキストを貼り付け | 同じ |

#### ビュー切り替え
| ショートカット | アクション |
|----------|--------|
| `F2` | クエリ/履歴ビューに切り替え |
| `F3` | テーブルビューに切り替え |
| `F4` | トレースビューに切り替え(トレース有効時) |
| `F5` | AI会話ビューに切り替え |
| `F6` | テーブルヘッダーでカラムデータ型の表示を切り替え |

#### スクロール＆テーブルナビゲーション
| ショートカット | アクション | macOS代替 |
|----------|--------|-------------------|
| `PgUp`/`PgDn` | ビューポートをページ単位でスクロール / 利用可能な場合はデータをロード | `Fn+↑`/`Fn+↓` |
| `Space` | より多くのデータが利用可能な場合は次のページをロード | 同じ |
| `Enter` (空の入力) | より多くのデータが利用可能な場合は次のページをロード | 同じ |
| `Alt+↑`/`Alt+↓` | ビューポートを1行ずつスクロール(行境界を尊重) | `Option+↑`/`Option+↓` |
| `Alt+←`/`Alt+→` | テーブルを水平スクロール(幅広いテーブル) | `Option+←`/`Option+→` |
| `↑`/`↓` | テーブル行をナビゲート(ナビゲーションモード時) | 同じ |

#### ナビゲーションモード(テーブル/トレースビュー)
テーブルまたはトレースを表示しているときに`Esc`を押してナビゲーションモードを切り替えます。

| ショートカット | ナビゲーションモードでのアクション |
|----------|---------------------------|
| `j` / `k` | 1行ずつ下/上にスクロール |
| `d` / `u` | 半ページずつ下/上にスクロール |
| `g` / `G` | 結果の先頭/末尾にジャンプ |
| `<` / `>` | 10カラムずつ左/右にスクロール |
| `{` / `}` | 50カラムずつ左/右にスクロール |
| `0` / `$` | 最初/最後のカラムにジャンプ |
| `Esc` | ナビゲーションモードを終了 / ページネーションがアクティブな場合はキャンセル |

#### マウスサポート
| アクション | 機能 |
|--------|----------|
| マウスホイール | 自動データロード付きで垂直スクロール |
| Alt+マウスホイール | テーブル内で水平スクロール |
| Shift+マウスホイール | 水平スクロール(代替) |
| Ctrl+マウスホイール | 水平スクロール(代替) |
| Shift+クリック+ドラッグ | コピー用のテキスト選択 |
| Ctrl+Shift+C | 選択したテキストをクリップボードにコピー |
| ミドルクリック | 選択バッファから貼り付け(Linux/Unix) |

**macOSユーザーへの注意:**
- ほとんどの`Ctrl`ショートカットはmacOSでもそのまま動作しますが、代替として`⌘`(Command)キーも使用できます
- `Alt`キーはMacキーボードでは`Option`と表示されます
- ファンクションキー(F1-F6)は、Macの設定によっては`Fn`キーを押す必要がある場合があります

### タブ補完

CQLAIは、ワークフローを高速化するためのインテリジェントなコンテキスト対応タブ補完を提供します。任意のポイントで`Tab`を押すと、利用可能な補完が表示されます。

#### 補完可能なもの

**CQLキーワード＆コマンド:**
- すべてのCQLキーワード: `SELECT`、`INSERT`、`CREATE`、`ALTER`、`DROP`など
- メタコマンド: `DESCRIBE`、`CONSISTENCY`、`COPY`、`SHOW`など
- データ型: `TEXT`、`INT`、`UUID`、`TIMESTAMP`など
- 一貫性レベル: `ONE`、`QUORUM`、`ALL`、`LOCAL_QUORUM`など

**スキーマオブジェクト:**
- キースペース名
- テーブル名(現在のキースペース内)
- カラム名(コンテキストが許可する場合)
- ユーザー定義型名
- 関数と集約名
- インデックス名

**コンテキスト対応補完:**
```sql
-- SELECT後、カラム名とキーワードを提案
SELECT <Tab>           -- 表示: *, カラム名, DISTINCT, JSON など

-- FROM後、テーブル名を提案
SELECT * FROM <Tab>    -- 表示: 現在のキースペース内の利用可能なテーブル

-- USE後、キースペース名を提案
USE <Tab>              -- 表示: 利用可能なキースペース

-- DESCRIBE後、オブジェクトタイプを提案
DESCRIBE <Tab>         -- 表示: KEYSPACE, TABLE, TYPE など

-- 一貫性コマンド後
CONSISTENCY <Tab>      -- 表示: ONE, QUORUM, ALL など
```

**ファイルパス補完:**
```sql
-- ファイルパスを受け入れるコマンド用
SOURCE '<Tab>          -- 表示: 現在のディレクトリのファイル
SOURCE '/path/<Tab>    -- 表示: /path/のファイル
```

#### 補完動作

- **大文字小文字を区別しない:** `sel<Tab>`と入力すると`SELECT`が得られます
- **部分マッチ:** 単語の一部を入力してTabを押します
- **複数のマッチ:** 複数の補完が利用可能な場合:
  - 最初のTab: 一意の場合はインライン補完を表示
  - 2回目のTab: モーダルですべての利用可能なオプションを表示
- **スマートフィルタリング:** 補完は現在のコンテキストに基づいてフィルタリングされます
- **Escでキャンセル:** `Esc`を押して補完モーダルを閉じます

#### 例

```sql
-- テーブル名を補完
SELECT * FROM us<Tab>
-- 補完結果: SELECT * FROM users

-- 一貫性レベルを補完
CONSISTENCY LOC<Tab>
-- 表示: LOCAL_ONE, LOCAL_QUORUM, LOCAL_SERIAL

-- SELECT後のカラム名を補完
SELECT id, na<Tab> FROM users
-- 補完結果: SELECT id, name FROM users

-- SOURCEコマンドのファイルパスを補完
SOURCE 'sche<Tab>
-- 補完結果: SOURCE 'schema.cql'

-- COPYコマンドオプションを補完
COPY users TO 'file.csv' WITH <Tab>
-- 表示: HEADER, DELIMITER, NULLVAL, PAGESIZE など

-- 複数のテーブルが存在する場合にすべてのテーブルを表示
SELECT * FROM <Tab>
-- モーダルで表示: users, orders, products など
```

#### 効果的な使用のためのヒント

1. **Tabを自由に使う:** 補完システムはスマートでコンテキスト対応です
2. **最小限の文字を入力:** 多くの場合、2〜3文字で一意の補完が得られます
3. **発見に使用:** 空の入力でTabを押すと、利用可能なものが表示されます
4. **ファイルパス:** ファイルパス補完には引用符を含めることを忘れないでください
5. **補完のナビゲート:** 矢印キーを使用して複数のオプションから選択します

## ⚙️ 利用可能なコマンド

CQLAIは、拡張機能のための追加のメタコマンドに加えて、すべての標準CQLコマンドをサポートしています。

### CQLコマンド
Cassandraクラスタがサポートする任意の有効なCQLステートメントを実行:
- DDL: `CREATE`、`ALTER`、`DROP`(KEYSPACE、TABLE、INDEXなど)
- DML: `SELECT`、`INSERT`、`UPDATE`、`DELETE`
- DCL: `GRANT`、`REVOKE`
- その他: `USE`、`TRUNCATE`、`BEGIN BATCH`など

### メタコマンド

メタコマンドは、標準CQLを超える追加機能を提供します:

#### セッション管理
- **CONSISTENCY** `<level>` - 一貫性レベルを設定(ONE、QUORUM、ALLなど)
  ```sql
  CONSISTENCY QUORUM
  CONSISTENCY LOCAL_ONE
  ```

- **PAGING** `<size>` | OFF - 結果のページサイズを設定
  ```sql
  PAGING 1000
  PAGING OFF
  ```

- **TRACING** ON | OFF - クエリトレースを有効/無効化
  ```sql
  TRACING ON
  SELECT * FROM users;
  TRACING OFF
  ```

- **OUTPUT** [FORMAT] - 出力形式を設定
  ```sql
  OUTPUT          -- 現在の形式を表示
  OUTPUT TABLE    -- テーブル形式(デフォルト)
  OUTPUT JSON     -- JSON形式
  OUTPUT EXPAND   -- 拡張垂直形式
  OUTPUT ASCII    -- ASCIIテーブル形式
  ```

#### スキーマ記述
- **DESCRIBE** - スキーマ情報を表示
  ```sql
  DESCRIBE KEYSPACES                    -- すべてのキースペースをリスト
  DESCRIBE KEYSPACE <name>              -- キースペース定義を表示
  DESCRIBE TABLES                       -- 現在のキースペースのテーブルをリスト
  DESCRIBE TABLE <name>                 -- テーブル構造を表示
  DESCRIBE TYPES                        -- ユーザー定義型をリスト
  DESCRIBE TYPE <name>                  -- UDT定義を表示
  DESCRIBE FUNCTIONS                    -- ユーザー関数をリスト
  DESCRIBE FUNCTION <name>              -- 関数定義を表示
  DESCRIBE AGGREGATES                   -- ユーザー集約をリスト
  DESCRIBE AGGREGATE <name>             -- 集約定義を表示
  DESCRIBE MATERIALIZED VIEWS           -- マテリアライズドビューをリスト
  DESCRIBE MATERIALIZED VIEW <name>     -- ビュー定義を表示
  DESCRIBE INDEX <name>                 -- インデックス定義を表示
  DESCRIBE CLUSTER                      -- クラスタ情報を表示
  DESC <keyspace>.<table>               -- テーブル記述の短縮形
  ```

#### データエクスポート/インポート
- **COPY TO** - テーブルデータをCSVまたはParquetファイルにエクスポート
  ```sql
  -- CSVへの基本的なエクスポート
  COPY users TO 'users.csv'

  -- Parquet形式へのエクスポート(拡張子から自動検出)
  COPY users TO 'users.parquet'

  -- 明示的な形式と圧縮でParquetにエクスポート
  COPY users TO 'data.parquet' WITH FORMAT='PARQUET' AND COMPRESSION='SNAPPY'

  -- 特定のカラムをエクスポート
  COPY users (id, name, email) TO 'users_partial.csv'

  -- オプション付きでエクスポート
  COPY users TO 'users.csv' WITH HEADER = TRUE AND DELIMITER = '|'

  -- 標準出力にエクスポート
  COPY users TO STDOUT WITH HEADER = TRUE

  -- 利用可能なオプション:
  -- FORMAT = 'CSV'/'PARQUET' -- 出力形式(デフォルト: CSV、自動検出)
  -- HEADER = TRUE/FALSE      -- カラムヘッダーを含める(CSVのみ)
  -- DELIMITER = ','          -- フィールド区切り文字(CSVのみ)
  -- NULLVAL = 'NULL'        -- NULL値に使用する文字列
  -- PAGESIZE = 1000         -- 大規模エクスポートのページあたりの行数
  -- COMPRESSION = 'SNAPPY'  -- Parquet用: SNAPPY, GZIP, ZSTD, LZ4, NONE
  -- CHUNKSIZE = 10000       -- Parquetのチャンクあたりの行数
  ```

- **COPY FROM** - CSVまたはParquetデータをテーブルにインポート
  ```sql
  -- CSVファイルからの基本的なインポート
  COPY users FROM 'users.csv'

  -- Parquetファイルからのインポート(自動検出)
  COPY users FROM 'users.parquet'

  -- 明示的な形式でParquetからインポート
  COPY users FROM 'data.parquet' WITH FORMAT='PARQUET'

  -- ヘッダー行付きでインポート(CSV)
  COPY users FROM 'users.csv' WITH HEADER = TRUE

  -- 特定のカラムをインポート
  COPY users (id, name, email) FROM 'users_partial.csv'

  -- 標準入力からインポート
  COPY users FROM STDIN

  -- カスタムオプション付きでインポート
  COPY users FROM 'users.csv' WITH HEADER = TRUE AND DELIMITER = '|' AND NULLVAL = 'N/A'

  -- 利用可能なオプション:
  -- HEADER = TRUE/FALSE      -- 最初の行にカラム名が含まれる
  -- DELIMITER = ','          -- フィールド区切り文字
  -- NULLVAL = 'NULL'        -- NULL値を表す文字列
  -- MAXROWS = -1            -- インポートする最大行数(-1 = 無制限)
  -- SKIPROWS = 0            -- スキップする初期行数
  -- MAXPARSEERRORS = -1     -- 許容される最大パースエラー数(-1 = 無制限)
  -- MAXINSERTERRORS = 1000  -- 許容される最大挿入エラー数
  -- MAXBATCHSIZE = 20       -- バッチ挿入あたりの最大行数
  -- MAXREQUESTS = 6         -- 並行バッチワーカー数（並列処理）
  -- MINBATCHSIZE = 2        -- バッチ挿入あたりの最小行数
  -- CHUNKSIZE = 5000        -- 進捗更新間の行数
  -- ENCODING = 'UTF8'       -- ファイルエンコーディング
  -- QUOTE = '"'             -- 文字列の引用文字
  ```

- **CAPTURE** - クエリ出力をファイルにキャプチャ(連続記録)
  ```sql
  CAPTURE 'output.txt'          -- テキストファイルへのキャプチャを開始
  CAPTURE JSON 'output.json'    -- JSONとしてキャプチャ
  CAPTURE CSV 'output.csv'      -- CSVとしてキャプチャ
  SELECT * FROM users;
  CAPTURE OFF                   -- キャプチャを停止
  ```

- **SAVE** - 表示されたクエリ結果をファイルに保存(再実行なし)
  ```sql
  -- まずクエリを実行
  SELECT * FROM users WHERE status = 'active';

  -- 次に表示された結果をさまざまな形式で保存:
  SAVE                           -- 対話ダイアログ(形式とファイル名を選択)
  SAVE 'users.csv'               -- CSVに保存(形式は自動検出)
  SAVE 'users.json'              -- JSONに保存(形式は自動検出)
  SAVE 'users.txt' ASCII         -- ASCIIテーブルとして保存
  SAVE 'data.csv' CSV            -- 明示的に形式を指定

  -- CAPTUREとの主な違い:
  -- - SAVEは現在表示されている結果をエクスポート
  -- - クエリを再実行する必要なし
  -- - ターミナルに表示されているデータをそのまま保持
  -- - ページ分割された結果でも動作(ロードされたページのみ保存)
  ```

#### 情報表示
- **SHOW** - セッション情報を表示
  ```sql
  SHOW VERSION          -- Cassandraバージョンを表示
  SHOW HOST            -- 現在の接続詳細を表示
  SHOW SESSION         -- すべてのセッション設定を表示
  ```

- **EXPAND** ON | OFF - 拡張出力モードを切り替え
  ```sql
  EXPAND ON            -- 垂直出力(1行に1フィールド)
  SELECT * FROM users WHERE id = 1;
  EXPAND OFF           -- 通常のテーブル出力
  ```

#### スクリプト実行
- **SOURCE** - ファイルからCQLスクリプトを実行
  ```sql
  SOURCE 'schema.cql'           -- スクリプトを実行
  SOURCE '/path/to/script.cql'  -- 絶対パス
  ```

#### ヘルプ
- **HELP** - コマンドヘルプを表示
  ```sql
  HELP                 -- すべてのコマンドを表示
  HELP DESCRIBE        -- 特定のコマンドのヘルプ
  HELP CONSISTENCY     -- 一貫性レベルのヘルプ
  ```

### AIコマンド
- **.ai** `<natural language query>` - 自然言語からCQLを生成
  ```sql
  .ai show all users with active status
  .ai create a table for storing user sessions
  .ai find orders placed in the last 30 days
  ```

## 🛠️ 設定

CQLAIは、既存のCassandraセットアップとの最大限の柔軟性と互換性のために、複数の設定方法をサポートしています。

### 設定の優先順位

設定ソースは次の順序でロードされます(後のソースが前のソースを上書きします):

1. **CQLSHRCファイル**(既存のcqlshセットアップとの互換性のため)
   - `~/.cassandra/cqlshrc`(標準の場所)
   - `~/.cqlshrc`(代替の場所)
   - `$CQLSH_RC`(環境変数が設定されている場合)

2. **CQLAI JSON設定ファイル**
   - `./cqlai.json`(現在のディレクトリ)
   - `~/.cqlai.json`(ユーザーホームディレクトリ)
   - `~/.config/cqlai/config.json`(XDG設定ディレクトリ)

3. **環境変数**
   - `CQLAI_HOST`、`CQLAI_PORT`、`CQLAI_KEYSPACE`など
   - `CASSANDRA_HOST`、`CASSANDRA_PORT`(互換性のため)

4. **コマンドラインフラグ**(最高優先度)
   - `--host`、`--port`、`--keyspace`、`--username`、`--password`など

### CQLSHRC互換性

CQLAIは、従来の`cqlsh`ツールで使用される標準のCQLSHRCファイルを読み取ることができ、移行をシームレスにします。

**サポートされるCQLSHRCセクション:**
- `[connection]` - hostname、port、ssl設定
- `[authentication]` - keyspace、認証情報ファイルパス
- `[auth_provider]` - 認証モジュールとusername
- `[ssl]` - SSL/TLS証明書設定

**CQLSHRCファイルの例:**
```ini
; ~/.cassandra/cqlshrc
[connection]
hostname = cassandra.example.com
port = 9042
ssl = true

[authentication]
keyspace = my_keyspace
credentials = ~/.cassandra/credentials

[ssl]
certfile = ~/certs/ca.pem
userkey = ~/certs/client-key.pem
usercert = ~/certs/client-cert.pem
validate = true
```

完全なCQLSHRC互換性の詳細については、[CQLSHRC_SUPPORT.md](docs/CQLSHRC_SUPPORT_jp.md)を参照してください。

### CQLAI JSON設定

高度な機能とAI設定のために、CQLAIは独自のJSON形式を使用します:

**`cqlai.json`の例:**
```json
{
  "host": "127.0.0.1",
  "port": 9042,
  "keyspace": "",
  "username": "cassandra",
  "password": "cassandra",
  "requireConfirmation": true,
  "consistency": "LOCAL_ONE",
  "pageSize": 100,
  "maxMemoryMB": 10,
  "connectTimeout": 10,
  "requestTimeout": 10,
  "debug": false,
  "historyFile": "~/.cqlai/history",
  "aiHistoryFile": "~/.cqlai/ai_history",
  "ssl": {
    "enabled": false,
    "certPath": "/path/to/client-cert.pem",
    "keyPath": "/path/to/client-key.pem",
    "caPath": "/path/to/ca-cert.pem",
    "hostVerification": true,
    "insecureSkipVerify": false
  },
  "ai": {
    "provider": "openai",
    "apiKey": "sk-...",
    "model": "gpt-4-turbo-preview"
  }
}
```

**注意:** `url`フィールドを使用してOpenAI互換APIのAPIエンドポイントを上書きできます:
```json
{
  "ai": {
    "provider": "openai",
    "apiKey": "your-api-key",
    "url": "https://api.synthetic.new/openai/v1",
    "model": "hf:Qwen/Qwen3-235B-A22B-Instruct-2507"
  }
}
```

**設定オプション:**

| オプション | 型 | デフォルト | 説明 |
|--------|------|---------|-------------|
| `host` | string | `127.0.0.1` | Cassandraホストアドレス |
| `port` | number | `9042` | Cassandraポート |
| `keyspace` | string | `""` | 使用するデフォルトキースペース |
| `username` | string | `""` | 認証ユーザー名 |
| `password` | string | `""` | 認証パスワード |
| `requireConfirmation` | boolean | `true` | 破壊的コマンド(DROP、DELETE、TRUNCATE)の確認を要求 |
| `consistency` | string | `LOCAL_ONE` | デフォルトの一貫性レベル (ANY, ONE, TWO, THREE, QUORUM, ALL, LOCAL_QUORUM, EACH_QUORUM, LOCAL_ONE) |
| `pageSize` | number | `100` | ページあたりの行数 |
| `maxMemoryMB` | number | `10` | クエリ結果の最大メモリ(MB) |
| `connectTimeout` | number | `10` | 接続タイムアウト(秒) |
| `requestTimeout` | number | `10` | リクエストタイムアウト(秒) |
| `historyFile` | string | `~/.cqlai/history` | CQLコマンド履歴ファイルのパス(`~`展開をサポート) |
| `aiHistoryFile` | string | `~/.cqlai/ai_history` | AIコマンド履歴ファイルのパス(`~`展開をサポート) |
| `debug` | boolean | `false` | デバッグログを有効化 |

### 設定ファイルの場所

CQLAIは次の場所で設定ファイルを検索します:

**CQLSHRCファイル:**
1. `$CQLSH_RC`(環境変数が設定されている場合)
2. `~/.cassandra/cqlshrc`(標準のcqlshの場所)
3. `~/.cqlshrc`(代替の場所)

**CQLAI JSONファイル:**
1. `./cqlai.json`(現在の作業ディレクトリ)
2. `~/.cqlai.json`(ユーザーホームディレクトリ)
3. `~/.config/cqlai/config.json`(Linux/macOSのXDG設定ディレクトリ)

### 環境変数

一般的な環境変数:
- `CQLAI_HOST`または`CASSANDRA_HOST` - Cassandraホスト
- `CQLAI_PORT`または`CASSANDRA_PORT` - Cassandraポート
- `CQLAI_KEYSPACE` - デフォルトのキースペース
- `CQLAI_USERNAME` - 認証ユーザー名
- `CQLAI_PASSWORD` - 認証パスワード
- `CQLAI_PAGE_SIZE` - バッチモードのページサイズ(デフォルト: 100)
- `CQLAI_NO_CONFIRM` - `true`または`1`に設定して破壊的コマンドの確認プロンプトを無効化
- `CQLSH_RC` - カスタムCQLSHRCファイルへのパス

### cqlshからの移行

`cqlsh`から移行する場合、CQLAIは既存の`~/.cassandra/cqlshrc`ファイルを自動的に読み取ります。既存のCassandra設定でCQLAIの使用を開始するための変更は必要ありません。

## 🤖 AI駆動のクエリ生成

CQLAIには、自然言語をCQLクエリに変換する組み込みのAI機能が含まれています。リクエストの前に`.ai`を付けるだけです:

### 例

```sql
-- シンプルなクエリ
.ai show all users
.ai find products with price less than 100
.ai count orders from last month

-- 複雑な操作
.ai create a table for storing customer feedback with id, customer_id, rating, and comment
.ai update user status to inactive where last_login is older than 90 days
.ai delete all expired sessions

-- スキーマ探索
.ai what tables are in this keyspace
.ai describe the structure of the users table
```

### 動作の仕組み

1. **自然言語入力**: `.ai`に続けてプレーンな英語でリクエストを入力
2. **スキーマコンテキスト**: CQLAIは自動的に現在のスキーマを抽出してコンテキストを提供
3. **クエリ生成**: AIが構造化されたクエリプランを生成
4. **プレビュー＆確認**: 実行前に生成されたCQLを確認
5. **実行または編集**: クエリを実行、編集、またはキャンセルすることを選択

### サポートされるAIプロバイダー

`cqlai.json`で優先するAIプロバイダーを設定:

- **[OpenAI](https://openai.com/)** (GPT-4、GPT-3.5)
- **[Anthropic](https://www.anthropic.com/)** (Claude 3)
- **[Google Gemini](https://ai.google.dev/)**
- **[Synthetic](https://synthetic.new/)** (複数のオープンソースモデル)
- **[Ollama](https://ollama.ai/)** (ローカルモデルまたはOpenAI互換API)
- **[OpenRouter](https://openrouter.ai/)** (複数のモデルへのアクセス)
- **Mock** (デフォルト、APIキーなしでのテスト用)

#### Synthetic (複数のオープンソースモデル)

Syntheticを使用して、非常に合理的な価格で多数のオープンソースAIモデルにアクセスできます。SyntheticはOpenAI互換APIを提供し、さまざまなオープンソースモデルとの作業を容易にします。

- **開始:** [synthetic.new](https://synthetic.new/)
- **APIドキュメント:** [dev.synthetic.new/docs](https://dev.synthetic.new/docs)
- **推奨モデル:**
  - `hf:Qwen/Qwen3-235B-A22B-Instruct-2507` (推奨、ただしすべてのモデルを広範にテストしたわけではありません)
- **利用可能なモデル:** [Always-On Models](https://dev.synthetic.new/docs/api/models#always-on-models)を参照

**設定:**
```json
{
  "ai": {
    "provider": "openai",
    "apiKey": "your-synthetic-api-key",
    "url": "https://api.synthetic.new/openai/v1",
    "model": "hf:Qwen/Qwen3-235B-A22B-Instruct-2507"
  }
}
```

**主な利点:**
- さまざまなオープンソースモデルへのアクセス
- 費用対効果の高い価格設定
- 簡単な統合のためのOpenAI互換API
- ベンダーロックインなし

**注意:**
- Syntheticは OpenAI 互換インターフェースを提供するため、設定では `openai` プロバイダーを使用します
- `url` フィールドはデフォルトのOpenAIエンドポイントをSyntheticに向けるように上書きします
- APIキーが必要です - [synthetic.new](https://synthetic.new/)から取得してください

### 安全機能

- **デフォルトで読み取り専用**: AIは明示的に変更を求められない限り、SELECTクエリを優先
- **危険な操作の警告**: DROP、DELETE、TRUNCATE操作は警告を表示
- **確認が必要**: 破壊的な操作には追加の確認が必要
- **スキーマ検証**: クエリは現在のスキーマに対して検証されます

### 確認プロンプトの無効化

自動化やスクリプト用に、破壊的コマンド(DROP、DELETE、TRUNCATE)の確認プロンプトを無効にする方法:

1. **コマンドラインフラグ**:
   ```bash
   cqlai --no-confirm -e "TRUNCATE my_table;"
   ```

2. **環境変数**:
   ```bash
   export CQLAI_NO_CONFIRM=true
   cqlai -e "DROP TABLE old_data;"
   ```

3. **設定ファイル** (`cqlai.json`):
   ```json
   {
     "requireConfirmation": false
   }
   ```

**注意**: 本番環境では注意して使用してください。これらの設定は、偶発的なデータ損失を防ぐための安全プロンプトを無効にします。

## 📦 Apache Parquetサポート

CQLAIは、Apache Parquet形式の包括的なサポートを提供し、モダンなデータエコシステムとの統合に最適です。

### 主な利点

- **効率的なストレージ**: 優れた圧縮を備えたカラムナー形式(CSVより50〜80%小さい)
- **高速分析**: Spark、Presto、その他のエンジンでの分析クエリ用に最適化
- **型の保持**: コレクションとUDTを含むCassandraデータ型を維持
- **機械学習対応**: pandas、PyArrow、MLフレームワークと直接互換
- **ストリーミングサポート**: 大規模データセット用のメモリ効率の良いストリーミング

### クイック例

```sql
-- Parquetにエクスポート(拡張子から自動検出)
COPY users TO 'users.parquet';

-- 圧縮付きでエクスポート
COPY events TO 'events.parquet' WITH FORMAT='PARQUET' AND COMPRESSION='ZSTD';

-- Parquetからインポート
COPY users FROM 'users.parquet';

-- Parquet形式でクエリ結果をキャプチャ
CAPTURE 'results.parquet' FORMAT='PARQUET';
SELECT * FROM large_table WHERE condition = true;
CAPTURE OFF;
```

### サポートされる機能

- すべてのCassandraプリミティブ型(int、text、timestamp、uuidなど)
- コレクション型(list、set、map)
- ユーザー定義型(UDT)
- フローズンコレクション
- MLワークロード用のベクトル型(Cassandra 5.0+)
- 複数の圧縮アルゴリズム(Snappy、GZIP、ZSTD、LZ4)

詳細なドキュメントについては、[Parquetサポートガイド](docs/PARQUET_jp.md)を参照してください。

## ⚠️ 既知の制限事項

### JSON出力(CAPTURE JSONと--format json)

データをJSONとして出力する場合、基礎となるgocqlドライバーが動的型を処理する方法により、いくつかの制限があります:

#### NULL値
- **問題**: プリミティブカラム(int、boolean、textなど)のNULL値が`null`ではなくゼロ値(`0`、`false`、`""`)として表示されます
- **原因**: gocqlドライバーは、動的型(`interface{}`)にスキャンする際にNULLに対してゼロ値を返します
- **回避策**: `SELECT JSON`クエリを使用して、Cassandraサーバー側から適切なJSONを返します

#### ユーザー定義型(UDT)
- **問題**: JSON出力でUDTカラムが空のオブジェクト`{}`として表示されます
- **原因**: gocqlドライバーは、コンパイル時にその構造を知らないとUDTを適切にアンマーシャルできません
- **回避策**: 適切なUDTシリアライゼーションには`SELECT JSON`クエリを使用します

#### 例
```sql
-- 通常のSELECT(制限があります)
SELECT * FROM users;
-- 返却: {"id": 1, "age": 0, "active": false}  -- ageとactiveはNULLの可能性があります

-- SELECT JSONを使用(型を正しく保持)
SELECT JSON * FROM users;
-- 返却: {"id": 1, "age": null, "active": null}  -- NULLが適切に表現されます
```

**注意**: 複雑な型(list、set、map、vector)はJSON出力で適切に保持されます。

## 🔨 開発

`cqlai`で作業するには、Go(≥ 1.24)が必要です。

#### セットアップ

```bash
# リポジトリをクローン
git clone https://github.com/axonops/cqlai.git
cd cqlai

# 依存関係をインストール
go mod download
```

#### ビルド

```bash
# 標準バイナリをビルド
make build

# レース検出付きの開発バイナリをビルド
make build-dev
```

#### テストとリンターの実行

```bash
# すべてのテストを実行
make test

# カバレッジレポート付きでテストを実行
make test-coverage

# リンターを実行
make lint

# すべてのチェックを実行(format、lint、test)
make check
```


## 🏗️ 技術スタック

- **言語:** Go
- **TUIフレームワーク:** [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **TUIコンポーネント:** [Bubbles](https://github.com/charmbracelet/bubbles)
- **スタイリング:** [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Cassandraドライバー:** [gocql](https://github.com/gocql/gocql)

## 🙏 謝辞

CQLAIは、特にApache Cassandraをはじめとする複数のオープンソースプロジェクトの基盤の上に構築されています。分散データベースの分野への優れた仕事と貢献に対して、Apache Cassandraコミュニティに心から感謝いたします。

Apache Cassandraは、無料でオープンソースの分散型ワイドカラムストアNoSQLデータベース管理システムで、多数の汎用サーバー上で大量のデータを処理するように設計されており、単一障害点のない高可用性を提供します。

### Apache Cassandraリソース

- **公式ウェブサイト**: [cassandra.apache.org](https://cassandra.apache.org/)
- **ソースコード**: [GitHub](https://github.com/apache/cassandra)またはApache Gitリポジトリ（`gitbox.apache.org/repos/asf/cassandra.git`）で利用可能
- **ドキュメント**: [Apache Cassandraウェブサイト](https://cassandra.apache.org/)で包括的なガイドとリファレンスが利用可能

CQLAIは、さまざまなCassandraツールとユーティリティの機能を組み込み拡張し、CassandraデベロッパーとDBAに最新で効率的なターミナル体験を提供するために強化しています。

ユーザーには、メインのApache Cassandraプロジェクトを探索して貢献することをお勧めします。また、[GitHubディスカッション](https://github.com/axonops/cqlai/discussions)や[Issues](https://github.com/axonops/cqlai/issues)ページを通じて、CQLAIへのフィードバックや提案をお寄せください。

## 💬 コミュニティ & サポート

### 参加する
- 💡 **アイデアを共有**: [GitHubディスカッション](https://github.com/axonops/cqlai/discussions)で新機能を提案してください
- 🐛 **問題を報告**: バグを見つけましたか？ [Issueを開く](https://github.com/axonops/cqlai/issues/new/choose)
- 🤝 **貢献**: プルリクエストを歓迎します！ガイドラインについては[CONTRIBUTING.md](CONTRIBUTING.md)を参照してください
- ⭐ **スターをつける**: CQLAIが役に立つと思ったら、リポジトリにスターをつけてください！

### つながりを保つ
- 🌐 **ウェブサイト**: [axonops.com](https://axonops.com)
- 📧 **お問い合わせ**: サポートオプションについては当社のウェブサイトをご覧ください

## 📝 ライセンス

このプロジェクトはApache 2.0ライセンスの下でライセンスされています。詳細については[LICENSE](LICENSE)ファイルを参照してください。

サードパーティの依存関係ライセンスは、[THIRD-PARTY-LICENSES](THIRD-PARTY-LICENSES/)ディレクトリで入手できます。ライセンス帰属を再生成するには、`make licenses`を実行してください。

## ⚖️ 法的通知

*このプロジェクトには、プロジェクト、製品、またはサービスの商標またはロゴが含まれている場合があります。第三者の商標またはロゴの使用は、それらの第三者のポリシーに従います。*

- **AxonOps**はAxonOps Limitedの登録商標です。
- **Apache**、**Apache Cassandra**、**Cassandra**、**Apache Spark**、**Spark**、**Apache TinkerPop**、**TinkerPop**、**Apache Kafka**、**Kafka**は、Apache Software Foundationまたはその子会社のカナダ、米国および/またはその他の国における登録商標または商標です。
- **DataStax**は、DataStax, Inc.およびその子会社の米国および/またはその他の国における登録商標です。

---

<div align="center">
  <p><a href="https://axonops.com">AxonOps</a>チームが❤️を込めて作成</p>
</div>
