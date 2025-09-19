FROM --platform=$BUILDPLATFORM golang:1.24-bookworm AS build-stage

ARG TARGETOS 
ARG TARGETARCH
ARG APPVERSION
ARG BUILDNUMBER

WORKDIR /app

COPY . ./

RUN go mod verify && go mod vendor

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags "-X github.com/SneaksAndData/nexus-core/pkg/buildmeta.AppVersion=$APPVERSION -X github.com/SneaksAndData/nexus-core/pkg/buildmeta.BuildNumber=$BUILDNUMBER" -o /app -v ./...

FROM --platform=$BUILDPLATFORM public.ecr.aws/amazonlinux/amazonlinux:2 AS aws-cli-stage

ARG TARGETARCH

RUN yum update -y && yum install -y unzip curl \
    && cli_arch=$(test "$TARGETARCH" = "amd64" && echo "x86_64" || echo "aarch64") \
    && curl "https://awscli.amazonaws.com/awscli-exe-linux-$cli_arch.zip" -o "awscliv2.zip" \
    && unzip awscliv2.zip \
  # The --bin-dir is specified so that we can copy the
  # entire bin directory from the installer stage into
  # into /usr/local/bin of the final stage without
  # accidentally copying over any other executables that
  # may be present in /usr/local/bin of the installer stage.
    && ./aws/install --bin-dir /aws-cli-bin/


# Deploy the application binary into a lean image
FROM gcr.io/distroless/python3-debian12 AS build-release-stage

WORKDIR /app

COPY --from=build-stage /app /app

COPY --from=aws-cli-stage /usr/local/aws-cli/ /usr/local/aws-cli/
COPY --from=aws-cli-stage /aws-cli-bin/ /usr/local/bin/
 
COPY .container/appconfig.yaml /app/appconfig.yaml

USER nonroot:nonroot

ENTRYPOINT ["/app/nexus"]
