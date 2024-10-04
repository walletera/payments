package app

import "log/slog"

type Option func(app *App)

func WithRabbitMQUrl(url string) func(app *App) {
    return func(app *App) { app.rabbitmqUrl = url }
}

func WithHttpServerPort(port int) func(app *App) {
    return func(app *App) { app.httpServerPort = port }
}

func WithESDBUrl(url string) func(app *App) {
    return func(app *App) { app.esdbUrl = url }
}

func WithLogHandler(handler slog.Handler) func(app *App) {
    return func(app *App) { app.logHandler = handler }
}
