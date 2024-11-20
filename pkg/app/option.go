package app

type opt func(app *App)

func WithID(id string) opt {
	return func(app *App) {
		app.id = id
	}
}

func WithName(name string) opt {
	return func(app *App) {
		app.name = name
	}
}

func WithVersion(version string) opt {
	return func(app *App) {
		app.version = version
	}
}

func WithDebug(debug bool) opt {
	return func(app *App) {
		app.debug = debug
	}
}

func WithEndpoint(endpoint []string) opt {
	return func(app *App) {
		app.endpoint = endpoint
	}
}

func WithMetadata(metadata map[string]string) opt {
	return func(app *App) {
		app.metadata = metadata
	}
}
