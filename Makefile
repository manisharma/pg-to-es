up:
	touch .env
	echo "PG_HOST=postgres" >> .env
	echo "PG_PORT=4532" >> .env
	echo "PG_USERNAME=user" >> .env
	echo "PG_PASSWORD=secret" >> .env
	echo "PG_DB_NAME=db" >> .env
	echo "PG_LISTENER_CHANNEL=core_db_event" >> .env
	echo "ES_HOST=http://elasticsearch:9200" >> .env
	echo "ES_INDEX=root" >> .env
	echo "SERVER_PORT=8080" >> .env
	docker-compose up --build --remove-orphans

down:
	docker-compose down

fmt:
	go fmt ./...

test:
	go vet ./...
	go test -v ./... -count=1