FROM golang:1.25-alpine AS builder
WORKDIR /app

# 先拉依赖（利用 layer 缓存）
COPY go.mod go.sum ./
RUN go mod download

# 拷贝源码并编译（嵌入前端和迁移）
COPY . .
COPY web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -tags embed -o server .

# 最终镜像（极小）
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080
CMD ["./server"]
