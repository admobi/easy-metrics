package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestExpose(t *testing.T) {
	r, err := NewTrackRegistry("httptestreg", 10, time.Millisecond, false)
	if err != nil {
		t.Errorf("unable to create registry: %s", err)
	}

	c := NewCounter("httpcounter")
	r.AddMetrics(c)

	listReq, err := http.NewRequest("GET", "http://example.com/easy-metrics", nil)
	if err != nil {
		t.Errorf("unable to create request: %s", err)
	}

	w1 := httptest.NewRecorder()
	exposeMetrics(w1, listReq)

	if w1.Code != 200 {
		t.Errorf("unable to request expose handler. Got code %v, %v", w1.Code, w1.Body)
	}

	wrongReq1, err := http.NewRequest("GET", "http://example.com/easy-metrics?show=unset", nil)
	if err != nil {
		t.Errorf("unable to create request: %s", err)
	}
	w2 := httptest.NewRecorder()
	exposeMetrics(w2, wrongReq1)
	if w2.Code != 404 {
		t.Errorf("request error, should be 404 code, but got: %v", w2.Code)
	}

	wrongReq2, err := http.NewRequest("GET", "http://example.com/easy-metrics?show=", nil)
	if err != nil {
		t.Errorf("unable to create request: %s", err)
	}
	w3 := httptest.NewRecorder()
	exposeMetrics(w3, wrongReq2)
	if w3.Code != 406 {
		t.Errorf("request error, should be 406 code, but got: %v", w3.Code)
	}

	chartsReq, err := http.NewRequest("GET", "http://example.com/easy-metrics?show=httptestreg", nil)
	if err != nil {
		t.Errorf("unable to create request: %s", err)
	}

	time.Sleep(time.Second)
	w4 := httptest.NewRecorder()
	exposeMetrics(w4, chartsReq)
	if w4.Code != 200 {
		t.Errorf("request error, should be 200 code, but got: %v", w4.Code)
	}
}
