# 標準化されたCassandraデータ型処理

## 概要

cqlaiには、`internal/db/types.go`の`CQLTypeHandler`を通じてすべてのCassandra/CQLデータ型を処理するための包括的で標準化されたメカニズムが含まれています。

## アーキテクチャ

### 型ハンドラー(`CQLTypeHandler`)
`cqlai/internal/db/types.go`に配置

```go
type CQLTypeHandler struct {
    TimeFormat      string // 時刻表示形式(デフォルトRFC3339)
    HexPrefix       string // 16進値のプレフィックス(デフォルト"0x")
    NullString      string // null値を表示する文字列(デフォルト"null")
    CollectionLimit int    // コレクションに表示する最大アイテム数(0 = 無制限)
    TruncateStrings int    // 文字列の最大長(0 = 切り詰めなし)
}
```

### 2層型解決

1. **型情報ベース**: gocqlからCQL型情報が利用可能な場合
2. **実行時型検出**: Goの型システムを使用したフォールバック

## サポートされるCQLデータ型

### ネイティブ型(完全サポート ✅)

| CQL型 | Go型 | 表示形式 | 例 |
|----------|---------|----------------|---------|
| **ascii** | string | プレーンテキスト | `"hello"` |
| **bigint** | int64 | 10進数 | `9223372036854775807` |
| **blob** | []byte | 16進数 | `0x48656c6c6f` |
| **boolean** | bool | true/false | `true` |
| **counter** | int64 | 10進数 | `42` |
| **date** | time.Time | YYYY-MM-DD | `2024-01-15` |
| **decimal** | *inf.Dec/string | 10進数文字列 | `123.456` |
| **double** | float64 | 科学表記 | `3.14159` |
| **duration** | gocql.Duration | 人間が読める形式 | `2mo5d100ns` |
| **float** | float32 | 科学表記 | `3.14` |
| **inet** | net.IP | IPアドレス | `192.168.1.1` |
| **int** | int32 | 10進数 | `2147483647` |
| **smallint** | int16 | 10進数 | `32767` |
| **text** | string | プレーンテキスト | `"Hello World"` |
| **time** | time.Duration | 期間文字列 | `13h45m30s` |
| **timestamp** | time.Time | RFC3339 | `2024-01-15T14:30:00Z` |
| **timeuuid** | gocql.UUID | UUID文字列 | `550e8400-e29b-41d4-a716-446655440000` |
| **tinyint** | int8 | 10進数 | `127` |
| **uuid** | gocql.UUID | UUID文字列 | `123e4567-e89b-12d3-a456-426614174000` |
| **varchar** | string | プレーンテキスト | `"text value"` |
| **varint** | *big.Int | 10進数文字列 | `123456789012345678901234567890` |

### コレクション型(完全サポート ✅)

| CQL型 | Go型 | 表示形式 | 例 |
|----------|---------|----------------|---------|
| **list<T>** | []T | 角括弧 | `[1, 2, 3]` |
| **set<T>** | []T | 角括弧 | `[a, b, c]` |
| **map<K,V>** | map[K]V | 波括弧 | `{key1: val1, key2: val2}` |

### 複雑な型(完全サポート ✅)

| CQL型 | Go型 | 表示形式 | 例 |
|----------|---------|----------------|---------|
| **tuple<T1,T2,...>** | []interface{} | 丸括弧 | `(1, "text", true)` |
| **UDT** | map[string]interface{} | マップのように | `{field1: val1, field2: val2}` |
| **frozen<T>** | Tと同じ | Tと同じ | 透過的 |

### 特殊型(サポート ✅)

| CQL型 | Go型 | 表示形式 | 注意事項 |
|----------|---------|----------------|-------|
| **vector<float, n>** | []float32 | 角括弧 | `[0.1, 0.2, 0.3]` - SAIベクトル検索 |
| **custom** | interface{} | 文字列表現 | ユーザー定義型 |

## 型固有の機能

### NULL処理
- すべての型はnull値を適切に処理
- 設定可能なnull表示文字列(デフォルト: `"null"`)
- 時刻型のゼロ値はnullとして表示

### コレクション機能
- **切り詰め**: 表示されるコレクションアイテムのオプション制限
- **ネストされたコレクション**: ネストされたマップ/リストが適切にフォーマット
- **型の保持**: 要素の型情報を維持

### バイナリデータ
- Blob型は設定可能なプレフィックス付きの16進数で表示
- 空のblobはプレフィックスのみを表示(例: `0x`)

### 時刻型
- 設定可能な時刻形式(デフォルトRFC3339)
- Date型は日付部分のみを表示
- Duration型は人間が読める形式を使用
- ゼロ時刻はnullとして表示

### 数値型
- 適切な場合は浮動小数点数の科学表記
- decimalとvarint型の完全な精度
- カウンターカラムの適切な処理

## cqlaiでの使用方法

### SELECTクエリ
```go
// visitor_stubs.goで
typeHandler := db.NewCQLTypeHandler()
for _, rowMap := range rows {
    for i, col := range columns {
        if val, ok := rowMap[col.Name]; ok {
            // 最良のフォーマットのために利用可能な場合は型情報を使用
            row[i] = typeHandler.FormatValue(val, col.TypeInfo)
        }
    }
}
```

### 設定オプション
```go
handler := db.NewCQLTypeHandler()
handler.TimeFormat = "2006-01-02 15:04:05"  // カスタム時刻形式
handler.HexPrefix = ""                       // 16進数のプレフィックスなし
handler.NullString = "NULL"                  // SQLスタイルのnull
handler.CollectionLimit = 10                 // 最大10アイテムを表示
handler.TruncateStrings = 100               // 長い文字列を切り詰め
```

## 利点

1. **一貫性**: アプリケーション全体で統一的にフォーマットされたすべての型
2. **完全性**: ベクトルとUDTを含むすべての30以上のCQL型を処理
3. **柔軟性**: 設定可能なフォーマットオプション
4. **パフォーマンス**: 効率的な型検出とフォーマット
5. **保守性**: 集約型処理ロジック
6. **拡張性**: 新しい型や形式の追加が容易

## エラー処理

- 不明な型でパニックしない
- 認識できない型の場合は`fmt.Sprintf("%v", val)`にフォールバック
- nilポインターを適切に処理
- 型アサーション失敗を安全に管理

## テストカバレッジ

型ハンドラーがカバーする内容:
- すべてのネイティブCQL型
- さまざまな要素型を持つすべてのコレクション型
- ネストされたコレクション(リストのマップ、マップのリストなど)
- ユーザー定義型
- ML/AIワークロード用のベクトル型
- NullおよびZero値処理
- エッジケース(空のコレクション、ゼロ時刻など)

## 将来の拡張

潜在的な改善:
1. JSON出力形式オプション
2. CSV互換フォーマット
3. アプリケーション固有の型のカスタム型レジストリ
4. ロケール固有の数値フォーマット
5. バイナリデータエンコーディングオプション(base64など)
6. 深くネストされた構造のコレクション深度制限
