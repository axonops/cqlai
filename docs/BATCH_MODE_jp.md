# バッチモードドキュメント

CQLAIは`cqlsh`と互換性のあるバッチ実行モードをサポートしており、自動ページネーションと複数の出力形式でCQLコマンドを非対話的に実行できます。

## 概要

バッチモードは次の場合にアクティブになります:
- `-e`フラグを使用してコマンドを直接実行する
- `-f`フラグを使用してファイルからコマンドを実行する
- `cqlai`に入力をパイプする(stdinが端末でない)

バッチモードでは、CQLAIは手動介入なしにすべての結果ページを自動的に反復処理します。これは`cqlsh`と同様です。

## コマンドラインオプション

### CQLを直接実行(`-e`)

単一のCQLステートメントを実行して終了:

```bash
cqlai -e "SELECT * FROM system.local;"
```

### ファイルから実行(`-f`)

ファイルからCQLステートメントを実行:

```bash
cqlai -f script.cql
```

ファイルにはセミコロンで区切られた複数のステートメントを含めることができます。

### パイプ入力

stdinからCQLを実行:

```bash
echo "SELECT * FROM system.local;" | cqlai

# またはhere-documentから
cqlai <<EOF
USE my_keyspace;
SELECT * FROM my_table;
EOF
```

## 出力形式

`--format`フラグを使用して出力形式を指定:

### ASCIIテーブル(デフォルト)

```bash
cqlai -e "SELECT * FROM table;" --format ascii
```

`cqlsh`と同様の出力を生成:

```
 column1 | column2 | column3
---------+---------+---------
 value1  | value2  | value3
 value4  | value5  | value6

(2 rows)
```

### JSON

```bash
cqlai -e "SELECT * FROM table;" --format json
```

JSON配列出力を生成:

```json
[
  {
    "column1": "value1",
    "column2": "value2",
    "column3": "value3"
  },
  {
    "column1": "value4",
    "column2": "value5",
    "column3": "value6"
  }
]
```

### CSV

```bash
cqlai -e "SELECT * FROM table;" --format csv
```

CSV出力を生成:

```csv
column1,column2,column3
value1,value2,value3
value4,value5,value6
```

#### CSVオプション

- `--no-header`: ヘッダー行を省略
- `--field-separator`: カスタムフィールド区切り文字を使用(デフォルト: カンマ)

```bash
# ヘッダーなし
cqlai -e "SELECT * FROM table;" --format csv --no-header

# セミコロン区切り
cqlai -e "SELECT * FROM table;" --format csv --field-separator ";"

# タブ区切り値
cqlai -e "SELECT * FROM table;" --format csv --field-separator $'\t'
```

## 自動ページネーション

バッチモードでは、CQLAIは自動的にすべての結果ページを取得します:

```bash
# これは自動的にすべての行を取得します
cqlai -e "SELECT * FROM large_table;" > all_data.csv
```

ページネーションは透過的に行われます:
- 結果は取得されるとストリームで出力されます
- メモリ使用量はバッチ処理により最適化されます
- いつでもCtrl+Cで中断できます

## 接続オプション

すべての標準接続オプションはバッチモードで動作します:

```bash
cqlai --host cassandra.example.com \
      --port 9042 \
      --username myuser \
      --password mypass \
      --keyspace mykeyspace \
      -e "SELECT * FROM mytable;"
```

## 例

### CSVにエクスポート

```bash
# テーブルをCSVにエクスポート
cqlai -e "SELECT * FROM products;" --format csv > products.csv

# ヘッダーなしでエクスポート
cqlai -e "SELECT * FROM products;" --format csv --no-header > products_data.csv
```

### JSONにエクスポート

```bash
# jqで処理するためにエクスポート
cqlai -e "SELECT * FROM users;" --format json | jq '.[] | select(.age > 25)'
```

### 複数のステートメントを実行

スクリプトファイル`queries.cql`を作成:

```sql
USE my_keyspace;
SELECT COUNT(*) FROM users;
SELECT * FROM products WHERE price < 100;
```

実行:

```bash
cqlai -f queries.cql
```

### パイプライン処理

```bash
# 行数をカウント
echo "SELECT * FROM large_table;" | cqlai --format csv | wc -l

# 特定のカラムを抽出
echo "SELECT email FROM users;" | cqlai --format csv --no-header | sort | uniq
```

### 自動レポート

```bash
#!/bin/bash
# 日次レポートスクリプト

DATE=$(date +%Y-%m-%d)

# レポートを生成
cqlai -e "
  SELECT date, metric, value
  FROM metrics
  WHERE date = '$DATE'
" --format csv > "report_$DATE.csv"
```

## 対話モードとの違い

バッチモードでは:
- 対話型UIなし
- コマンド履歴なし
- 自動補完なし
- 自動ページネーション(手動ページコントロールなし)
- 結果はstdoutに直接ストリーム
- エラーはstderrへ
- 終了コードは成功(0)または失敗(非ゼロ)を示します

## cqlshとの互換性

CQLAIのバッチモードは`cqlsh`と互換性があるように設計されています:

```bash
# これらのコマンドは両方のツールで同様に動作します
cqlsh -e "SELECT * FROM system.local"
cqlai -e "SELECT * FROM system.local;"

# パイプ互換性
echo "SELECT * FROM table;" | cqlsh
echo "SELECT * FROM table;" | cqlai
```

主な違い:
- CQLAIは追加の出力形式(JSON)をサポート
- CQLAIはCQLステートメントにセミコロンが必要(対話型cqlshのように)
- CQLAIはPythonのcassandra-driverの代わりにGoのgocqlドライバーを使用
