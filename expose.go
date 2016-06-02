package metrics

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"
)

func init() {
	http.Handle("/easy-metrics", http.HandlerFunc(exposeMetrics))
}

// exposeMetrics shows all registries via http
func exposeMetrics(w http.ResponseWriter, r *http.Request) {
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
		type ChData struct {
			Index template.JS
			X     []string
			Y     []template.JS
		}
		type Charts map[template.JS]ChData
		data := struct {
			Title      string
			RegName    string
			Items      map[string]string
			Charts     Charts
			ChartNames []template.JS
			Snapshots  []struct {
				Ts string
				M  map[string]string
			}
		}{
			Title:   qv.Get("show") + " :: metrics",
			RegName: qv.Get("show"),
			Items:   make(map[string]string, len(reg.GetMetrics())),
		}

		t, _ := template.New("registries").Parse(metricsTpl)

		for name, m := range reg.GetMetrics() {
			data.Items[name] = m.String()
		}

		switch reg.(type) {
		case Tracker:
			charts := Charts{}
			for _, snapshot := range reg.(Tracker).GetSnapshots() {
				ms := snapshot.GetMetrics()
				msData := make(map[string]string, len(ms))
				idx := 1
				for name, metric := range ms {
					msData[name] = metric.String()
					ch := ChData{}
					ch.Index = template.JS(fmt.Sprintf("trace%d", idx))
					ch.X = append(charts[template.JS(name)].X, snapshot.GetTimestamp().Format("2006-01-02 15:04:05"))
					ch.Y = append(charts[template.JS(name)].Y, template.JS(metric.String()))
					charts[template.JS(name)] = ch
					idx++
				}
				data.Snapshots = append(data.Snapshots, struct {
					Ts string
					M  map[string]string
				}{
					Ts: snapshot.GetTimestamp().Format("2006-01-02 15:04:05"),
					M:  msData,
				})
			}
			data.Charts = charts

			for _, cc := range charts {
				data.ChartNames = append(data.ChartNames, cc.Index)
			}

		default:
		}

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
	<body style="font-family:Arial,Helvetica,sans-serif;font-size:16px;">
	<div style="font-size:20px; padding: 10px">Registries:</div>
	<ul>
		{{range .Items}}<li><a href="/easy-metrics?show={{ . | urlquery }}">{{ . }}</a></li>{{else}}<li><strong>no registries</strong></li>{{end}}
		</ul>
	</body>
</html>`

const metricsTpl = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>{{.Title}}</title>
		{{if .Snapshots}}
		<script src="https://cdn.plot.ly/plotly-1.12.0.min.js"></script>
		{{end}}
	</head>
	<body style="font-family:Arial,Helvetica,sans-serif;font-size:14px;margin:0;padding:0">
		<h1 style="font-size: 26px;font-weight:500;margin: 0 0 10px 0;padding: 15px 0 10px 20px;text-align: left;position: relative;box-shadow: 0px 3px 19px -9px rgba(0,0,0,.3);z-index: 2;background: #fff">{{.RegName}}</h1>
		<div style="float:left;margin: -10px 0 0 0;padding: 30px 35px 20px 20px;position: relative;z-index: 1;box-shadow: -1px -9px 19px 4px rgba(0,0,0,.15);min-height: 550px;font-family:monospace">
			<div style="font:18px Arial,Helvetica,sans-serif;margin:10px 0 10px 0;padding: 0;">Current:</div>
			{{range $key, $val := .Items}}
				<div>{{ $key }}: {{ $val }}</div>
			{{else}}
				<div><strong>no metrics found</strong></div>
			{{end}}
			
			{{if .Snapshots}}
				<div style="font:18px Arial,Helvetica,sans-serif;margin:20px 0 0 0;padding:0;">Snapshots:</div>
			{{end}}
			{{range $key, $val := .Snapshots}}
				<div style="margin-top:10px;font-size:12px">[{{$val.Ts}}]</div>
				<div>
					{{range $k, $v := $val.M}}
						<div>{{ $k }}: {{ $v }}</div>
					{{end}}
				</div>
			{{end}}
		</div>
		{{if .Snapshots}}
			<div id="chartsDiv" style="position: fixed;margin: -60px auto 0 auto;left: 20%;top: 130px;width:70%"></div>
			<script>
				{{range $name, $data := .Charts}}
					var {{$data.Index}} = {
					x: [
						{{range $index, $v := $data.X}}
							{{if $index}},{{end}}{{$v}}
						{{end}}
					], 
					y: [
						{{range $index, $v := $data.Y}}
							{{if $index}},{{end}}{{$v}}
						{{end}}
					], 
					type: 'scatter',
					name: '{{$name}}'
					};
				{{end}}
				var data = [
				//	trace1, trace2, trace3, trace4
				{{range $index, $v := .ChartNames}}
							{{if $index}},{{end}}{{$v}}
				{{end}}
				];
				Plotly.newPlot('chartsDiv', data);
			</script>
		{{end}}
	</body>
</html>`
