create table nexus.checkpoints_by_tag
(
    tag                       text,
    algorithm                 text,
    id                        text,
    PRIMARY KEY ((tag), id)
);
