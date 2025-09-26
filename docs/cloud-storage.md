# Cloud Storage Support

## Important Change

As of version 2.0, CQLAI no longer includes built-in cloud storage SDKs for S3, Azure Blob, or Google Cloud Storage. This change reduces the binary size by approximately 6MB and follows the Unix philosophy of "do one thing well."

## Using Cloud Storage with CQLAI

You can still work with cloud storage by mounting it as a local filesystem using tools like **rclone**.

### Quick Start with rclone

1. **Install rclone**:
```bash
# macOS
brew install rclone

# Linux
curl https://rclone.org/install.sh | sudo bash

# Windows
winget install Rclone.Rclone
```

2. **Configure your cloud storage**:
```bash
rclone config
# Follow the interactive setup for your provider
```

3. **Mount cloud storage as a local filesystem**:
```bash
# Mount S3 bucket
rclone mount s3:my-bucket /mnt/s3 --daemon --vfs-cache-mode writes

# Mount Azure container
rclone mount azure:container /mnt/azure --daemon --vfs-cache-mode writes

# Mount Google Cloud Storage
rclone mount gcs:bucket /mnt/gcs --daemon --vfs-cache-mode writes
```

4. **Use with CQLAI**:
```sql
-- Export to mounted cloud storage
COPY users TO '/mnt/s3/data/users.parquet';

-- Import from mounted cloud storage
COPY users FROM '/mnt/s3/data/users.csv';

-- Capture to cloud storage
CAPTURE '/mnt/azure/query-results.json' FORMAT='JSON';
```

## Migration from Previous Versions

If you were using direct cloud URLs in previous versions:

### Before (Direct URLs - No Longer Supported):
```sql
COPY users TO 's3://my-bucket/users.parquet';
COPY users FROM 'gs://bucket/data.csv';
COPY users TO 'az://container/export.parquet';
```

### After (Via Mounted Filesystem):
```sql
COPY users TO '/mnt/s3/users.parquet';
COPY users FROM '/mnt/gcs/data.csv';
COPY users TO '/mnt/azure/export.parquet';
```

## Alternative Mounting Tools

Besides rclone, you can use any tool that mounts cloud storage as a filesystem:

### AWS S3
- **s3fs-fuse**: Native S3 filesystem
  ```bash
  s3fs mybucket /mnt/s3 -o passwd_file=~/.passwd-s3fs
  ```
- **goofys**: High-performance S3 mounting
  ```bash
  goofys my-bucket /mnt/s3
  ```

### Azure Blob Storage
- **blobfuse**: Microsoft's official FUSE driver
  ```bash
  blobfuse /mnt/azure --tmp-path=/tmp/blobfuse \
    --config-file=/path/to/config.cfg
  ```

### Google Cloud Storage
- **gcsfuse**: Google's official FUSE adapter
  ```bash
  gcsfuse my-bucket /mnt/gcs
  ```

## Benefits of This Approach

1. **Flexibility**: Use any of 70+ storage backends supported by rclone (Dropbox, OneDrive, Box, etc.)
2. **Separation of Concerns**: CQLAI focuses on CQL operations, rclone handles storage
3. **Better Caching**: rclone provides sophisticated caching options
4. **Unified Interface**: All storage looks like local filesystem to CQLAI
5. **Smaller Binary**: 6MB reduction in CQLAI binary size
6. **No Credential Management**: rclone handles all authentication

## Performance Tips

1. **Use VFS Caching**: The `--vfs-cache-mode writes` flag improves write performance
2. **Adjust Buffer Size**: For large files, increase buffer size
   ```bash
   rclone mount s3:bucket /mnt/s3 --daemon \
     --vfs-cache-mode full \
     --buffer-size 256M
   ```
3. **Parallel Transfers**: For multiple files, use parallel transfers
   ```bash
   rclone mount s3:bucket /mnt/s3 --daemon \
     --transfers 16
   ```

## Troubleshooting

### Error: "cloud storage URLs are no longer supported"
This means you're trying to use old-style cloud URLs. Mount your storage first using rclone or similar tools.

### Permission Denied
Ensure the mount point has proper permissions:
```bash
sudo mkdir -p /mnt/s3
sudo chown $USER:$USER /mnt/s3
rclone mount s3:bucket /mnt/s3 --daemon
```

### Mount Not Visible
Check if the mount is active:
```bash
mount | grep rclone
ps aux | grep rclone
```

### Slow Performance
Try adjusting cache settings:
```bash
rclone mount s3:bucket /mnt/s3 --daemon \
  --vfs-cache-mode full \
  --vfs-cache-max-size 10G \
  --vfs-cache-max-age 1h
```

## Examples

### Backup Cassandra Table to S3
```bash
# Setup
rclone mount s3:backups /mnt/backups --daemon

# In CQLAI
COPY users TO '/mnt/backups/users-$(date +%Y%m%d).parquet';
```

### Import Dataset from Azure
```bash
# Setup
rclone mount azure:datasets /mnt/datasets --daemon

# In CQLAI
COPY events FROM '/mnt/datasets/events.csv' WITH HEADER=true;
```

### Export to Multiple Cloud Providers
```bash
# Mount multiple providers
rclone mount s3:backup /mnt/s3 --daemon
rclone mount azure:backup /mnt/azure --daemon

# In CQLAI - export to both
COPY users TO '/mnt/s3/users.parquet';
COPY users TO '/mnt/azure/users.parquet';
```

## Further Reading

- [rclone Documentation](https://rclone.org/docs/)
- [s3fs-fuse GitHub](https://github.com/s3fs-fuse/s3fs-fuse)
- [gcsfuse Documentation](https://cloud.google.com/storage/docs/gcsfuse)
- [blobfuse Documentation](https://github.com/Azure/azure-storage-fuse)