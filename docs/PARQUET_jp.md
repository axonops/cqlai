# CQLAIにおけるParquetサポート

CQLAIは、Apache Parquet形式の包括的なサポートを提供し、CassandraとParquetファイル間の効率的なデータインポートとエクスポートを可能にします。この機能は、データ分析、機械学習ワークフロー、データアーカイブに特に役立ちます。

## 概要

Parquetは、優れた圧縮とエンコーディングスキームを提供するカラムナーストレージ形式であり、大規模データセットの保存と処理に最適です。CQLAIのParquet統合により、次のことが可能になります:

- CassandraテーブルデータをParquetファイルにエクスポート
- ParquetファイルをCassandraテーブルにインポート
- コレクション、UDT、ベクトルを含む複雑なCassandraデータ型を処理
- さまざまな圧縮アルゴリズムでストレージを最適化

## COPY TO Parquet

CassandraテーブルからParquet形式にデータをエクスポート。

### 基本的な使用方法

```sql
-- テーブル全体をParquetにエクスポート
COPY users TO '/path/to/users.parquet' WITH FORMAT='PARQUET';

-- 特定のカラムをエクスポート
COPY users (id, name, email) TO '/path/to/users.parquet' WITH FORMAT='PARQUET';

-- WHERE句を使用してエクスポート(Cassandraバージョンでサポートされている場合)
COPY users TO '/path/to/active_users.parquet' WITH FORMAT='PARQUET' WHERE status='active';
```

### パーティション分割データセット

より良い組織とクエリパフォーマンスのために、パーティション分割ディレクトリ構造にデータをエクスポート:

```sql
-- 単一のパーティションカラムでエクスポート
COPY events TO '/data/customers/' WITH FORMAT='PARQUET' AND PARTITION='customer_name';

-- 複数のパーティションカラムでエクスポート
COPY metrics TO '/data/metrics/' WITH FORMAT='PARQUET' AND PARTITION='year,month,day';

-- TimeUUID仮想カラムでエクスポート(時間コンポーネントを抽出)
COPY events TO '/data/events/' WITH FORMAT='PARQUET' AND PARTITION='event_id.year,event_id.month';
-- event_idがTimeUUIDカラムの場合

-- 結果のディレクトリ構造:
-- /data/metrics/
-- ├── year=2024/
-- │   ├── month=01/
-- │   │   ├── day=01/
-- │   │   │   └── part-00000.parquet
-- │   │   └── day=02/
-- │   │       └── part-00000.parquet
-- │   └── month=02/
-- │       └── day=01/
-- │           └── part-00000.parquet
```

#### TimeUUID仮想カラム抽出

TimeUUIDカラムから時間コンポーネントを抽出してインテリジェントなパーティショニングを行います:

```sql
-- Cassandraでの適切な時系列テーブル構造
CREATE TABLE events (
    event_name text,           -- パーティションキー(例: 'temperature', 'cpu_usage')
    event_time timeuuid,        -- 時間ベースの順序付け用クラスタリングカラム
    event_value double,
    metadata map<text, text>,
    PRIMARY KEY (event_name, event_time)
) WITH CLUSTERING ORDER BY (event_time DESC);

-- TimeUUIDから抽出された時間コンポーネントでパーティション分割
COPY events TO '/data/events/' WITH FORMAT='PARQUET'
AND PARTITION='event_time.year,event_time.month,event_time.day,event_time.hour';
-- 階層構造を作成:
-- /data/events/event_time.year=2024/event_time.month=01/event_time.day=15/event_time.hour=14/part-00000.parquet

-- TimeUUIDで利用可能な仮想カラム:
-- .year   - 年を抽出(例: 2024)
-- .month  - 月を抽出(1-12)
-- .day    - 日を抽出(1-31)
-- .hour   - 時間を抽出(0-23)
-- .date   - YYYY-MM-DD文字列として日付を抽出

-- センサーデータの例
CREATE TABLE sensor_data (
    sensor_id text,
    reading_time timeuuid,
    temperature double,
    humidity double,
    PRIMARY KEY (sensor_id, reading_time)
) WITH CLUSTERING ORDER BY (reading_time DESC);

-- 分析用に時間ごとのパーティションでエクスポート
COPY sensor_data TO '/data/sensor-data/' WITH FORMAT='PARQUET'
AND PARTITION='reading_time.date,reading_time.hour';
-- 作成: /data/sensor-data/reading_time.date=2024-01-15/reading_time.hour=14/part-00000.parquet
```

