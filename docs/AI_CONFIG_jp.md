# CQLAIのAI設定

CQLAIは、AI駆動の自然言語からCQLクエリへの生成をサポートしています。環境変数を使用して優先するAIプロバイダーを設定できます。

## 使用方法

AI機能を使用するには、`.ai`に続けて自然言語のリクエストを入力します:

```
> .ai show all users with age greater than 25
```

これにより:
1. リクエストに基づいてCQLクエリプランを生成
2. 生成されたCQLをプレビュー表示
3. クエリを実行、編集、またはキャンセルすることができます

## 設定

AIプロバイダーは`cqlai.json`設定ファイルで設定されます。`cqlai.json.example`を`cqlai.json`にコピーし、AIセクションを更新してください:

### 基本設定
```json
{
  "host": "127.0.0.1",
  "port": 9042,
  ...
  "ai": {
    "provider": "mock",  // オプション: mock, openai, anthropic, gemini
    "apiKey": "",        // 一般的なAPIキー(プロバイダー固有のもので上書きされます)
    "model": ""          // 一般的なモデル(プロバイダー固有のもので上書きされます)
  }
}
```

### OpenAI設定
```json
"ai": {
  "provider": "openai",
  "openai": {
    "apiKey": "your-openai-api-key-here",
    "model": "gpt-4-turbo-preview"  // オプション、デフォルトはgpt-4-turbo-preview
  }
}
```

### Anthropic設定
```json
"ai": {
  "provider": "anthropic",
  "anthropic": {
    "apiKey": "your-anthropic-api-key-here",
    "model": "claude-3-sonnet-20240229"  // オプション
  }
}
```

### Google Gemini設定
```json
"ai": {
  "provider": "gemini",
  "gemini": {
    "apiKey": "your-gemini-api-key-here",
    "model": "gemini-pro"  // オプション、デフォルトはgemini-pro
  }
}
```

### Mockプロバイダー(デフォルト)
mockプロバイダーはテスト用にデフォルトで使用され、APIキーは必要ありません。リクエスト内のキーワードに基づいて簡単な例示クエリを生成します:

```json
"ai": {
  "provider": "mock"
}
```

## 機能

### クエリプラン生成
AIは次を含む構造化されたクエリプランを生成します:
- 操作タイプ(SELECT、INSERT、UPDATE、DELETE、CREATEなど)
- ターゲットキースペースとテーブル
- 選択または変更するカラム
- WHERE条件
- ORDER BY句
- LIMIT指定
- 信頼度レベル

### 安全機能
- **デフォルトで読み取り専用**: AIは明示的にデータ変更を求められない限り、SELECTクエリを優先します
- **危険な操作の警告**: 破壊的な操作(DROP、DELETE、TRUNCATE)は警告を表示します
- **確認が必要**: 有効になっている場合、危険な操作には追加の確認が必要です
- **スキーマ検証**: クエリは現在のCassandraスキーマに対して検証されます

### モーダルコントロール
AIがクエリを生成すると、次のことができます:
- **P**: CQLクエリとJSONクエリプラン間の表示を切り替え
- **Enter**: クエリを実行
- **Tab/矢印キー**: キャンセル、実行、編集ボタン間をナビゲート
- **編集**: 生成されたCQLを手動編集のために入力に配置
- **Esc**: 実行せずにキャンセル

## 実装状況

現在実装済み:
- ✅ テスト用のMockプロバイダー
- ✅ クエリプラン生成と検証
- ✅ プランからのCQLレンダリング
- ✅ プレビューと確認用のUIモーダル
- ✅ スキーマコンテキスト抽出

TODO:
- ⏳ 実際のOpenAI API統合
- ⏳ 実際のAnthropic API統合
- ⏳ 実際のGoogle Gemini API統合
- ⏳ クエリ最適化提案
- ⏳ 既存クエリの自然言語説明
