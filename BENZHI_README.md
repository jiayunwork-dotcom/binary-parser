# binary-parser — Go 二进制文件格式解析与结构查询 HTTP 服务

Binary container format parser and validator with record indexing, diff,
merge, tree visualization, query, and transformation capabilities.
Provides HTTP API for binary file parsing and structural validation.

## Build / Run / Test

```bash
go build -o binary-parser .
./binary-parser serve -addr :8080
./binary-parser parse -input example/sample.bin
go test ./...
```

## Evaluation Image

Evaluation-specific files (do not overwrite project Dockerfile/README):

- `benzhi.Dockerfile`
- `build_benzhi_docker.sh`
- `BENZHI_README.md` (this file)

Build and verify in container:

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh <image-name> linux/arm64
./build_benzhi_docker.sh <image-name> linux/amd64
docker run -it <image-name>:latest
```
