# CQLSHRCファイルサポート

CQLAIは標準のCQLSHRCファイルからの設定読み込みをサポートしており、既存のCassandra/cqlsh設定と互換性があります。

## 設定の優先順位

設定は次の順序でロードされます(後のソースが前のソースを上書きします):

1. **CQLSHRCファイル** - `~/.cassandra/cqlshrc`または`~/.cqlshrc`
2. **JSON設定ファイル** - `cqlai.json`、`~/.cqlai.json`、`~/.config/cqlai/config.json`、または`/etc/cqlai/config.json`
3. **環境変数** - `CQLAI_*`、`CASSANDRA_*`など

## サポートされるCQLSHRCセクションとパラメータ

### [connection]
- `hostname` - `Host`にマップ(デフォルト: localhost)
- `port` - `Port`にマップ(デフォルト: 9042)
- `ssl` - `SSL.Enabled`にマップ(true/false)

### [authentication]
- `keyspace` - `Keyspace`にマップ
- `credentials` - ユーザー名/パスワードを含む認証情報ファイルへのパス

### [auth_provider]
- `module` - `AuthProvider.Module`にマップ(例: "cassandra.auth")
- `classname` - `AuthProvider.ClassName`にマップ(例: "PlainTextAuthProvider")
- `username` - `Username`にマップ

### [ssl]
- `certfile` - `SSL.CAPath`にマップ(CA証明書)
- `userkey` - `SSL.KeyPath`にマップ(クライアント秘密鍵)
- `usercert` - `SSL.CertPath`にマップ(クライアント証明書)
- `validate` - `SSL.InsecureSkipVerify`にマップ(反転: validate=falseはInsecureSkipVerify=trueを意味します)

## CQLSHRCファイルの例

```ini
; CQLSHRCの設定例
[connection]
hostname = cassandra.example.com
port = 9042
ssl = true

[authentication]
keyspace = my_keyspace
credentials = ~/.cassandra/credentials

[auth_provider]
module = cassandra.auth
classname = PlainTextAuthProvider
username = cassandra_user

[ssl]
certfile = ~/certs/ca.pem
userkey = ~/certs/client-key.pem
usercert = ~/certs/client-cert.pem
validate = true
```

## 認証情報ファイル形式

`[authentication]`セクションで指定されている場合、認証情報ファイルには次が含まれている必要があります:

```ini
[PlainTextAuthProvider]
username = your_username
password = your_password
```

## パス展開

- `~`で始まるパスは自動的にユーザーのホームディレクトリに展開されます
- これは設定内のすべてのファイルパス(認証情報、SSL証明書など)に適用されます

## 注意事項

- CQLSHRCファイルのコメントは`;`または`#`で始まります
- すべてのCQLSHRCオプションが現在サポートされているわけではありません(例: UI設定、COPYオプション)
- JSON設定ファイルと環境変数はCQLSHRC設定よりも優先されます
- セキュリティのため、パスワードはメインのCQLSHRCではなく、別の認証情報ファイルに保存する必要があります
