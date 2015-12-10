package proxy

import (
	"net/http"
	"io/ioutil"
	"io"
	"strings"
	"compress/gzip"
	"bytes"
	"strconv"
	"log"
	"errors"
	"regexp"
	"crypto/tls"

	"github.com/PuerkitoBio/goquery"
	"github.com/djimenez/iconv-go"
	"net/url"
	"fmt"
)

// Proxy is the main structure
type Proxy struct {
	Config Config
}

// New creates a new proxy instance
func New(c Config) Proxy {
	return Proxy{c}
}

// Handler must be linked to the http server to handle the proxy main logic
func (p Proxy) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := normalizeURL(r.URL)

		res, err := request(r)
		if nil != err {
			log.Printf("Error requesting %s because of %s", url, err)
			if e := passthrough(res, w); nil != e {
				log.Printf("Cannot passthrough %s because of %s", url, e)
			}
			return
		}

		ct := contentType(res.Header)
		if !isHtml(ct) {
			if e := passthrough(res, w); nil != e {
				log.Printf("Cannot passthrough %s because of %s", url, e)
			}
			return
		}

		enc := encoding(res.Header);
		c, err := read(enc, ct, res.Body)
		if nil != err {
			log.Printf("Cannot read content of %s because of %s", url, err)
			if e := passthrough(res, w); nil != e {
				log.Printf("Cannot passthrough %s because of %s", url, e)
			}
			return
		}

		err = p.proxy(w, c, res.Header, url)
		if nil != err {
			log.Printf("Cannot proxy content for %s because of %s", url, err);
			if e := passthrough(res, w); nil != e {
				log.Printf("Cannot passthrough %s because of %s", url, e)
			}
			return
		}
	}
}

// encoding extracts the Content-Encoding
func encoding(h http.Header) string {
	return h.Get("Content-Encoding")
}

// contentType extract the Content-Type
func contentType(h http.Header) string {
	return h.Get("Content-Type")
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

// isHtml checks if the document is an html document
func isHtml(ct string) bool {
	return strings.Contains(ct, "text/html")
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

// proxy forwards the manipulated content to the output writer
func (p Proxy) proxy(w http.ResponseWriter, content string, header http.Header, url string) error {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if nil != err {
		return err
	}

	p.modify(doc, url)

	c, err := doc.Html()
	if nil != err {
		return err
	}

	if encoding(header) == "gzip" {
		var b bytes.Buffer
		tw := gzip.NewWriter(&b)
		if _, err := tw.Write([]byte(content)); err != nil {
			return err
		}
		if err := tw.Flush(); err != nil {
			return err
		}
		if err := tw.Close(); err != nil {
			return err
		}
		c = b.String()
	}

	for h, v := range header {
		switch h {
		case "Content-Length":
			w.Header().Add("Content-Length", strconv.Itoa(len(c)))
			break;
		default:
			for _, val := range v {
				w.Header().Add(h, val)
			}
		}
	}

	io.WriteString(w, c)
	return nil
}

// modify manipulates the document based on the configuration
func (p Proxy) modify(d *goquery.Document, url string) {
	for _, mod := range p.Config.Modifications {
		if m, err := regexp.MatchString(mod.URLMatch, url); nil == err && !m {
			continue;
		}
		i := mod.Index
		e := d.Find(mod.Selector)
		if i > e.Length() {
			log.Printf("No element found for selector %s on index %d (%s)", mod.Selector, i, url)
			continue;
		}
		elm := e.Get(i)

		t, tMode, err := target(mod, d)
		if nil != err {
			log.Printf("%s (%s)", err, url)
			continue;
		}

		var v string
		if 0 < len(mod.Attribute) {
			for _, attr := range elm.Attr {
				if attr.Key == mod.Attribute {
					v = attr.Val
				}
			}
		} else {
			v = elm.FirstChild.Data
		}
		if (mod.Trim) {
			v = strings.TrimSpace(v)
		}
		if 0 < len(mod.Wrapper) && strings.Contains(mod.Wrapper, "%s") {
			v = fmt.Sprintf(mod.Wrapper, v)
		}

		switch tMode {
		case APPEND:
			t.AppendHtml(v)
			break;
		case REPLACE:
			t.ReplaceWithHtml(v)
			break;
		default:
			log.Printf("No action (replace/append) found for %s", url)
		}
	}
}

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