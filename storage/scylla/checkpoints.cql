create table nexus.checkpoints
(
    algorithm                 text,
    id                        text,
    lifecycle_stage           text,
    payload_uri               text,
    result_uri                text,
    algorithm_failure_cause   text,
    algorithm_failure_details text,
    received_by_host          text,
    received_at               timestamp,
    sent_at                   timestamp,
    applied_configuration     text,
    configuration_overrides   text,
    content_hash              text,
    last_modified             timestamp,
    tag                       text,
    api_version               text,
    job_uid                   text,
    parent                    text,
    payload_valid_for         text,
    PRIMARY KEY ((algorithm, id))
);

create index submission_tag ON nexus.checkpoints (tag);

create index host ON nexus.checkpoints (received_by_host);

create index lifecycle_stage ON nexus.checkpoints (lifecycle_stage);
