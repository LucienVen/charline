module github.com/LucienVen/charline/server

go 1.25.5

require (
	github.com/LucienVen/charline/pkg v0.0.0
	github.com/go-chi/chi/v5 v5.2.4
	go.uber.org/zap v1.27.1
)

require go.uber.org/multierr v1.10.0 // indirect

replace github.com/LucienVen/charline/pkg => ../pkg
