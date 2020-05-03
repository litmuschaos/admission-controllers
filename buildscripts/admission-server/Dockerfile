FROM ubuntu:16.04

RUN apt-get update && apt-get install -y \
    iproute2

ADD admission-server /usr/local/bin/admission-server


ARG BUILD_DATE
LABEL org.label-schema.name="admission-server"
LABEL org.label-schema.description="webhook admission server policy for Litmus"
LABEL org.label-schema.url="http://www.litmuschaos.io/"
LABEL org.label-schema.vcs-url="https://github.com/litmuschaos/litmus-admission-webhook"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$BUILD_DATE

ENTRYPOINT ["/usr/local/bin/admission-server"]
