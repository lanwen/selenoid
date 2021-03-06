package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/aandryashin/matchers"
	. "github.com/aandryashin/matchers/httpresp"
	"github.com/aandryashin/selenoid/config"
)

var (
	srv *httptest.Server
)

func init() {
	srv = httptest.NewServer(handler())
}

func TestNewSessionWithGet(t *testing.T) {
	manager = &HTTPTest{Handler: Selenium()}

	rsp, err := http.Get(With(srv.URL).Path("/wd/hub/session"))
	AssertThat(t, err, Is{nil})
	AssertThat(t, rsp, Code{http.StatusMethodNotAllowed})

	AssertThat(t, queue.Used(), EqualTo{0})
}

func TestBadJsonFormat(t *testing.T) {
	rsp, err := http.Post(With(srv.URL).Path("/wd/hub/session"), "", nil)
	AssertThat(t, err, Is{nil})
	AssertThat(t, rsp, Code{http.StatusBadRequest})

	AssertThat(t, queue.Used(), EqualTo{0})
}

func TestServiceStartupFailure(t *testing.T) {
	manager = &StartupError{}

	rsp, err := http.Post(With(srv.URL).Path("/wd/hub/session"), "", bytes.NewReader([]byte("{}")))
	AssertThat(t, err, Is{nil})
	AssertThat(t, rsp, Code{http.StatusInternalServerError})

	AssertThat(t, queue.Used(), EqualTo{0})
}

func TestBrowserNotFound(t *testing.T) {
	manager = &BrowserNotFound{}

	rsp, err := http.Post(With(srv.URL).Path("/wd/hub/session"), "", bytes.NewReader([]byte("{}")))
	AssertThat(t, err, Is{nil})
	AssertThat(t, rsp, Code{http.StatusBadRequest})

	AssertThat(t, queue.Used(), EqualTo{0})
}

func TestMalformedScreeResolutionCapability(t *testing.T) {
	manager = &BrowserNotFound{}

	rsp, err := http.Post(With(srv.URL).Path("/wd/hub/session"), "", bytes.NewReader([]byte(`{"desiredCapabilities":{"screenResolution":"1024x768"}}`)))
	AssertThat(t, err, Is{nil})
	AssertThat(t, rsp, Code{http.StatusBadRequest})

	AssertThat(t, queue.Used(), EqualTo{0})
}

func TestNewSessionNotFound(t *testing.T) {
	manager = &HTTPTest{Handler: Selenium()}

	rsp, err := http.Get(With(srv.URL).Path("/wd/hub/session/123"))
	AssertThat(t, err, Is{nil})
	AssertThat(t, rsp, Code{http.StatusNotFound})

	AssertThat(t, queue.Used(), EqualTo{0})
}

func TestNewSessionHostDown(t *testing.T) {
	canceled := false
	ch := make(chan bool)
	manager = &HTTPTest{
		Handler: Selenium(),
		Action: func(s *httptest.Server) {
			log.Println("Host is going down...")
			s.Close()
			log.Println("Now Host is down...")
		},
		Cancel: ch,
	}

	rsp, err := http.Post(With(srv.URL).Path("/wd/hub/session"), "", bytes.NewReader([]byte("{}")))
	AssertThat(t, err, Is{nil})
	AssertThat(t, rsp, Code{http.StatusInternalServerError})

	canceled = <-ch
	AssertThat(t, canceled, Is{true})

	AssertThat(t, queue.Used(), EqualTo{0})
}

func TestNewSessionBadHostResponse(t *testing.T) {
	canceled := false
	ch := make(chan bool)
	manager = &HTTPTest{
		Handler: HTTPResponse("Bad Request", http.StatusBadRequest),
		Cancel:  ch,
	}

	rsp, err := http.Post(With(srv.URL).Path("/wd/hub/session"), "", bytes.NewReader([]byte("{}")))
	AssertThat(t, err, Is{nil})
	AssertThat(t, rsp, Code{http.StatusBadRequest})

	canceled = <-ch
	AssertThat(t, canceled, Is{true})

	AssertThat(t, queue.Used(), EqualTo{0})
}

func TestSessionCreated(t *testing.T) {
	manager = &HTTPTest{Handler: Selenium()}

	resp, err := http.Post(With(srv.URL).Path("/wd/hub/session"), "", bytes.NewReader([]byte("{}")))
	AssertThat(t, err, Is{nil})
	var sess map[string]string
	AssertThat(t, resp, AllOf{Code{http.StatusOK}, IsJson{&sess}})

	resp, err = http.Get(With(srv.URL).Path("/status"))
	AssertThat(t, err, Is{nil})
	var state config.State
	AssertThat(t, resp, AllOf{Code{http.StatusOK}, IsJson{&state}})
	AssertThat(t, state.Used, EqualTo{1})
	AssertThat(t, queue.Used(), EqualTo{1})
	sessions.Remove(sess["sessionId"])
	queue.Release()
}

func TestProxySession(t *testing.T) {
	manager = &HTTPTest{Handler: Selenium()}

	resp, err := http.Post(With(srv.URL).Path("/wd/hub/session"), "", bytes.NewReader([]byte("{}")))
	AssertThat(t, err, Is{nil})
	var sess map[string]string
	AssertThat(t, resp, AllOf{Code{http.StatusOK}, IsJson{&sess}})

	resp, err = http.Get(With(srv.URL).Path(fmt.Sprintf("/wd/hub/session/%s/url", sess["sessionId"])))
	AssertThat(t, err, Is{nil})
	AssertThat(t, resp, Code{http.StatusOK})

	AssertThat(t, queue.Used(), EqualTo{1})
	sessions.Remove(sess["sessionId"])
	queue.Release()
}

