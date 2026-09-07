create table nexus_indexed.checkpoints
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
    PRIMARY KEY ((algorithm, id))
);

create index submission_tag ON nexus_indexed.checkpoints (tag);

create index host ON nexus_indexed.checkpoints (received_by_host);

create index lifecycle_stage ON nexus_indexed.checkpoints (lifecycle_stage);

INSERT INTO nexus_indexed.checkpoints (algorithm, id, lifecycle_stage, payload_uri, result_uri, algorithm_failure_cause,
                               algorithm_failure_details, received_by_host, received_at, sent_at, applied_configuration,
                               configuration_overrides, content_hash, last_modified, tag, api_version,
                               job_uid, parent)
VALUES ('test-algorithm', 'f47ac10b-58cc-4372-a567-0e02b2c3d479', 'NEW', 'http://localhost/payload',
        'http://localhost/result', NULL, NULL, 'host123', '2023-10-01T12:00:00.000Z',
        '2023-10-01T12:30:00.000Z', '{}', '{}', 'new_hash', '2023-10-01T12:45:00.000Z',
        'new_tag', 'v1.0', 'd94c16c8-2c1e-4f3a-85d1-2d9c3b7f0a24', '{}');

INSERT INTO nexus_indexed.checkpoints (algorithm, id, lifecycle_stage, payload_uri, result_uri, algorithm_failure_cause,
                               algorithm_failure_details, received_by_host, received_at, sent_at, applied_configuration,
                               configuration_overrides, content_hash, last_modified, tag, api_version,
                               job_uid, parent)
VALUES ('test-algorithm', '2c7b6e8d-cc3c-4b5b-a3f6-5d7b9e2c7f2a', 'BUFFERED', 'http://localhost/payload',
        'http://localhost/result', NULL, NULL, 'host123', '2023-10-02T10:00:00.000Z',
        '2023-10-02T10:30:00.000Z', '{}', '{}', 'buffered_hash', '2023-10-02T10:45:00.000Z',
        'buffered_tag', 'v2.0', '1f7b6e8d-cc3c-4b5b-a3f6-5d7b9e2c7f2a', '{}');

INSERT INTO nexus_indexed.checkpoints (algorithm, id, lifecycle_stage, payload_uri, result_uri, algorithm_failure_cause,
                               algorithm_failure_details, received_by_host, received_at, sent_at, applied_configuration,
                               configuration_overrides, content_hash, last_modified, tag, api_version,
                               job_uid, parent)
VALUES ('test-algorithm', '8a0c8aa9-fc9d-4b2e-9c5c-2c8d7f1e7a3f', 'RUNNING', 'http://localhost/payload',
        'http://localhost/result', NULL, NULL, 'host789', '2023-10-03T08:00:00.000Z',
        '2023-10-03T08:15:00.000Z', '{}', '{}', 'running_hash', '2023-10-03T08:30:00.000Z',
        'running_tag', 'v3.0', 'dfc7e9b5-8c3a-429b-9a33-d4a0a1f4b8d3', '{}');

INSERT INTO nexus_indexed.checkpoints (algorithm, id, lifecycle_stage, payload_uri, result_uri, algorithm_failure_cause,
                               algorithm_failure_details, received_by_host, received_at, sent_at, applied_configuration,
                               configuration_overrides, content_hash, last_modified, tag, api_version,
                               job_uid, parent)
VALUES ('test-algorithm', '550e8400-e29b-41d4-a716-446655440000', 'COMPLETED', 'http://localhost/payload',
        'http://localhost/result', NULL, NULL, 'host999', '2023-10-04T09:00:00.000Z',
        '2023-10-04T09:30:00.000Z', '{}', '{}', 'completed_hash', '2023-10-04T09:45:00.000Z',
        'completed_tag', 'v4.0', '395c9b92-c1f0-4e3b-9f43-7c5d4a3c7e9d', '{}');

INSERT INTO nexus_indexed.checkpoints (algorithm, id, lifecycle_stage, payload_uri, result_uri, algorithm_failure_cause,
                               algorithm_failure_details, received_by_host, received_at, sent_at, applied_configuration,
                               configuration_overrides, content_hash, last_modified, tag, api_version,
                               job_uid, parent)
VALUES ('test-algorithm', 'c3b7f1e7-9c5c-4b2e-8a0c-8aa9fc9d7f1e', 'FAILED', 'http://localhost/payload',
        'http://localhost/result', 'error', 'Unhandled exception occurred.', 'host321',
        '2023-10-05T10:00:00.000Z', '2023-10-05T10:30:00.000Z', '{}', '{}', 'failed_hash',
        '2023-10-05T10:45:00.000Z', 'failed_tag', 'v5.0', '3e7b1c8d-4d1c-4f5b-a3f7-9e2c7f1e8a9c', '{}');

INSERT INTO nexus_indexed.checkpoints (algorithm, id, lifecycle_stage, payload_uri, result_uri, algorithm_failure_cause,
                               algorithm_failure_details, received_by_host, received_at, sent_at, applied_configuration,
                               configuration_overrides, content_hash, last_modified, tag, api_version,
                               job_uid, parent)
VALUES ('test-algorithm', '7f1e8a9c-4d1c-4f5b-a3f7-9c5b1e7b3c8d', 'SCHEDULING_FAILED', 'http://localhost/payload',
        'http://localhost/result', 'scheduling_error', 'Could not schedule the job.', 'host654',
        '2023-10-06T11:00:00.000Z', '2023-10-06T11:30:00.000Z', '{}', '{}', 'scheduling_failed_hash',
        '2023-10-06T11:45:00.000Z', 'scheduling_failed_tag', 'v6.0', '1e7b3c8d-9f2c-4a0c-8b1c-7f1e9a3f4d5b', '{}');

INSERT INTO nexus_indexed.checkpoints (algorithm, id, lifecycle_stage, payload_uri, result_uri, algorithm_failure_cause,
                               algorithm_failure_details, received_by_host, received_at, sent_at, applied_configuration,
                               configuration_overrides, content_hash, last_modified, tag, api_version,
                               job_uid, parent)
VALUES ('test-algorithm', '9c5b1e7b-4d1c-4f5b-a3f7-1e7b3c8d8a9f', 'DEADLINE_EXCEEDED', 'http://localhost/payload',
        'http://localhost/result', 'deadline_exceeded', 'The job ran out of time.', 'host987',
        '2023-10-07T12:00:00.000Z', '2023-10-07T12:30:00.000Z', '{}', '{}', 'deadline_exceeded_hash',
        '2023-10-07T12:45:00.000Z', 'deadline_exceeded_tag', 'v7.0', '3c8d8a9f-4a0c-8b1c-1e7b4d5b9f2c', '{}');

INSERT INTO nexus_indexed.checkpoints (algorithm, id, lifecycle_stage, payload_uri, result_uri, algorithm_failure_cause,
                               algorithm_failure_details, received_by_host, received_at, sent_at, applied_configuration,
                               configuration_overrides, content_hash, last_modified, tag, api_version,
                               job_uid, parent)
VALUES ('test-algorithm', '4d5b9f2c-1e7b-3c8d-8a9f-4a0c8b1e7b3f', 'CANCELLED', 'http://localhost/payload',
        'http://localhost/result', 'cancelled_by_user', 'The job was cancelled by the user.', 'host111',
        '2023-10-08T13:00:00.000Z', '2023-10-08T13:30:00.000Z', '{}', '{}', 'cancelled_hash',
        '2023-10-08T13:45:00.000Z', 'cancelled_tag', 'v8.0', '8a9f4a0c-1e7b-3c8d-4d5b9f2c7f1e', '{}');

