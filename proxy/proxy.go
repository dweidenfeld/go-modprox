package proxy

import (
	"net/http"
	"io"
	"strings"
	"compress/gzip"
	"bytes"
	"strconv"
	"log"
	"regexp"
	"github.com/PuerkitoBio/goquery"
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
		log.Printf("Processing: %s", url)

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

	c = p.sslRewrite(c)

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
		switch strings.ToLower(h) {
		case "content-length":
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

func (p Proxy) sslRewrite(c string) string {
	for _, rw := range p.Config.SSLRewrite {
		c = strings.Replace(c, fmt.Sprintf("https://%s", rw), fmt.Sprintf("http://%s", rw), -1)
	}
	return c
}

// modify manipulates the document based on the configuration
func (p Proxy) modify(d *goquery.Document, url string) {
	for _, mod := range p.Config.Modifications {
		if m, err := regexp.MatchString(mod.URLMatch, url); nil == err && !m {
			continue;
		}
		i := mod.Index
		e := d.Find(mod.Selector)
		if i >= e.Length() {
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
			tVal, tErr := trim(v)
			if nil != tErr {
				log.Printf("Cannot trim value '%s' (%s)", v, url)
			} else {
				v = tVal
			}
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