パーティショニングの利点:
- **パーティションプルーニング**: パーティションカラムでフィルタリングする際に関連するパーティションのみを読み取り
- **並列処理**: 異なるパーティションを同時に処理可能
- **増分更新**: 既存データを書き直さずに新しいパーティションを追加
- **ストレージの最適化**: 古いパーティションを個別にアーカイブ

### 圧縮オプション

Parquetは複数の圧縮アルゴリズムをサポート:

```sql
-- Snappy圧縮を使用(デフォルト、最良のバランス)
COPY users TO 'users.parquet' WITH FORMAT='PARQUET' AND COMPRESSION='SNAPPY';

-- GZIP圧縮を使用(より良い圧縮率)
COPY users TO 'users.parquet' WITH FORMAT='PARQUET' AND COMPRESSION='GZIP';

-- ZSTD圧縮を使用(最良の圧縮率)
COPY users TO 'users.parquet' WITH FORMAT='PARQUET' AND COMPRESSION='ZSTD';

-- LZ4圧縮を使用(最速)
COPY users TO 'users.parquet' WITH FORMAT='PARQUET' AND COMPRESSION='LZ4';

-- 圧縮なし
COPY users TO 'users.parquet' WITH FORMAT='PARQUET' AND COMPRESSION='NONE';
```

### パフォーマンス最適化

大規模データセットでより良いパフォーマンスを得るためにチャンクサイズを制御:

```sql
-- チャンクサイズを50,000行に設定
COPY large_table TO 'data.parquet' WITH FORMAT='PARQUET' AND CHUNKSIZE=50000;

-- 短縮表記を使用
COPY large_table TO 'data.parquet' WITH FORMAT='PARQUET' AND CHUNKSIZE='50K';
COPY huge_table TO 'data.parquet' WITH FORMAT='PARQUET' AND CHUNKSIZE='1M';
```

## COPY FROM Parquet

ParquetファイルからCassandraテーブルにデータをインポート。

### 基本的な使用方法

```sql
-- Parquetファイル全体をインポート
COPY users FROM '/path/to/users.parquet' WITH FORMAT='PARQUET';

-- 特定のカラムをインポート
COPY users (id, name, email) FROM '/path/to/users.parquet' WITH FORMAT='PARQUET';

-- カラムマッピングでインポート
COPY users (user_id, full_name) FROM 'data.parquet' WITH FORMAT='PARQUET';
```

### パーティション分割データセットのインポート

パーティション分割Parquetデータセットからデータをインポート:

```sql
-- パーティション分割ディレクトリからインポート
COPY events FROM '/data/events/' WITH FORMAT='PARQUET';

-- パーティションフィルターでインポート(スキャンされるデータを削減)
COPY events FROM '/data/events/' WITH FORMAT='PARQUET' AND PARTITION_FILTER='year=2024,month=01';

-- パターンで特定のパーティションをインポート
COPY metrics FROM '/data/metrics/year=2024/' WITH FORMAT='PARQUET';
```

パーティション分割データセットをインポートする場合:
- パーティションカラムはディレクトリ構造から自動的に検出されます
- パーティション値はインポートされたデータのカラムとして含まれます
- Hiveスタイルのパーティショニング規約(key=value)をサポート
- 特殊文字とNULL値を適切に処理

### インポートオプション

```sql
-- ヘッダー行をスキップ
COPY users FROM 'users.parquet' WITH FORMAT='PARQUET' AND SKIPROWS=1;

-- インポートする行数を制限
COPY users FROM 'users.parquet' WITH FORMAT='PARQUET' AND MAXROWS=10000;

-- オプションを組み合わせ
COPY users FROM 'users.parquet' WITH FORMAT='PARQUET' AND SKIPROWS=1 AND MAXROWS=5000;

-- パーティション分割データセットのバッチサイズを設定
COPY events FROM '/data/events/' WITH FORMAT='PARQUET' AND CHUNKSIZE=5000;
```

## データ型サポート

CQLAIのParquet統合は、すべての主要なCassandraデータ型をサポート:

### 基本型