func TestSessionDeleted(t *testing.T) {
	canceled := false
	ch := make(chan bool)
	manager = &HTTPTest{
		Handler: Selenium(),
		Cancel:  ch,
	}

	resp, err := http.Post(With(srv.URL).Path("/wd/hub/session"), "", bytes.NewReader([]byte("{}")))
	AssertThat(t, err, Is{nil})
	var sess map[string]string
	AssertThat(t, resp, AllOf{Code{http.StatusOK}, IsJson{&sess}})

	req, _ := http.NewRequest(http.MethodDelete,
		With(srv.URL).Path(fmt.Sprintf("/wd/hub/session/%s", sess["sessionId"])), nil)
	http.DefaultClient.Do(req)

	resp, err = http.Get(With(srv.URL).Path("/status"))
	AssertThat(t, err, Is{nil})
	var state config.State
	AssertThat(t, resp, AllOf{Code{http.StatusOK}, IsJson{&state}})
	AssertThat(t, state.Used, EqualTo{0})

	canceled = <-ch
	AssertThat(t, canceled, Is{true})

	AssertThat(t, queue.Used(), EqualTo{0})
}

func TestSessionOnClose(t *testing.T) {
	manager = &HTTPTest{Handler: Selenium()}

	resp, err := http.Post(With(srv.URL).Path("/wd/hub/session"), "", bytes.NewReader([]byte("{}")))
	AssertThat(t, err, Is{nil})
	var sess map[string]string
	AssertThat(t, resp, AllOf{Code{http.StatusOK}, IsJson{&sess}})

	req, _ := http.NewRequest(http.MethodDelete,
		With(srv.URL).Path(fmt.Sprintf("/wd/hub/session/%s/window", sess["sessionId"])), nil)
	http.DefaultClient.Do(req)

	AssertThat(t, queue.Used(), EqualTo{1})
	sessions.Remove(sess["sessionId"])
	queue.Release()
}

func TestProxySessionCanceled(t *testing.T) {
	canceled := false
	ch := make(chan bool)
	manager = &HTTPTest{
		Handler: Selenium(),
		Cancel:  ch,
	}

	timeout = 100 * time.Millisecond
	resp, err := http.Post(With(srv.URL).Path("/wd/hub/session"), "", bytes.NewReader([]byte("{}")))
	AssertThat(t, err, Is{nil})
	var sess map[string]string
	AssertThat(t, resp, AllOf{Code{http.StatusOK}, IsJson{&sess}})

	_, ok := sessions.Get(sess["sessionId"])
	AssertThat(t, ok, Is{true})

	req, _ := http.NewRequest(http.MethodGet, With(srv.URL).Path(fmt.Sprintf("/wd/hub/session/%s/url?timeout=1s", sess["sessionId"])), nil)
	ctx, cancel := context.WithCancel(context.Background())
	req = req.WithContext(ctx)
	go func() {
		http.DefaultClient.Do(req)
	}()
	<-time.After(50 * time.Millisecond)
	cancel()
	<-time.After(100 * time.Millisecond)
	_, ok = sessions.Get(sess["sessionId"])
	AssertThat(t, ok, Is{false})

	canceled = <-ch
	AssertThat(t, canceled, Is{true})

	AssertThat(t, queue.Used(), EqualTo{0})
}

func TestNewSessionTimeout(t *testing.T) {
	canceled := false
	ch := make(chan bool)
	manager = &HTTPTest{
		Handler: Selenium(),
		Cancel:  ch,
	}

	timeout = 30 * time.Millisecond
	resp, err := http.Post(With(srv.URL).Path("/wd/hub/session"), "", bytes.NewReader([]byte("{}")))
	AssertThat(t, err, Is{nil})
	var sess map[string]string
	AssertThat(t, resp, AllOf{Code{http.StatusOK}, IsJson{&sess}})

	_, ok := sessions.Get(sess["sessionId"])
	AssertThat(t, ok, Is{true})

	<-time.After(50 * time.Millisecond)
	_, ok = sessions.Get(sess["sessionId"])
	AssertThat(t, ok, Is{false})

	canceled = <-ch
	AssertThat(t, canceled, Is{true})

	AssertThat(t, queue.Used(), EqualTo{0})
}

func TestProxySessionTimeout(t *testing.T) {
	canceled := false
	ch := make(chan bool)
	manager = &HTTPTest{
		Handler: Selenium(),
		Cancel:  ch,
	}

	timeout = 30 * time.Millisecond
	resp, err := http.Post(With(srv.URL).Path("/wd/hub/session"), "", bytes.NewReader([]byte("{}")))
	AssertThat(t, err, Is{nil})
	var sess map[string]string
	AssertThat(t, resp, AllOf{Code{http.StatusOK}, IsJson{&sess}})

	_, ok := sessions.Get(sess["sessionId"])
	AssertThat(t, ok, Is{true})

	<-time.After(20 * time.Millisecond)
	_, ok = sessions.Get(sess["sessionId"])
	AssertThat(t, ok, Is{true})
	http.Get(With(srv.URL).Path(fmt.Sprintf("/wd/hub/session/%s/url", sess["sessionId"])))

	<-time.After(20 * time.Millisecond)
	_, ok = sessions.Get(sess["sessionId"])
	AssertThat(t, ok, Is{true})

	<-time.After(50 * time.Millisecond)
	_, ok = sessions.Get(sess["sessionId"])
	AssertThat(t, ok, Is{false})

	canceled = <-ch
	AssertThat(t, canceled, Is{true})

	AssertThat(t, queue.Used(), EqualTo{0})
}
