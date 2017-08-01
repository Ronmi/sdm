/*
Package mysql implements MySQL syntax generator.

YOU ARE SUGGESTED NOT TO CREATE TABLE WITH THIS DRIVER IN PRODUCTION. It might cause huge
performance lost if issuing incorrect size for column/index.

This driver accepts four DSN parameters:

  - charset=utf8: Default character set for table and string fields.
  - collate=utf8_general_ci: Default collation for table and string fields.
  - stringKeySize=256: Max length of indexed string fields, cannot exceed 256.
  - blobKeySize=2048: Max key length for indexed []byte fields.

  For example:
  mysql:stringKeySize=32


String field mapping

String fields are mapped to TEXT type if not indexed.
For example:

  type A struct {
          S string `sdm:"s"`
  }

The column "s" will be TEXT type.

For indexed string fields, VARCHAR is used:

  type A struct {
          S1 string `sdm:"s1,pk_key1"`
          S2 string `sdm:"s2,idx_key2"`
          S3 string `sdm:"s3,uniq_key3"`
          S4 string `sdm:"s4"`
  }

Column "s4" will be TEXT type, others are VARCHAR(256) type.

Converting date and time types

This driver supports auto-converting no matter what database/sql/driver your use.
*/
package mysql
