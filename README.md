`snakepit` is a [cobra](https://github.com/spf13/cobra)/[viper](https://github.com/spf13/viper) commands repository.

## Utils

Besides the `cobra` commands, `snakepit` offers a few utils useful for building web APIs:

- A standardized API error format
- A lightweight [ffjson](https://github.com/pquerna/ffjson) based JSON renderer


## Commands
### Root

The `root` command is meant to be the entrypoint of your `cobra` based app.
It exports a `viper` instance allowing other commands to build on it.

### Run

The `run` command is a simple [graceful](https://github.com/tylerb/graceful) server building and running a HTTP handler.
The `Builder` is typically set from a local `run` command in your app.

## License

MIT
