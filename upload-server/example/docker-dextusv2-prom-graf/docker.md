# Docker Compose

`docker-compose up`
`docker-compose down`

## Docker initial notes, use docker-compose

## Grafana

$ mkdir grafana; cd grafana

$ docker volume create grafana

docker run -d --name=grafana -p 3000:3000 -v grafana:/var/lib/grafana grafana/grafana

## ----------------------------------------------------------

## Prometheus 

docker run -p 9090:9090 \
    --name=prometheus \
    -v /c/Users/Cos/code/grafana-prometheus/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml \
    -v prometheus-data:/prometheus \
    prom/prometheus

## ----------------------------------------------------------

## Service

docker build . -t dex-upload-server

docker run --name=dextusdserver -p 8080:8080 \
 -v ${pwd}/uploads:/app/uploads \ 
 dex-upload-server:latest

