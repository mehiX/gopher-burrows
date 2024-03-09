# GopherNet - a platform for managing rentals of gopher burrows

Requires:
- Go >= 1.22.0

## Build and run

Run `make` to get a list of available options.

Build the binary with `make build`.

View available options with `./dist/burrows serve --help`

Start a server for testing: `./dist/burrows serve -v -t 3s --repos-freq 20s`

Endpoint to test the REST API:

```shell
# Show the current status
curl -s http://localhost:8080/ | jq '.'

# Rent a burrow
curl -sX POST http://127.0.0.1:8080/rent | jq '.'
```

When you are done testing press `CTRL+C` to shutdown the server. Before exiting completely the server will generate a dump file in the current directory with the current status of all the burrows. This file can then be used for successive runs.