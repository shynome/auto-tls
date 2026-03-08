FROM alpine:3.23.3@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659
RUN   apk add --no-cache tzdata ca-certificates
WORKDIR /app
VOLUME [ "/app/pb_data/" ]
EXPOSE 9443

COPY auto-tls /app/auto-tls
# start PocketBase
ENTRYPOINT [ "/app/auto-tls", "serve", "--http=0.0.0.0:9443" ]
CMD []
