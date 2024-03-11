FROM golang:1.20-alpine AS BUILD

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on

WORKDIR /opt/source

COPY src/go.mod /opt/source/go.mod
COPY src/go.sum /opt/source/go.sum

RUN go mod download
COPY src/.. /opt/source/
RUN go build -a -o /opt/spark-ui-reverse-proxy main.go

FROM gcr.io/distroless/base as RUNTIME

COPY --from=BUILD /opt/spark-ui-reverse-proxy /usr/bin/
COPY --from=BUILD /opt/source/templates /templates

ENV KUBERNETES_SERVICE_HOST="kubernetes.default.svc" \
    KUBERNETES_SERVICE_PORT="443" \
    GIN_MODE=release

ENTRYPOINT ["/usr/bin/spark-ui-reverse-proxy"]
