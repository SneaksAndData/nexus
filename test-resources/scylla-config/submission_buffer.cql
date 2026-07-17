create table nexus.submission_buffer
(
    algorithm text,
    id        text,
    cluster   text,
    template  text,
    PRIMARY KEY ((algorithm, id))
);

INSERT INTO nexus.submission_buffer (algorithm, id, cluster, template)
VALUES ('test-algorithm',
        '123e4567-e89b-12d3-a456-426614174000',
        'test-cluster',
        '{
            "apiVersion": "batch/v1",
            "kind": "Job",
            "metadata": {
                "name": "unit-test-job",
                "labels": {
                    "app": "unit-test"
                }
            },
            "spec": {
                "template": {
                    "metadata": {
                        "labels": {
                            "app": "unit-test"
                        }
                    },
                    "spec": {
                        "containers": [
                            {
                                "name": "unit-test-container",
                                "image": "busybox",
                                "command": ["sh", "-c", "echo Running unit test..."],
                       "resources": {
                           "requests": {
                               "cpu": "100m",
                               "memory": "64Mi"
                           },
                           "limits": {
                               "cpu": "200m",
                               "memory": "128Mi"
                           }
                       }
                   }
               ],
               "restartPolicy": "Never"
           }
       },
       "backoffLimit": 1
   }
}');

