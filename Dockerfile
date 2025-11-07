FROM alpine:latest
RUN   apk add --no-cache tzdata ca-certificates
WORKDIR /app
VOLUME [ "/app/pb_data/" ]
EXPOSE 9443

COPY auto-tls /app/auto-tls
# start PocketBase
ENTRYPOINT [ "/app/bilive-auth", "serve", "--http=0.0.0.0:9443" ]
CMD []
