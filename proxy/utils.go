package proxy

import (
	"strings"
	"github.com/djimenez/iconv-go"
	"log"
	"io/ioutil"
	"net/http"
	"errors"
	"io"
	"crypto/tls"
	"net/url"
	"compress/gzip"
	"github.com/PuerkitoBio/goquery"
	"fmt"
)

// encoding extracts the Content-Encoding
func encoding(h http.Header) string {
	return h.Get("Content-Encoding")
}

// contentType extract the Content-Type
func contentType(h http.Header) string {
	return h.Get("Content-Type")
}

// isHtml checks if the document is an html document
func isHtml(ct string) bool {
	return strings.Contains(ct, "text/html")
}

// normalizeURL normalizes the url and checks if all parameters are set correctly
func normalizeURL(url *url.URL) string {
	if "" == url.Scheme {
		if strings.Contains(url.String(), ":443") {
			url.Scheme = "https"
		} else {
			url.Scheme = "http"
		}
	}
	return url.String()
}

// read reads a stream based on its encoding
func read(enc string, ct string, body io.ReadCloser) (string, error) {
	defer body.Close()
	var cs string
	if strings.Contains(ct, "charset") {
		cs = ct[(strings.Index(ct, "charset=") + 8):]
	} else {
		cs = "ISO-8859-1"
	}

	var s io.Reader
	if enc == "gzip" {
		reader, err := gzip.NewReader(body)
		if nil != err {
			return "", err
		}
		s = reader
	} else {
		s = body
	}
	if strings.ToLower(cs) != "utf-8" {
		us, err := iconv.NewReader(s, cs, "utf-8")
		if nil == err {
			s = us
		} else {
			log.Printf("Cannot convert %s to utf-8", ct)
		}
	}
	b, err := ioutil.ReadAll(s)
	if nil != err {
		return "", err
	}
	return string(b), nil
}

// passthrough is a fallback if the document should not be handled
func passthrough(r *http.Response, w http.ResponseWriter) error {
	if nil == r {
		return errors.New("Response is unset")
	}
	for h, v := range r.Header {
		for _, val := range v {
			w.Header().Add(h, val)
		}
	}
	io.Copy(w, r.Body)
	return nil
}

// request executes an http request based on the initial proxy call
func request(r *http.Request) (*http.Response, error) {
	url := r.URL.String()
	header := r.Header
	method := r.Method

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest(method, url, nil)
	if nil != err {
		return nil, err
	}

	for h, v := range header {
		for _, val := range v {
			req.Header.Add(h, val)
		}
	}

	return client.Do(req)
}

// target gets the correct target selection and the type of replace or append
func target(mod Modification, d *goquery.Document) (*goquery.Selection, int, error) {
	var m int
	var s string
	var t *goquery.Selection
	if 0 < len(mod.AppendTo) {
		m = APPEND
		s = mod.AppendTo
		t = d.Find(mod.AppendTo)
	} else if 0 < len(mod.Replace) {
		m = REPLACE
		s = mod.Replace
		t = d.Find(mod.Replace)
	}

	if 0 == t.Length() {
		return nil, m, errors.New(fmt.Sprintf("No element found for selector %s", s))
	}
	return t, m, nil
}