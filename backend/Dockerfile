FROM golang:1.25-alpine AS builder
WORKDIR /app

# 先拉依赖（利用 layer 缓存）
COPY go.mod go.sum ./
RUN go mod download

# 拷贝源码并编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# 最终镜像（极小）
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/db/migrations ./db/migrations

EXPOSE 8080
CMD ["./server"]