| Cassandra型 | Parquet型 | 注意事項 |
|---------------|--------------|-------|
| text/varchar | STRING (UTF8) | 完全なUnicodeサポート |
| int | INT32 | 32ビット符号付き整数 |
| bigint | INT64 | 64ビット符号付き整数 |
| float | FLOAT | 32ビット浮動小数点 |
| double | DOUBLE | 64ビット浮動小数点 |
| boolean | BOOLEAN | True/false値 |
| timestamp | TIMESTAMP_MILLIS | ミリ秒精度 |
| date | DATE | エポックからの日数 |
| time | TIME_MILLIS | 深夜からのミリ秒 |
| uuid/timeuuid | STRING | フォーマットされた文字列として保存 |
| blob | BYTE_ARRAY | バイナリデータ |
| decimal | DECIMAL | 任意精度 |

### コレクション型

```sql
-- リスト
CREATE TABLE products (
    id int PRIMARY KEY,
    tags list<text>,
    prices list<decimal>
);

-- セット
CREATE TABLE users (
    id int PRIMARY KEY,
    emails set<text>,
    roles set<text>
);

-- マップ
CREATE TABLE settings (
    user_id int PRIMARY KEY,
    preferences map<text, text>,
    scores map<text, int>
);
```

### ユーザー定義型(UDT)

```sql
-- UDTを定義
CREATE TYPE address (
    street text,
    city text,
    zip_code text,
    country text
);

-- テーブルで使用
CREATE TABLE customers (
    id int PRIMARY KEY,
    name text,
    home_address address,
    work_address address
);

-- エクスポート/インポートはUDT構造を保持
COPY customers TO 'customers.parquet' WITH FORMAT='PARQUET';
COPY customers FROM 'customers.parquet' WITH FORMAT='PARQUET';
```

### ベクトル型(Cassandra 5.0+)

機械学習と類似性検索のユースケースをサポート:

```sql
-- ベクトルカラムを持つテーブルを作成
CREATE TABLE embeddings (
    id int PRIMARY KEY,
    content text,
    vector vector<float, 1536>,  -- 1536次元のベクトル埋め込み
    metadata text
);

-- ベクトルをParquetにエクスポート
COPY embeddings TO 'embeddings.parquet' WITH FORMAT='PARQUET';

-- ベクトルはParquetで固定サイズLIST型として保存されます
-- Apache ArrowとPandasと互換性があります
```

## クラウドストレージのサポート

クラウドストレージ統合の場合、rcloneなどのツールを使用してクラウドストレージをローカルファイルシステムとしてマウントします。詳細については、[クラウドストレージドキュメント](cloud-storage.md)を参照してください。

## 高度な機能

### 大規模データセットのストリーミング

非常に大きなテーブルの場合、CQLAIはメモリ使用量を最小限に抑えるためにストリーミングを使用します:

```sql
-- 最適化されたストリーミングで大きなテーブルをエクスポート
COPY large_events_table TO 'events.parquet'
WITH FORMAT='PARQUET'
AND COMPRESSION='ZSTD'
AND CHUNKSIZE='100K';
```

### キャプチャモードの統合

CQLAIのCAPTUREコマンドは、Parquetファイルへのクエリ結果を保存する対話的な方法を提供します。これは、COPYコマンドとは根本的に異なります:

#### パーティション分割キャプチャ

キャプチャされたクエリ結果をパーティション分割データセットに保存してより良い組織化を実現:

```sql
-- 単一のパーティションカラムでパーティション分割キャプチャを開始
CAPTURE PARQUET '/data/analysis/' WITH PARTITION='date';

-- 後続のクエリは日付の値でパーティション分割されます
SELECT * FROM events WHERE date >= '2024-01-01';
-- 作成: /data/analysis/date=2024-01-01/part-00000.parquet
--      /data/analysis/date=2024-01-02/part-00000.parquet
--      など

-- 複数カラムのパーティショニング
CAPTURE PARQUET '/data/metrics/' WITH PARTITION='year,month,day';

SELECT * FROM metrics WHERE year = 2024;
-- 作成: /data/metrics/year=2024/month=01/day=01/part-00000.parquet
--      /data/metrics/year=2024/month=01/day=02/part-00000.parquet

CAPTURE OFF;
```

##### TimeUUIDからの仮想カラム抽出

パーティション分割キャプチャの強力な機能は、パーティショニングのためにTimeUUIDカラムから時間コンポーネントを抽出できることです:

