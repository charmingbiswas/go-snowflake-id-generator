all:
	go run ./cmd/main.go

build:
	go build -o snowflake ./cmd/main.go

image:
	docker build -t snowflake:v1 .

deploy:
	kubectl apply -f infra/k8s/snowflake-service.yml -f infra/k8s/snowflake-statefulset.yml

destroy:
	kubectl delete -f infra/k8s/snowflake-service.yml -f infra/k8s/snowflake-statefulset.yml

install:
	helm install snowflake-helm-chart ./infra/snowflake-helm -f ./infra/snowflake-helm/values.yaml

uninstall:
	helm uninstall snowflake-helm-chart

.PHONY: all build image deploy destroy install uninstall