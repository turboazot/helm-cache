FROM golang:1.17.11-alpine3.16 as builder
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o helm-cache .


# generate clean, final image for end users
FROM golang:1.17-stretch
COPY --from=builder /build/helm-cache .
RUN wget https://get.helm.sh/helm-v3.7.0-linux-amd64.tar.gz \
 && tar -xf helm-v3.7.0-linux-amd64.tar.gz \
 && mv ./linux-amd64/helm /usr/local/bin/helm \
 && chmod +x /usr/local/bin/helm

# executable
ENTRYPOINT [ "./helm-cache" ]