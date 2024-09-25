main:
	go run main.go

genpb:
	protoc --go_out=. ./protos/*

setup:
	docker compose  -f zscripts/setup/docker-compose.yml up -d
	mysql -h 127.0.0.1 -P 23306 -u root -p < zscripts/setup/init.sql
