package metrics

import (
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
		reg, err := GetRegistryByName(qv.Get("show"))
		if err != nil {
			switch err.(type) {
			case ErrEmptyRegistryName:
				http.Error(w, err.Error(), http.StatusNotAcceptable)
			case ErrRegistryUnknown:
				http.Error(w, err.Error(), http.StatusNotFound)
			default:
				http.Error(w, "Error", http.StatusInternalServerError)
			}
			return
		}

		t, _ := template.New("registries").Parse(metricsTpl)
		data := struct {
			Title string
			Items map[string]string
		}{
			Title: qv.Get("show") + "metrics",
			Items: make(map[string]string, len(reg.GetMetrics())),
		}

		for name, m := range reg.GetMetrics() {
			data.Items[name] = m.String()
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		t.Execute(w, data)
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

const metricsTpl = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>{{.Title}}</title>
	</head>
	<body>
		{{range $key, $val := .Items}}
			<div>{{ $key }}: {{ $val }}</div>
		{{else}}
			<div><strong>no metrics found</strong></div>
		{{end}}
	</body>
</html>`
