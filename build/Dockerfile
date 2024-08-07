# Multi-stage docker build
# Build stage
FROM golang:alpine AS builder

LABEL maintainer="LitmusChaos"

ARG TARGETPLATFORM

ADD .. /admission-controller
WORKDIR /admission-controller

RUN export GOOS=$(echo ${TARGETPLATFORM} | cut -d / -f1) && \
    export GOARCH=$(echo ${TARGETPLATFORM} | cut -d / -f2)

RUN go env

RUN CGO_ENABLED=0 go build -buildvcs=false -o /output/admission-controller -v ./main.go

# Packaging stage
FROM registry.access.redhat.com/ubi9/ubi-minimal:9.4

LABEL maintainer="LitmusChaos"

ENV AdmissionController=/usr/local/bin/admission-controller

COPY --from=builder /output/admission-controller ${AdmissionController}
RUN chown 65534:0 ${AdmissionController} && chmod 755 ${AdmissionController}

USER 65534

ENTRYPOINT ["/usr/local/bin/admission-controller"]