```sql
-- 適切な時系列テーブル構造
CREATE TABLE events (
    event_name text,           -- パーティションキー
    event_time timeuuid,        -- クラスタリングカラム
    event_value double,
    metadata map<text, text>,
    PRIMARY KEY (event_name, event_time)
) WITH CLUSTERING ORDER BY (event_time DESC);

-- TimeUUIDから抽出された時間コンポーネントでパーティション分割
CAPTURE PARQUET '/data/events/' WITH PARTITION='event_time.year,event_time.month,event_time.day';

SELECT * FROM events WHERE event_name = 'temperature';
-- 作成: /data/events/event_time.year=2024/event_time.month=01/event_time.day=15/part-00000.parquet
--      /data/events/event_time.year=2024/event_time.month=01/event_time.day=16/part-00000.parquet

-- TimeUUIDで抽出可能な仮想カラム:
-- .year   - TimeUUIDから年を抽出
-- .month  - TimeUUIDから月を抽出
-- .day    - TimeUUIDから日を抽出
-- .hour   - TimeUUIDから時間を抽出
-- .date   - YYYY-MM-DD文字列として日付を抽出

CAPTURE OFF;
```

仮想カラムはディレクトリパーティショニングにのみ使用され、Parquetファイル自体には保存されません。DuckDBやApache Sparkなどのツールでクエリする場合、これらのパーティション値はHiveスタイルのディレクトリ構造に基づいてカラムとして自動的に利用可能になります。

##### 圧縮とパフォーマンスオプション

最適なパフォーマンスのために圧縮とファイルサイズを制御:

```sql
-- より良い圧縮率のためにZSTD圧縮を使用
CAPTURE PARQUET '/data/compressed/' WITH COMPRESSION='ZSTD' AND PARTITION='date';

-- 最速の圧縮のためにLZ4を使用
CAPTURE PARQUET '/data/fast/' WITH COMPRESSION='LZ4';

-- 最大ファイルサイズを制御(パーティション分割データセットに便利)
CAPTURE PARQUET '/data/sized/' WITH MAX_FILE_SIZE='500MB' AND PARTITION='date';
-- パーティションファイルが500MBを超えると、新しいファイル(part-00001.parquet)が作成されます

-- すべてのオプションを組み合わせ
CAPTURE PARQUET '/data/optimized/' WITH
    PARTITION='event_id.year,event_id.month'
    AND COMPRESSION='ZSTD'
    AND MAX_FILE_SIZE='1GB';

CAPTURE OFF;
```

パーティション分割キャプチャの利点:
- 時間やカテゴリで大規模データセットを整理
- 効率的なデータライフサイクル管理を可能に
- 増分処理パイプラインをサポート
- ダウンストリーム分析クエリを最適化
- TimeUUIDからの自動仮想カラム抽出
- Hiveスタイルのパーティション分割データセットと互換性

#### COPYとの主な違い

| 側面 | COPY | CAPTURE |
|--------|------|---------|
| **目的** | テーブル全体の一括エクスポート/インポート | アドホッククエリの結果を保存 |
| **範囲** | 単一テーブル操作 | 任意のテーブルにまたがる複数のクエリ |
| **使用ケース** | データ移行、バックアップ、ETL | 探索的分析、レポート |
| **実行** | 即時、単一操作 | セッションベース、継続的 |
| **データソース** | オプションのフィルター付きテーブルデータ | 任意のSELECTクエリ結果 |

#### キャプチャの仕組み

```sql
-- キャプチャを開始 - 後続のクエリ結果が保存されます
CAPTURE PARQUET '/tmp/analysis_results.parquet';

-- クエリを実行 - 結果がParquetファイルに書き込まれます
SELECT * FROM users WHERE country='US';
-- これにより、次のカラムを持つParquetファイルが作成されます: id, name, email, countryなど

-- 重要: 後続のクエリは同じスキーマを持つ必要があります
SELECT * FROM users WHERE country='UK';  -- ✓ 動作 - 同じカラム
SELECT * FROM users WHERE age > 18;      -- ✓ 動作 - 同じカラム

-- これは失敗するか問題を引き起こします - 異なるカラム!
-- SELECT id, order_total FROM orders;   -- ✗ 異なるスキーマ

-- キャプチャを停止
CAPTURE OFF;
```

**スキーマの制限**: Parquetにキャプチャする場合、キャプチャセッション内のすべてのクエリは同じ順序で同じカラムを返す必要があります。これは、Parquetファイルがファイルの途中で変更できない固定スキーマを持つためです。

#### ページング動作

