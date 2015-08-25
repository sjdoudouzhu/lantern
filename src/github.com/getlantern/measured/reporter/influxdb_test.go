package reporter

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getlantern/measured"
	"github.com/getlantern/testify/assert"
)

func TestWriteLineProtocol(t *testing.T) {
	chReq := make(chan []string, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := ioutil.ReadAll(r.Body)
		user, pass, ok := r.BasicAuth()
		assert.True(t, ok, "should send basic auth")
		chReq <- []string{user, pass, string(b)}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()
	ir := NewInfluxDBReporter(ts.URL, "test-user", "test-password", "testdb", nil)
	e := ir.Submit(&measured.Stats{
		Type: "errors",
		Tags: map[string]string{
			"server": "fl-nl-xxx",
			"error":  "test error",
			"empty":  "",
		},
		Fields: map[string]interface{}{"value": 3, "empty": ""}})
	assert.NoError(t, e, "", "")
	req := <-chReq
	assert.Equal(t, req[0], "test-user", "")
	assert.Equal(t, req[1], "test-password", "")
	assert.Contains(t, req[2], "errors,", "should send correct InfluxDB line protocol measurement")
	assert.Contains(t, req[2], "error=test\\ error", "should send correct InfluxDB line protocol tag")
	assert.Contains(t, req[2], "server=fl-nl-xxx", "should send correct InfluxDB line protocol tag")
	assert.Contains(t, req[2], " value=3i", "should send correct InfluxDB line protocol field")

}

func TestRealProxyServer(t *testing.T) {
	ir := NewInfluxDBReporter("https://influx.getiantem.org/", "test", "test", "lantern", nil)
	e := ir.Submit(&measured.Stats{
		Tags: map[string]string{
			"server": "fl-nl-xxx",
			"error":  "test error",
		},
		Fields: map[string]interface{}{"value": 3}})
	assert.NoError(t, e, "", "")
}
