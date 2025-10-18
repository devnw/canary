FROM scratch
ARG TARGETPLATFORM
ENTRYPOINT ["/usr/bin/canary"]
COPY $TARGETPLATFORM/canary /usr/bin/