大きな結果セットをキャプチャする場合、CQLAIは自動的にページングを処理します:

```sql
-- Parquet形式でキャプチャを開始
CAPTURE PARQUET '/tmp/large_results.parquet';

-- このクエリは数百万行を返す可能性があります
SELECT * FROM events WHERE date >= '2024-01-01';
-- CQLAIは自動的に結果をページングします:
-- - データをチャンク単位で取得(デフォルトページあたり5000行)
-- - 各ページをParquetファイルに書き込み
-- - 進捗を表示: "Page 1 of 1000..."
-- - すべてのデータがキャプチャされるまで継続
-- - メモリ効率的 - メモリには一度に1ページのみ

CAPTURE OFF;
```

#### キャプチャ構文の例

```sql
-- Parquetへの基本的なキャプチャ
CAPTURE PARQUET '/tmp/results.parquet';

-- 圧縮付きでキャプチャ
CAPTURE PARQUET '/tmp/compressed.parquet' WITH COMPRESSION='ZSTD';

-- パーティショニング付きでキャプチャ
CAPTURE PARQUET '/tmp/partitioned/' WITH PARTITION='date';

-- すべてのオプション付きでキャプチャ
CAPTURE PARQUET '/data/output/' WITH
    PARTITION='year,month'
    AND COMPRESSION='LZ4'
    AND MAX_FILE_SIZE='100MB';

-- 現在のキャプチャステータスを確認
CAPTURE;

-- キャプチャを停止
CAPTURE OFF;
```

#### Parquetでのキャプチャのユースケース

1. **同じスキーマ結果のフィルタリングと結合**
   ```sql
   -- 同じテーブルからのフィルタリングされた結果をキャプチャ
   CAPTURE '/tmp/filtered_users.parquet' FORMAT='PARQUET';
   SELECT * FROM users WHERE country='US' AND status='active';
   SELECT * FROM users WHERE country='UK' AND status='active';
   SELECT * FROM users WHERE country='CA' AND status='active';
   CAPTURE OFF;
   -- すべてのクエリが同じスキーマを持つため、正しく追加されます
   ```

2. **時系列データ収集**
   ```sql
   -- 同じスキーマで時間ごとのスナップショットをキャプチャ
   CAPTURE '/tmp/metrics_snapshot.parquet' FORMAT='PARQUET';
   SELECT hour, metric_name, value FROM metrics WHERE hour='2024-01-01 00:00:00';
   SELECT hour, metric_name, value FROM metrics WHERE hour='2024-01-01 01:00:00';
   SELECT hour, metric_name, value FROM metrics WHERE hour='2024-01-01 02:00:00';
   CAPTURE OFF;
   ```

3. **大きなテーブルのページ分割エクスポート**
   ```sql
   -- 大きなテーブルを管理可能なチャンクでエクスポート
   CAPTURE '/tmp/large_export.parquet' FORMAT='PARQUET';
   SELECT * FROM events WHERE date='2024-01-01' LIMIT 100000;
   SELECT * FROM events WHERE date='2024-01-02' LIMIT 100000;
   SELECT * FROM events WHERE date='2024-01-03' LIMIT 100000;
   CAPTURE OFF;
   ```

**注意**: 異なるスキーマを持つ異なるクエリから結果をキャプチャする場合は、代わりにJSONまたはCSV形式の使用を検討してください:
```sql
-- JSON形式は異なるスキーマを処理できます
CAPTURE '/tmp/mixed_results.json' FORMAT='JSON';
SELECT COUNT(*) as user_count FROM users;
SELECT id, name, email FROM users LIMIT 10;
SELECT order_id, total FROM orders LIMIT 10;
CAPTURE OFF;
```

#### 重要な注意事項

- **スキーマの一貫性**: Parquetキャプチャセッション内のすべてのクエリは同一のスキーマを持つ必要があります
- **追加動作**: 各クエリ結果は同じParquetファイルに行を追加します(同じスキーマが必要)
- **メモリ効率**: 大きな結果は自動的にページングされ、メモリ使用量は一定に保たれます
- **進捗表示**: 大きな結果セットの現在のページ番号を表示
- **形式の代替**: 異なるスキーマを持つクエリをキャプチャするにはJSONまたはCSV形式を使用

### ファイル検出

CQLAIはファイル拡張子からParquet形式を自動的に検出:

