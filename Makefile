run:
	go run cmd/main.go

build-local:
	docker build -f ./docker/local.Dockerfile -t sonyamoonglade/sancho-hub:notification-local . && docker push sonyamoonglade/sancho-hub:notification-local

build-prod:
	docker build -f ./docker/prod.Dockerfile -t sonyamoonglade/sancho-hub:notification-prod . && docker push sonyamoonglade/sancho-hub:notification-prod

cp-env:
	cp .env.prod ../deployment/notification/