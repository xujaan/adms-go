FROM scratch

COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY server /server
COPY templates /templates

EXPOSE 8000

ENV PORT=8000
ENV TZ=Asia/Jakarta

ENTRYPOINT ["/server"]