```sql
-- 自動形式検出
COPY users TO 'users.parquet';  -- 自動的にPARQUET形式を使用
COPY users FROM 'data.parquet'; -- PARQUETとして自動検出
```

## ユースケース

### 1. データ分析パイプライン

Apache Spark、pandas、またはその他の分析ツールで分析するためにCassandraデータをエクスポート:

```sql
-- Spark処理用にエクスポート
COPY events TO '/data/events.parquet' WITH FORMAT='PARQUET' AND COMPRESSION='SNAPPY';

-- Python pandasで直接読み取り可能:
-- df = pd.read_parquet('events.parquet')
```

### 2. データアーカイブ

優れた圧縮で過去データをアーカイブ:

```sql
-- 最大圧縮で古いデータをアーカイブ
COPY historical_data TO '/archive/data_2023.parquet'
WITH FORMAT='PARQUET'
AND COMPRESSION='ZSTD'
WHERE year=2023;
```

### 3. 機械学習ワークフロー

MLトレーニング用にベクトルと特徴をエクスポート:

```sql
-- 埋め込みと特徴をエクスポート
COPY ml_features TO 'training_data.parquet'
WITH FORMAT='PARQUET'
AND COMPRESSION='SNAPPY';

-- トレーニング用にPythonでロード:
-- features = pd.read_parquet('training_data.parquet')
-- X = np.stack(features['vector'].values)
```

### 4. データ移行

Cassandraクラスタ間でデータを移行:

```sql
-- ソースクラスタ: エクスポート
COPY users TO 'users_backup.parquet' WITH FORMAT='PARQUET';

-- ターゲットクラスタ: インポート
COPY users FROM 'users_backup.parquet' WITH FORMAT='PARQUET';
```

## パフォーマンスに関する考慮事項

### チャンクサイズ

- デフォルト: チャンクあたり10,000行
- 大きな行の場合: チャンクサイズを減らす(例: 1000-5000)
- 小さな行の場合: チャンクサイズを増やす(例: 50000-100000)

### 圧縮のトレードオフ

| 圧縮 | 速度 | 圧縮率 | 使用ケース |
|------------|-------|-------|----------|
| NONE | 最速 | なし | 一時ファイル、高速I/O |
| SNAPPY | 高速 | 良好 | デフォルト、バランスの取れたパフォーマンス |
| LZ4 | 非常に高速 | 良好 | リアルタイム処理 |
| GZIP | 遅い | より良い | ネットワーク転送 |
| ZSTD | より遅い | 最良 | 長期保存 |

### メモリ使用量

- ストリーミングモードはメモリフットプリントを最小化
- チャンクサイズはメモリ使用量に影響: `memory ≈ chunk_size × avg_row_size`
- 非常に幅広いテーブルの場合、チャンクサイズを減らす

## 制限事項

1. **ネストされたコレクション**: 深くネストされたコレクション(例: `list<map<text, set<int>>>`)はサポートが限定される可能性があります
2. **カスタム型**: カスタムCassandra型は文字列に変換される可能性があります
3. **ストリーミングのみ**: COPY FROMはストリーミングを使用 - ファイル全体がメモリにロードされません
4. **スキーママッチング**: ParquetとCassandra間でカラム名と型が互換性がある必要があります

## トラブルシューティング

### 一般的な問題

**問題**: "型XをParquetに変換できません"
```sql
-- 解決策: データ型の互換性を確認
DESCRIBE TABLE your_table;
-- すべての型がサポートされていることを確認
```

**問題**: メモリ不足エラー
```sql
-- 解決策: チャンクサイズを減らす
COPY large_table TO 'data.parquet' WITH FORMAT='PARQUET' AND CHUNKSIZE=1000;
```

**問題**: エクスポートのパフォーマンスが遅い
```sql
-- 解決策: チャンクサイズを増やし、より高速な圧縮を使用
COPY table TO 'data.parquet' WITH FORMAT='PARQUET'
AND COMPRESSION='LZ4'
AND CHUNKSIZE='50K';
```

## 例

### 完全なエクスポート/インポートワークフロー

