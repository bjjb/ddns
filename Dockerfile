# golang 1.17.3-alpine3.14 (amd64)
FROM golang:1.17-alpine AS builder
WORKDIR /go/ddns
RUN apk add --no-cache git ca-certificates build-base
RUN adduser -S ddns
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
ADD go.mod go.sum ./
RUN go mod download && go mod verify
ADD . .
RUN go build -a -installsuffix cgo -ldflags='-extldflags -static -w -s' -tags timetzdata -o /bin/ddns

FROM scratch
COPY --from=builder /bin/ddns /bin/ddns
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/
COPY --from=builder /etc/group /etc/
EXPOSE 80
USER ddns:nogroup
ENV ADDR=":http"
CMD ["/bin/ddns", "server"]
