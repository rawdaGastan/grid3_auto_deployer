PWD := $(shell pwd)

swagger: 
	swagger validate swagger.yml
	docker run -i yousan/swagger-yaml-to-html < swagger.yml > docs/swagger.html

run: 
	docker compose up