```sql
-- 1. ソーステーブルを作成
CREATE KEYSPACE IF NOT EXISTS analytics
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE analytics;

CREATE TABLE IF NOT EXISTS user_events (
    user_id uuid,
    event_time timestamp,
    event_type text,
    properties map<text, text>,
    vector list<float>,
    PRIMARY KEY (user_id, event_time)
) WITH CLUSTERING ORDER BY (event_time DESC);

-- 2. サンプルデータを挿入
INSERT INTO user_events (user_id, event_time, event_type, properties, vector)
VALUES (uuid(), toTimestamp(now()), 'click',
        {'page': 'home', 'button': 'signup'},
        [0.1, 0.2, 0.3, 0.4, 0.5]);

-- 3. 圧縮付きでParquetにエクスポート
COPY user_events TO '/tmp/events.parquet'
WITH FORMAT='PARQUET'
AND COMPRESSION='ZSTD';

-- 4. 宛先テーブルを作成
CREATE TABLE IF NOT EXISTS user_events_archive (
    user_id uuid,
    event_time timestamp,
    event_type text,
    properties map<text, text>,
    vector list<float>,
    PRIMARY KEY (user_id, event_time)
);

-- 5. Parquetからインポート
COPY user_events_archive FROM '/tmp/events.parquet'
WITH FORMAT='PARQUET';

-- 6. インポートを確認
SELECT COUNT(*) FROM user_events_archive;
```

### 完全なパーティション分割キャプチャワークフロー

```sql
-- 1. TimeUUIDを持つテーブルを作成
CREATE TABLE IF NOT EXISTS events (
    event_id timeuuid PRIMARY KEY,
    event_type text,
    user_id int,
    data text
);

-- 2. サンプルデータを挿入
INSERT INTO events (event_id, event_type, user_id, data)
VALUES (now(), 'login', 123, 'user logged in');
INSERT INTO events (event_id, event_type, user_id, data)
VALUES (now(), 'purchase', 123, 'bought item ABC');
INSERT INTO events (event_id, event_type, user_id, data)
VALUES (now(), 'logout', 123, 'user logged out');

-- 3. TimeUUIDから年と月でパーティション分割キャプチャを開始
CAPTURE PARQUET '/data/events/' WITH
    PARTITION='event_id.year,event_id.month'
    AND COMPRESSION='ZSTD';

-- 4. クエリを実行 - 結果は自動的にパーティション分割されます
SELECT * FROM events WHERE event_type = 'login';
SELECT * FROM events WHERE event_type = 'purchase';
SELECT * FROM events WHERE user_id = 123;

-- 5. キャプチャを停止
CAPTURE OFF;

-- 6. DuckDBでパーティション分割データをクエリ
-- ファイルは次のように整理されます:
-- /data/events/event_id.year=2024/event_id.month=01/part-00000.parquet
-- /data/events/event_id.year=2024/event_id.month=02/part-00000.parquet

-- 7. DuckDBでパーティションプルーニングを使用してクエリ
-- 2024年1月のファイルのみを読み取ります
-- duckdb -c "SELECT * FROM '/data/events/**/*.parquet' WHERE \"event_id.year\" = 2024 AND \"event_id.month\" = 1;"
```

## ベストプラクティス

1. **常にFORMAT='PARQUET'を指定** - .parquet拡張子を使用する場合でも明確にする
2. **本番環境のエクスポートには圧縮を使用** - SNAPPYまたはZSTD推奨
3. **特定のデータパターンでチャンクサイズをテスト**
4. **大規模エクスポート中のメモリ使用量を監視**
5. **インポート後にCOUNT(*)とサンプルクエリでデータを検証**
6. **ユースケースに基づいて適切な圧縮を使用**(ストレージ vs. 速度)
7. **大規模エクスポートのパーティショニングを検討** - 時間やその他の次元で

## 計画中の機能

以下の機能は将来のリリースで計画されています:

### 近い将来(v0.1.x)

1. **スキーマ進化サポート**
   - カラムの追加/削除の自動スキーママッピング
   - 型変換の警告とオプション
   - インポート前のスキーマ検証

2. **並列処理**
   - 大きなテーブルのマルチスレッドエクスポート
   - 同時チャンク処理
   - パーティションの並列ファイル書き込み

3. **マウントされたクラウドストレージ**
   ```sql
   -- マウントされたクラウドストレージにエクスポート(rclone、s3fsなど経由)
   COPY users TO '/mnt/cloud/users.parquet'
   WITH FORMAT='PARQUET';
   ```

4. **統計とメタデータ**
   - クエリ最適化のための行グループ統計
   - カラム統計(最小、最大、null数)
   - 効率的なフィルタリングのためのBloomフィルター
   - Parquetファイルのカスタムメタデータ

