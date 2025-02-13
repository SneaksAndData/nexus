create table nexus.submission_buffer
(
    algorithm text,
    id        text,
    template  text,
    PRIMARY KEY ((algorithm, id))
);

alter table nexus.submission_buffer
    with default_time_to_live = 86400;
