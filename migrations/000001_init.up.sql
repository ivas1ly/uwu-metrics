BEGIN TRANSACTION;

/*
                   Table "public.metrics"
 Column |       Type       | Collation | Nullable | Default
--------+------------------+-----------+----------+---------
 id     | text             |           | not null |
 mtype  | text             |           | not null |
 mdelta | bigint           |           |          |
 mvalue | double precision |           |          |
Indexes:
    "metrics_pkey" PRIMARY KEY, btree (id)
*/
CREATE TABLE IF NOT EXISTS metrics (
    id TEXT PRIMARY KEY,
    mtype TEXT NOT NULL,
    mdelta BIGINT NULL,
    mvalue DOUBLE PRECISION NULL
);

COMMIT;