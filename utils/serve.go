package utils

import (
	"fmt"
	"math/rand"
	"net/http"

	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/sirupsen/logrus"
)

func ServeGraphPage(page *components.Page, contents string) (string, *http.Server) {
	port := rand.Intn(20000) + 20000
	bindURL := fmt.Sprintf(":%d", port)

	srv := &http.Server{
		Addr: bindURL,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logrus.Debugf("Render page at %s", bindURL)
			page.Render(w)
			w.Write([]byte(contents))
		}),
	}

	go srv.ListenAndServe()

	return fmt.Sprintf("http://localhost%s", bindURL), srv
}
