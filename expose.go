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
			Title     string
			RegName   string
			Items     map[string]string
			Snapshots []struct {
				Ts string
				M  map[string]string
			}
		}{
			Title:   qv.Get("show") + " :: metrics",
			RegName: qv.Get("show"),
			Items:   make(map[string]string, len(reg.GetMetrics())),
		}

		for name, m := range reg.GetMetrics() {
			data.Items[name] = m.String()
		}

		switch reg.(type) {
		case Tracker:
			for _, snapshot := range reg.(Tracker).GetSnapshots() {
				ms := snapshot.GetMetrics()
				msData := make(map[string]string, len(ms))
				for name, metric := range ms {
					msData[name] = metric.String()
				}
				data.Snapshots = append(data.Snapshots, struct {
					Ts string
					M  map[string]string
				}{
					Ts: snapshot.GetTimestamp().Format("2006-01-02 15:04:05"),
					M:  msData,
				})
			}
		default:
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
		<h4>{{.RegName}}</h4>
		<h6>Current</h6>
		{{range $key, $val := .Items}}
			<div>{{ $key }}: {{ $val }}</div>
		{{else}}
			<div><strong>no metrics found</strong></div>
		{{end}}
		
		{{if .Snapshots}}
			<h6>Snapshots</h6>
		{{end}}
		{{range $key, $val := .Snapshots}}
			<div>[{{$val.Ts}}]</div>
			<div>
				{{range $k, $v := $val.M}}
					
					<div>{{ $k }}: {{ $v }}</div>
				{{end}}
			</div>
		{{end}}
		
	</body>
</html>`
