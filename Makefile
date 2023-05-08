build:
	docker build -t humorwang/getip:v1.0 .
tag:
	docker tag humorwang/getip:v1.0  humorwang/getip:latest
push:
	docker push humorwang/getip:latest