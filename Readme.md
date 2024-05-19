# Gwatch - Lightweight Auto reloader for Go applications

`Gwatch` is a lightweight, super-fast and highly configurable auto reloader for your Go applications. It monitors your Go application for changes and automatically recompile and restarts your application, ensuring you see your changes instantly.

## Installation

### - Via `go install` (Recommended)

This downloads gwatch, compiles and installs the binary in your `$GOPATH/bin` directory.

```bash
go install github.com/huboh/gwatch/cmd/gwatch@latest
```

### - Via `curl`

```bash
curl -sSL https://raw.githubusercontent.com/huboh/gwatch/main/install.sh | sh
```

## Usage

simply run gwatch in the root of your Go application. If the configuration file `gwatch.yml` does not exist in the current directory, gwatch will automatically generate one with it's default values.

```bash
gwatch
```

## Configuration (`gwatch.yml`)

`gwatch` uses a YAML configuration file (gwatch.yml) to define its behavior. the config is automatically generated with default value in the current directory where `gwatch` is executed.

Here's an example configuration file:

```yaml
# The root directory of your application
root: ./

# The build command to compile your application
build:
  cmd: go build -o ./bin/app main.go

# The command to run your application
run:
  bin: ./bin/app
  args: []

# The file extensions to watch for changes
exts:
  - go
  - tmp
  - tmpl
  - html

# The paths to watch
paths:
  - ./

# The wait delay duration before running commands after detecting changes
delay: 100ms

# The output prefix for your app log messages
log_prefix: your-app

# The directories to exclude from watching
exclude:
  - .git
  - bin
  - vendor
  - testdata

# Watch files recursively
recursive: true
```

## Features

- nice cli
- super-fast
- customizable log message prefix
- customizable build and run commands along with other configs
- gwatch will auto restart itself if it detect changes to it's config file

## Author

[Huboh](https://huboh.vercel.app/)

## Contributing

contributions are welcomed! Please fork the repository and submit pull requests. A typical workflow is:

1. Fork the repository. see [here]((http://blog.campoy.cat/2014/03/github-and-go-forking-pull-requests-and.html)) for tips.
2. [Create your feature branch.](http://learn.github.com/p/branching.html)
3. Add tests for your change.
4. Run `go test`. If your tests pass, return to the step 3.
5. Implement the change and ensure the steps from the previous step pass.
6. Add, commit and push your changes.
7. Submit a pull request.

## License

This project is licensed under the BSD-3-Clause License - see the [LICENSE](https://github.com/huboh/gwatch/blob/main/LICENCE) file for details.