### 中期(v0.2.x)

1. **高度なデータ型**
   - 完全なネストされたコレクションサポート(任意の深さ)
   - 空間データの幾何学型
   - JSONカラム型マッピング
   - カスタム型ハンドラー

2. **増分エクスポート/インポート**
   ```sql
   -- 前回のエクスポート以降の変更のみをエクスポート
   COPY users TO 'users_delta.parquet'
   WITH FORMAT='PARQUET'
   AND SINCE='2024-01-01 00:00:00';

   -- マージインポート(アップサート)
   COPY users FROM 'users_update.parquet'
   WITH FORMAT='PARQUET'
   AND MODE='UPSERT';
   ```

3. **データ変換**
   ```sql
   -- エクスポート中に変換
   COPY users TO 'users.parquet'
   WITH FORMAT='PARQUET'
   AND TRANSFORM='{"email": "LOWER", "created_at": "DATE_ONLY"}';
   ```

4. **圧縮プロファイル**
   ```sql
   -- 事前定義された最適化プロファイル
   COPY large_table TO 'data.parquet'
   WITH FORMAT='PARQUET'
   AND PROFILE='ANALYTICS';  -- Spark/Presto用に最適化

   COPY ml_data TO 'features.parquet'
   WITH FORMAT='PARQUET'
   AND PROFILE='ML';  -- Python/Arrow用に最適化
   ```

5. **進捗監視**
   - リアルタイム進捗バー
   - ETA計算
   - 操作中の詳細な統計
   - 再開可能な操作

### 長期(v0.3.x+)

1. **Apache Arrow統合**
   - ゼロコピーデータ転送
   - Arrow Flightプロトコルサポート
   - ダイレクトメモリ形式互換性
   - Python/pandasとの相互運用性の向上

2. **Delta Lake形式**
   ```sql
   -- Deltaテーブルとしてエクスポート
   COPY users TO '/delta/users'
   WITH FORMAT='DELTA';

   -- タイムトラベルクエリ
   COPY users FROM '/delta/users'
   WITH FORMAT='DELTA'
   AND VERSION='2024-01-01';
   ```

3. **ストリーミングCDC(変更データキャプチャ)**
   ```sql
   -- 変更の継続的なエクスポート
   CAPTURE STREAM changes TO 'kafka://topic'
   FROM users
   WITH FORMAT='PARQUET'
   AND MODE='CDC';
   ```

4. **クエリプッシュダウン**
   - Parquet述語プッシュダウン
   - カラムプルーニング最適化
   - 行グループフィルタリング
   - スマートデータスキップ

5. **データ品質機能**
   - データ検証ルール
   - 自動データクレンジング
   - 重複検出
   - データプロファイリングレポート

6. **MLフレームワークとの統合**
   ```sql
   -- ML形式への直接エクスポート
   COPY features TO 'model_data.tfrecord'
   WITH FORMAT='TENSORFLOW';

   COPY embeddings TO 'vectors.lance'
   WITH FORMAT='LANCE';  -- ベクトル検索用に最適化
   ```

7. **分散操作**
   - コーディネーター-ワーカーアーキテクチャ
   - ノード間での分散エクスポート
   - インポートのロードバランシング
   - フォールトトレランスと再試行ロジック

8. **高度なセキュリティ**
   - Parquetのカラムレベル暗号化
   - エクスポート中のフィールドレベルマスキング
   - すべての操作の監査ログ
   - エクスポートのロールベースアクセス制御

## 機能リクエスト

機能リクエストと貢献を歓迎します！アイデアを次の方法で送信してください:
- GitHub Issues: [github.com/axonops/cqlai/issues](https://github.com/axonops/cqlai/issues)
- Discussions: [github.com/axonops/cqlai/discussions](https://github.com/axonops/cqlai/discussions)

次の機能が優先されます:
1. 大規模操作のパフォーマンスを改善
2. データ分析エコシステムとの互換性を強化
3. 一般的なエンタープライズユースケースをサポート
4. 後方互換性を維持

## 関連ドキュメント

- [COPYコマンドリファレンス](./COPY.md)
- [データ型ガイド](./DATA_TYPES.md)
- [パフォーマンスチューニング](./PERFORMANCE.md)
- [Apache Parquet形式](https://parquet.apache.org/docs/)
- [Apache Arrow](https://arrow.apache.org/)
- [Delta Lake](https://delta.io/)
