create table nexus_indexed.payload_buffer
(
    algorithm                 text,
    id                        text,
    payload_content           text,
    PRIMARY KEY ((algorithm, id))
);

