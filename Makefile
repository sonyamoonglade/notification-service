build-local:
	docker build -f ./docker/local.Dockerfile -t sonyamoonglade/sancho-hub:notification-local . && docker push sonyamoonglade/sancho-hub:notification-local

build-prod:
	docker build -f ./docker/prod.Dockerfile -t sonyamoonglade/sancho-hub:notification-prod . && docker push sonyamoonglade/sancho-hub:notification-prod

run:
	docker-compose -f ./docker/docker-compose.dev.yaml --env-file ./.env.local up --build

stop:
	docker-compose -f ./docker/docker-compose.dev.yaml down