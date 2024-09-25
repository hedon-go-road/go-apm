main:
	go run main.go

genpb:
	protoc --go_out=. ./protos/*

setup:
	docker compose  -f scripts/setup/docker-compose.yml up -d
