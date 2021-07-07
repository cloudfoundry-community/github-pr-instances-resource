ARG base_image=scratch
ARG builder_image=golang:1.16

FROM ${builder_image} as builder
ADD . /go/src/github.com/aoldershaw/github-prs-resource
WORKDIR /go/src/github.com/aoldershaw/github-prs-resource
RUN curl -sL https://taskfile.dev/install.sh | sh
RUN ./bin/task build

FROM ${base_image} as resource
COPY --from=builder /bin/bash /bin/bash
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/aoldershaw/github-prs-resource/build /opt/resource

FROM resource
