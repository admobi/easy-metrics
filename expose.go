package easy-metrics

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"
)

func init() {
	http.Handle("/metrics", http.HandlerFunc(Index))
}

// Index shows all registries via http
func Index(w http.ResponseWriter, r *http.Request) {
	qv := r.URL.Query()
	if _, ok := qv["show"]; !ok {
		// Shows main page with registries list
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		t, _ := template.New("registries").Parse(listTpl)
		data := struct {
			Title string
			Items []string
		}{
			Title: "Registries",
		}

		for names := range GetRegistries() {
			data.Items = append(data.Items, names)
		}
		sort.Strings(data.Items)

		t.Execute(w, data)

	} else {
		// Shows the registry metrics
		d, err := DumpRegistry(qv.Get("show"))
		if err != nil {
			fmt.Println(err)
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(d))

		// w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
}

// Template for registries list
const listTpl = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>{{.Title}}</title>
	</head>
	<body>
		{{range .Items}}<div><a href="/metrics?show={{ . }}">{{ . }}</a></div>{{else}}<div><strong>no registries</strong></div>{{end}}
	</body>
</html>`
