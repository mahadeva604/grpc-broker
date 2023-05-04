package broker

//go:generate docker run -v $PWD:/defs namely/protoc-all -o . -f proto/broker.proto -l go
