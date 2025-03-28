FROM golang:1.24
COPY . /go/src/github.com/datagravity-ai/keel
WORKDIR /go/src/github.com/datagravity-ai/keel
RUN make install

FROM alpine:latest
LABEL name="keel"
LABEL org.opencontainers.image.description Kubernetes Operator to automate Helm, DaemonSet, StatefulSet & Deployment updates

RUN apk --no-cache add ca-certificates

VOLUME /data
ENV XDG_DATA_HOME /data

COPY --from=0 /go/bin/keel /bin/keel
ENTRYPOINT ["/bin/keel"]
EXPOSE 9300
