# CQLSHRC File Support

CQLAI now supports reading configuration from standard CQLSHRC files, making it compatible with existing Cassandra/cqlsh configurations.

## Configuration Priority

Configuration is loaded in the following order (later sources override earlier ones):

1. **CQLSHRC file** - `~/.cassandra/cqlshrc` or `~/.cqlshrc`
2. **JSON config file** - `cqlai.json`, `~/.cqlai.json`, `~/.config/cqlai/config.json`, or `/etc/cqlai/config.json`
3. **Environment variables** - `CQLAI_*`, `CASSANDRA_*`, etc.

## Supported CQLSHRC Sections and Parameters

### [connection]
- `hostname` - Maps to `Host` (default: localhost)
- `port` - Maps to `Port` (default: 9042)
- `ssl` - Maps to `SSL.Enabled` (true/false)

### [authentication]
- `keyspace` - Maps to `Keyspace`
- `credentials` - Path to credentials file containing username/password

### [auth_provider]
- `module` - Maps to `AuthProvider.Module` (e.g., "cassandra.auth")
- `classname` - Maps to `AuthProvider.ClassName` (e.g., "PlainTextAuthProvider")
- `username` - Maps to `Username`

### [ssl]
- `certfile` - Maps to `SSL.CAPath` (CA certificate)
- `userkey` - Maps to `SSL.KeyPath` (client private key)
- `usercert` - Maps to `SSL.CertPath` (client certificate)
- `validate` - Maps to `SSL.InsecureSkipVerify` (inverted: validate=false means InsecureSkipVerify=true)

## Example CQLSHRC File

```ini
; Example CQLSHRC configuration
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

## Credentials File Format

If specified in the `[authentication]` section, the credentials file should contain:

```ini
[PlainTextAuthProvider]
username = your_username
password = your_password
```

## Path Expansion

- Paths starting with `~` are automatically expanded to the user's home directory
- This applies to all file paths in the configuration (credentials, SSL certificates, etc.)

## Notes

- Comments in CQLSHRC files start with `;` or `#`
- Not all CQLSHRC options are currently supported (e.g., UI settings, COPY options)
- JSON configuration files and environment variables take precedence over CQLSHRC settings
- For security, passwords should be stored in a separate credentials file, not in the main CQLSHRC