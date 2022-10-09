package graphiql

import (
	"github.com/expeditioneer/gentoo-soko/pkg/config"
	"html/template"
	"net/http"
)

func Show(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")

	templates := template.Must(
		template.New("graphiql").
			ParseGlob("web/templates/api/explore/*.tmpl"))

	templates.ExecuteTemplate(w, "graphiql.tmpl", template.URL(config.GraphiqlEndpoint()))
}
