# YDB-specific configuration for stroppy tests

To support the authentication methods specific to YDB Managed Service, stroppy supports the additional environment variables when running in *client* mode only:
* `YDB_SERVICE_ACCOUNT_KEY_FILE_CREDENTIALS` - the path to the service account key file. When configured, the key file is used to authenticate the connection.
* `YDB_METADATA_CREDENTIALS` - when set to `1`, the service account key associated with the Cloud compute instance is used to authenticate the connection.
* `YDB_ACCESS_TOKEN_CREDENTIALS` - YDB access token. When configured, the access token is passed as is to authenticate the connection.

In addition, the following YDB-specific environment variables are supported:
* `YDB_STROPPY_HASH_TRANSFER_ID` - when set to `1`, Base64-encoded SHA-1 hash is written for `transfer_id` field in the `transfer` table;
* `YDB_STROPPY_PARTITIONS_COUNT` - [`AUTO_PARTITIONING_MIN_PARTITIONS_COUNT`](https://ydb.tech/en/docs/concepts/datamodel/table#auto_partitioning_partition_size_mb) setting value for `account` and `transfer` tables;
* `YDB_STROPPY_PARTITIONS_SIZE` - [`AUTO_PARTITIONING_PARTITION_SIZE_MB`](https://ydb.tech/en/docs/concepts/datamodel/table#auto_partitioning_min_partitions_count) setting value for `account` and `transfer` tables.
