create table nexus.checkpoints_by_host
(
    host                      text,
    lifecycle_stage           text,
    algorithm                 text,
    id                        text,
    PRIMARY KEY ((host, lifecycle_stage), id)
);

INSERT INTO nexus.checkpoints_by_host (algorithm, id, host, lifecycle_stage)
VALUES ('test-algorithm', 'f47ac10b-58cc-4372-a567-0e02b2c3d479','host123', 'NEW');

INSERT INTO nexus.checkpoints_by_host (algorithm, id, host, lifecycle_stage)
VALUES ('test-algorithm', '2c7b6e8d-cc3c-4b5b-a3f6-5d7b9e2c7f2a', 'host123', 'BUFFERED');

