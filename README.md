# grpc-broker

A simple gRPC based pub/sub message broker

Example of consumer:
https://github.com/mahadeva604/grpc-broker/tree/main/cmd/consumer

Example of producer:
https://github.com/mahadeva604/grpc-broker/tree/main/cmd/producer

API:
https://github.com/mahadeva604/grpc-broker/tree/main/pkg/broker

gRPC API:
https://github.com/mahadeva604/grpc-broker/tree/main/pkg/api/broker

<br>

Example of running grpc-broker server:
```
docker run -p 8086:8086 --name=grpc-broker --restart=always -d mahadeva604/grpc-broker:0.0.2
```

Example of running grpc-broker consumer:
```
docker run --net=host mahadeva604/grpc-broker-consumer:0.0.3 -s localhost:8086 -t topic1
```

Example of running grpc-broker producer:
```
docker run --net=host mahadeva604/grpc-broker-producer:0.0.3 -s localhost:8086 -t topic1 -k key1 -m message1
```
