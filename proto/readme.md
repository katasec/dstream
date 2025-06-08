# Command for generating go files from proto in the right folder

Run below from repo root:

```
protoc --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. proto/plugin.proto
```
