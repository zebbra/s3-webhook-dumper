FROM cgr.dev/chainguard/static:latest
LABEL org.opencontainers.image.source = "https://github.com/zebbra/s3-webhook-dumper"

COPY /s3-webhook-dumper /s3-webhook-dumper

CMD ["/s3-webhook-dumper"]
