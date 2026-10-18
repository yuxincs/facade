# Facade

An on-demand game server proxy written in Go. Facade aims to keep a lightweight
public endpoint available while starting the actual game server when players
join and stopping it after it becomes idle.

The project is initialized with a CLI and static binary build. Network forwarding,
game adapters, configuration discovery, and server management are planned.

Running `facade` without arguments invokes the engine with the Palworld adapter.
The engine is currently a placeholder and exits with an unsupported-operation error.

## Development

Requires Go 1.27 or newer.

```sh
go run ./cmd/facade -help
make check
make build
./bin/facade -version
```

GitHub Actions runs `make check` and `make build` on pushes and pull requests.
Formatting checks cover all Go files in the repository and fail without rewriting
files. Run `make fmt` to fix formatting locally.

Set a version when building a release:

```sh
make build VERSION=0.1.0
```

## Docker

The multi-stage Docker build produces a static executable with `CGO_ENABLED=0`
and copies it into a `scratch` image running as a non-root user.

```sh
docker build -t facade .
docker run --rm facade -version
```

Build Linux images for both common server architectures with Docker Buildx:

```sh
docker buildx build --platform linux/amd64,linux/arm64 -t facade .
```

Add `--push` and a registry-qualified tag when publishing, or use a single
platform with `--load` to load an image into the local Docker daemon.

## Design direction

- Forward TCP streams and UDP datagrams with minimal allocation and copying.
- Parse only the packets needed for discovery, join detection, and startup hints.
- Separate transport, game adapters, and server lifecycle management.
- Support Minecraft Java, Minecraft Bedrock, and Palworld through game adapters.
- Offer generic forwarding for unsupported games, with conservative activity detection.
- Infer settings from the server directory and launch arguments, with explicit overrides.
- Keep player monitoring failures distinct from a confirmed empty server.

Minecraft codec dependencies will be selected as adapters are implemented. The
application uses only the Go standard library; tests use `testify/require`.

## License

[MIT](LICENSE)
