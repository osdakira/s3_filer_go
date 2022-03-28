# s3_filer_go

TUI tool to view S3.

## Function

- Save the last directory accessed.
- Filters paths for easier access.
- View S3 files without downloading the first part of `ascii text`, `gzip`, or `parquet` (like the head command)
    - `parquet` downloads everything, so it may take longer for large files.

## Install

1. You can download binary from [release page](https://github.com/osdakira/s3_filer_go/releases) and place it in $PATH directory.
2. You will need to set up your configurations and credentials just as you would with awscli.

```
$ aws configure
```

## Usage

Run the command.

```
$ s3_filer
```

If the token is shown as expired at that time, please reacquire the token.

```
$ s3_filer
2022/03/28 09:16:18 operation error S3: ListBuckets, https response error StatusCode: 400, RequestID: ..., HostID: ..., api error ExpiredToken: The provided token has expired.
```

When the screen appears, the following commands can be executed.

- Ctrl+P: select (P)previous item
- Ctrl+N: select (N)ext item
- Ctrl+U: go (U)p to parent directory
- Enter: into item
    - directory: Go to a directory
    - file: Display the first 500 KB of the file
- Ctrl+D: (D)ownload file
- Ctrl+H: toggle file size to (H)uman readable
- Ctrl+Q: (Q)uit application with save the path

- cursor movement keys:
    - KeyUp: select (P)previous item
    - KeyDown: select (N)ext item
    - KeyLeft: go (U)p to parent directory
    - KeyRight: into item

# Development

```
$ docker-compose run --rm app
Creating s3_filer_go_app_run ... done

# make build
export GO111MODULE=on
env GOOS=linux go build -ldflags="-s -w" -o bin/s3_filer s3_filer/*.go
```
