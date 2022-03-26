# s3_filer_go

TUI tool to view S3.

## Function

- Save the last directory accessed.
- Filters paths for easier access.
- View S3 files without downloading the first part of `ascii text`, `gzip`, or `parquet` (like the head command)
    - `parquet` downloads everything, so it may take longer for large files.

## Usage

Use Ctrl key to function.

- Ctrl+P: select (P)previous item
- Ctrl+N: select (N)ext item
- Ctrl+U: go (U)p to parent directory
- Enter: into item
    - directory: Go to a directory
    - file: Display the first 500 KB of the file
- Ctrl+D: (D)ownload file
- Ctrl+H: toggle file size to (H)uman readable

- cursor movement keys:
    - KeyUp: select (P)previous item
    - KeyDown: select (N)ext item
    - KeyLeft: go (U)p to parent directory
    - KeyRight: into item

# Development

```
$ docker-compose run --rm app
Creating s3_filer_go_app_run ... done
root@36e0ffe445ad:/usr/src/app# make build
export GO111MODULE=on
env GOOS=linux go build -ldflags="-s -w" -o bin/s3_filer s3_filer/*.go
```
