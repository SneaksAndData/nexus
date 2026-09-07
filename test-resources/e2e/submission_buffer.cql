create table nexus.submission_buffer
(
    algorithm text,
    id        text,
    cluster   text,
    template  text,
    PRIMARY KEY ((algorithm, id))
);
