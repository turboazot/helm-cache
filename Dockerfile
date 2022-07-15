FROM golang:1.17.11-alpine3.16 as builder
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o helm-cache .

FROM alpine:3.16.0
COPY --from=builder /build/helm-cache /usr/local/bin/helm-cache

ENTRYPOINT [ "helm-cache" ]