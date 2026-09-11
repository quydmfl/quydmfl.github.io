#!/bin/sh
# pg_hba.conf mặc định KHÔNG tự cho phép kết nối "replication" dù đã có dòng "all" cho database
# thường -- "replication" là 1 pseudo-database riêng, cần khai báo tường minh.
set -e
echo "host replication replicator all scram-sha-256" >> "$PGDATA/pg_hba.conf"
