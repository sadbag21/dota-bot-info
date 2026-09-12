FROM golang:1.27 AS build
WORKDIR /src
COPY ["go.mod", "go.sum", "./"]
RUN go mod download
COPY ["cmd/", "./cmd/"]
COPY ["internal/", "./internal/"]
COPY ["pkg/", "./pkg/"]
RUN go test ./internal/... ./pkg/... -count=1 && go vet ./...
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/dota-bot ./cmd/bot
RUN mkdir -p /out/data && chmod 700 /out/data

FROM scratch
COPY --from=build ["/etc/ssl/certs/ca-certificates.crt", "/etc/ssl/certs/ca-certificates.crt"]
COPY --from=build ["/out/dota-bot", "/dota-bot"]
COPY --from=build --chown=65532:65532 ["/out/data", "/data"]
USER 65532:65532
ENV LOG_LEVEL=INFO
ENV DATA_DIR=/data
STOPSIGNAL SIGTERM
ENTRYPOINT ["/dota-bot"]
