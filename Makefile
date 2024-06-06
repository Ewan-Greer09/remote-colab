make:
	docker compose up --build -d
	docker ps

stop:
	docker compose down