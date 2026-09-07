#!/usr/bin/env bash

cqlsh localhost -e "CREATE KEYSPACE nexus WITH replication = { 'class': 'SimpleStrategy', 'replication_factor': 1 } AND tablets = { 'enabled': false };"
cqlsh localhost -e "CREATE KEYSPACE nexus_indexed WITH replication = { 'class': 'SimpleStrategy', 'replication_factor': 1 } AND tablets = { 'enabled': false };"

echo 'Applying checkpoints table'

cqlsh localhost -f /opt/storage/checkpoints_indexed.cql

echo 'Checking table'

cqlsh localhost -e 'SELECT * FROM nexus_indexed.checkpoints'

echo 'Applying submission_buffer table'

cqlsh localhost -f /opt/storage/submission_buffer.cql
cqlsh localhost -f /opt/storage/submission_buffer_indexed.cql

echo 'Checking table'

cqlsh localhost -e 'SELECT * FROM nexus.submission_buffer'
cqlsh localhost -e 'SELECT * FROM nexus_indexed.submission_buffer'

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
cqlsh localhost -f /opt/storage/payload_buffer_indexed.cql

echo 'Checking table'

cqlsh localhost -e 'SELECT * FROM nexus.payload_buffer'
cqlsh localhost -e 'SELECT * FROM nexus_indexed.payload_buffer'
