#!/usr/bin/env bash

cqlsh localhost -e "CREATE KEYSPACE nexus WITH replication = { 'class': 'SimpleStrategy', 'replication_factor': 1 } AND tablets = { 'enabled': false };"

echo 'Applying checkpoints table'

cqlsh localhost -f /opt/storage/checkpoints.cql

echo 'Applying submission_buffer table'

cqlsh localhost -f /opt/storage/submission_buffer.cql

echo 'Checking table'

cqlsh localhost -e 'SELECT * FROM nexus.submission_buffer'

echo 'Applying checkpoints_by_host table'

cqlsh localhost -f /opt/storage/checkpoints_by_host.cql

echo 'Checking table'

cqlsh localhost -e 'SELECT * FROM nexus.checkpoints_by_host'

echo 'Applying checkpoints_by_tag table'

cqlsh localhost -f /opt/storage/checkpoints_by_tag.cql

echo 'Checking table'

cqlsh localhost -e 'SELECT * FROM nexus.checkpoints_by_tag'

echo 'Applying payload_buffer table'

cqlsh localhost -f /opt/storage/payload_buffer.cql

echo 'Checking table'

cqlsh localhost -e 'SELECT * FROM nexus.payload_buffer'
