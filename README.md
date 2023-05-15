
# otel-go-metric

this docker-compose env helps understand basic golang otlp metric exporter.

one can visualize the exported metric in grafana.

follow the below steps to check it.

run `go mod vendor` in top dir.

`cd docker-compose`

execute `docker-compose up -d`

execute `docker ps -a` to see if all containers are in running state or not.

cd back to top of repo.

execute `go run main.go`

open grafana in browser : `localhost:3000`

in explore section, you would be able to see metric starting from `test_poc_*`

prometheus url : `localhost:9090`

otlp collector container's metrics : `localhost:8888/metrics`

metrics related to app : `localhost:8889/metrics`


Note that, the setup may not be secure.
