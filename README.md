# le
`le` is a simple file server written in Go with
* resume support
* local address QR code
* download logs with progress

## Usage

```sh
go run go.sakib.dev/le
```

## Optional parameters
- `--dir`: Directory to serve files from (default: current directory)
- `--port`: Port to run the server on (default: 8080)
