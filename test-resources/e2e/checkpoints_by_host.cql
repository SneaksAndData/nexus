create table nexus.checkpoints_by_host
(
    host                      text,
    lifecycle_stage           text,
    algorithm                 text,
    id                        text,
    PRIMARY KEY ((host, lifecycle_stage), id)
);
