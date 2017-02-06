package standup

import (
	"bytes"
	"encoding/json"
	// "fmt"
	"net/http"
	"text/template"

	// log "github.com/Sirupsen/logrus"
	"github.com/abourget/slick"
	"github.com/boltdb/bolt"
	"github.com/gigawattio/web"
	"github.com/gigawattio/web/generics"
	"github.com/gigawattio/web/route"
	"github.com/gorilla/mux"
	"github.com/nbio/hitch"
)

func (standup *Standup) activateRoutes() *hitch.Hitch {
	routes := []route.RouteMiddlewareBundle{
		route.RouteMiddlewareBundle{
			Middlewares: []func(http.Handler) http.Handler{
			// service.LoggerMiddleware,
			// web.StaticFilesMiddleware(service.staticFilesAssetProvider()),
			},
			RouteData: []route.RouteDatum{
				{"get", "/plugins/standup", standup.index},
				{"get", "/plugins/standup/v1/:date", standup.dateLookup},
				// {"get", "/plugins/standup/:date", standup.dateLookupHTML},
				// {"post", "/v1/archive/*url", service.archive},
				// {"post", "/v1/archive.json/*url", service.archiveJson},
				// {"post", "/v1/proxy/*url", service.proxy},
				// {"post", "/v1/proxy.json/*url", service.proxyJson},
			},
		},
	}
	h := route.Activate(routes)
	return h
}

func (standup *Standup) InitWebPlugin(bot *slick.Bot, privRouter *mux.Router, pubRouter *mux.Router) {
	privRouter.PathPrefix("/plugins/standup").Handler(standup.activateRoutes().Handler())
	// privRouter.Handle("/plugins/standup*", standup.activateRoutes().Handler())
	// privRouter.HandleFunc("/plugins/standup/date.json", func(w http.ResponseWriter, req *http.Request) {
	// })
}

func (standup *Standup) index(w http.ResponseWriter, req *http.Request) {
	// generics.GenericObjectEndpoint(w, req, func() (interface{}, error) {
	// 	fmt.Println(hitch.Params(req).ByName("id"))
	// 	return map[string]string{"hrm": "yep"}, nil
	// })
	t := template.Must(template.New("listing").Parse(reportTemplate))
	sm, err := standup.standupMap()
	if err != nil {
		web.RespondWithHtml(w, http.StatusInternalServerError, err.Error())
		return
	}
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, sm); err != nil {
		web.RespondWithHtml(w, http.StatusInternalServerError, err.Error())
		return
	}
	// sm["20"][0].Data.
	web.RespondWithHtml(w, http.StatusOK, buf.String())
}

const reportTemplate = `<html>
<head>
	<title>Standup Report</title>
	<style>
	table tr td:first-child {
		white-space: nowrap;
	}
	</style>
</head>
<body>
<table>
	<tr>
		<th>Date</th>
		<th>Name</th>
		<th>Updates</th>
	</tr>
	{{range $date, $entries := .}}
	{{range $i, $entry := $entries}}
	<tr>
		<td>{{$date}}</td>
		<td>{{$entry.Name}}</td>
		<td>
			<ul>
				<li>Yesterday: {{$entry.Data.Yesterday}}</li>
				<li>Today: {{$entry.Data.Today}}</li>
				<li>Blocking: {{$entry.Data.Blocking}}</li>
			</ul>
		</td>
	</tr>
	{{end}}
	{{end}}
</table>
</body>
</html>`

func (standup *Standup) dateLookup(w http.ResponseWriter, req *http.Request) {
	generics.GenericObjectEndpoint(w, req, func() (interface{}, error) {
		date := hitch.Params(req).ByName("date")

		// info := struct {
		// 	Users []*standupUser
		// }{
		// 	Users: make([]*standupUser, 0),
		// }
		sm, err := standup.standupMap()
		if err != nil {
			return nil, err
		}
		// for _, value := range *standup.data {
		// 	data.Users = append(data.Users, value)
		// }

		if entries, ok := sm[date]; ok {
			return entries, nil
		} else {
			web.RespondWithJson(w, http.StatusNotFound, web.Json{"error": "not found"})
			return nil, generics.RequestAlreadyHandled()
		}
	})
}

func (standup *Standup) standupMap() (standupMap, error) {
	sm := standupMap{}
	err := standup.bot.DB.View(func(tx *bolt.Tx) error {
		var (
			bucket = tx.Bucket([]byte(bucketKey))
			src    = bucket.Get([]byte(bucketKey))
		)
		if err := json.Unmarshal(src, &sm); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return sm, nil
}
