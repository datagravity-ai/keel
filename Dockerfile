FROM golang:1.24
COPY . /go/src/github.com/datagravity-ai/keel
WORKDIR /go/src/github.com/datagravity-ai/keel
RUN make install

FROM node:16.20.1-alpine
WORKDIR /app
COPY ui /app
RUN yarn
RUN yarn run lint --no-fix
RUN yarn run build

FROM alpine:latest
LABEL name="keel"
LABEL org.opencontainers.image.description Kubernetes Operator to automate Helm, DaemonSet, StatefulSet & Deployment updates


RUN apk --no-cache add ca-certificates

VOLUME /data
ENV XDG_DATA_HOME /data

COPY --from=0 /go/bin/keel /bin/keel
COPY --from=1 /app/dist /www
ENTRYPOINT ["/bin/keel"]
EXPOSE 9300
