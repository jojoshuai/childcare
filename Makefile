.PHONY: build dev clean

# 完整构建：前端 + 嵌入后端二进制
build:
	cd web && npm install && npm run build
	cp -r web/dist/* backend/web/dist/
	cd backend && CGO_ENABLED=0 go build -tags embed -o ../childcare .

# 开发模式：分别启动前后端
dev:
	@echo "启动后端 (API)..."
	@cd backend && go run . &
	@echo "启动前端 (Vite dev server)..."
	@cd web && npm run dev &
	@echo "按 Ctrl+C 停止所有进程"
	@wait

# 清理构建产物
clean:
	rm -f childcare
	rm -rf backend/web/dist/*
