FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-w -s \
    -X github.com/hsn0918/kubernetes-mcp/cmd/kubernetes-mcp/app.Version=${VERSION} \
    -X github.com/hsn0918/kubernetes-mcp/cmd/kubernetes-mcp/app.Commit=${COMMIT} \
    -X github.com/hsn0918/kubernetes-mcp/cmd/kubernetes-mcp/app.BuildDate=${BUILD_DATE}" \
    -o /app/kubernetes-mcp \
    ./cmd/kubernetes-mcp/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/kubernetes-mcp /app/kubernetes-mcp


EXPOSE 8080

ENTRYPOINT ["/app/kubernetes-mcp"]
# 可以通过 `docker run <image> version` 或 `docker run <image> server --transport=sse` 来覆盖
CMD ["server", "--transport=sse","-port=8080"]
