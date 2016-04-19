`snakepit` is a [cobra](https://github.com/spf13/cobra)/[viper](https://github.com/spf13/viper) commands repository and a productivity oriented toolbox.

API example seed: [snakepit-seed](https://github.com/solher/snakepit-seed)

## Commands
### Root

The `root` command is meant to be the entrypoint of your `cobra` based app.
It exports a `viper` instance allowing other commands to build on it.

### Run

The `run` command is a simple [graceful](https://github.com/tylerb/graceful) server building and running a HTTP handler.
The `Builder` is typically set from a local `run` command in your app.

## Toolbox

Besides the `cobra` commands, `snakepit` offers utils to build expressive web APIs:

- A standardized API error format.
- A suite of [net/context](https://godoc.org/golang.org/x/net/context) based middlewares:
    - `swagger` to expose [Swagger](http://swagger.io) documentation on `/swagger`.
    - `requestID`, inspired by the one from [Goji](https://github.com/zenazn/goji), to uniquely tag each request.
    - `logger` using [logrus](https://github.com/Sirupsen/logrus) setting a `requestID` tagged logger (if existing) in the request context.
    - `recoverer` recovering from panics, logging the error if `logger` is present and sending standardized `500` errors.
    - `timer` mesuring the middleware stack processing time and logging it if `logger` is present.
- A [ffjson](https://github.com/pquerna/ffjson) based JSON marshaller/unmarshaller that automatically log processing times if the `logger` middleware is present in the middleware stack and returns standardized `400` errors when unmarshallings fails. Also supports bulk requests unmarshalling.

## TODOs
* [ ] Write tests
* [ ] Add a lot of documentation

## License

MIT
