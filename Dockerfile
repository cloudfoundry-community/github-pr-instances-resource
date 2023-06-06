ARG base_image=alpine
ARG builder_image=golang

FROM ${builder_image} as builder
ADD . /go/src/github.com/cloudfoundry-community/github-pr-instances-resource
WORKDIR /go/src/github.com/cloudfoundry-community/github-pr-instances-resource
RUN curl -sL https://taskfile.dev/install.sh | sh
RUN ./bin/task build

FROM ${base_image} as resource

RUN apk add --update --no-cache \
    git \
    git-lfs \
    openssh

COPY scripts/askpass.sh /usr/local/bin/askpass.sh
ADD scripts/install_git_crypt.sh install_git_crypt.sh
RUN ./install_git_crypt.sh && rm ./install_git_crypt.sh

COPY --from=builder /go/src/github.com/cloudfoundry-community/github-pr-instances-resource/build /opt/resource

FROM resource
