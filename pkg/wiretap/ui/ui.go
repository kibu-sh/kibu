package ui

import (
	"embed"
	"github.com/NYTimes/gziphandler"
	"github.com/samber/lo"
	"io/fs"
	"net/http"
)

//go:generate pnpm build:unsafe

//go:embed dist
var uiFS embed.FS
var dist = lo.Must(fs.Sub(uiFS, "dist"))

func Handler() http.Handler {
	return gziphandler.GzipHandler(http.FileServer(http.FS(dist)))
}
