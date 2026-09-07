create table nexus.checkpoints_by_tag
(
    tag                       text,
    algorithm                 text,
    id                        text,
    PRIMARY KEY ((tag), id)
);

INSERT INTO nexus.checkpoints_by_tag (algorithm, id, tag)
VALUES ('test-algorithm', 'f47ac10b-58cc-4372-a567-0e02b2c3d479', 'new_tag');

INSERT INTO nexus.checkpoints_by_tag (algorithm, id, tag)
VALUES ('test-algorithm', '2c7b6e8d-cc3c-4b5b-a3f6-5d7b9e2c7f2a', 'buffered_tag');

INSERT INTO nexus.checkpoints_by_tag (algorithm, id, tag)
VALUES ('test-algorithm', '8a0c8aa9-fc9d-4b2e-9c5c-2c8d7f1e7a3f','running_tag');

INSERT INTO nexus.checkpoints_by_tag (algorithm, id, tag)
VALUES ('test-algorithm', '550e8400-e29b-41d4-a716-446655440000','completed_tag');

INSERT INTO nexus.checkpoints_by_tag (algorithm, id, tag)
VALUES ('test-algorithm', 'c3b7f1e7-9c5c-4b2e-8a0c-8aa9fc9d7f1e', 'failed_tag');

INSERT INTO nexus.checkpoints_by_tag (algorithm, id, tag)
VALUES ('test-algorithm', '7f1e8a9c-4d1c-4f5b-a3f7-9c5b1e7b3c8d', 'scheduling_failed_tag');

INSERT INTO nexus.checkpoints_by_tag (algorithm, id, tag)
VALUES ('test-algorithm', '9c5b1e7b-4d1c-4f5b-a3f7-1e7b3c8d8a9f','deadline_exceeded_tag');

INSERT INTO nexus.checkpoints_by_tag (algorithm, id, tag)
VALUES ('test-algorithm', '4d5b9f2c-1e7b-3c8d-8a9f-4a0c8b1e7b3f','cancelled_tag');

