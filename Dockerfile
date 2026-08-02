FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server/

FROM scratch
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /server /server
COPY templates /templates
COPY static /static

EXPOSE 8000
ENV PORT=8000
ENV TZ=Asia/Jakarta

ENTRYPOINT ["/server"]
