package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
)

type (
	jreq struct {
		Template  string            `json:"template,omitempty"`
		Templates map[string]string `json:"templates,omitempty"`
		Data      interface{}       `json:"data,omitempty"`
	}
	jresp struct {
		Status  int         `json:"status"`
		Results interface{} `json:"results,omitempty"`
		Error   string      `json:"error,omitempty"`
	}
)

func renderServer(w http.ResponseWriter, r *http.Request) {
	l.Noticef("new request ( %s %s ) from %s", r.Method, r.URL, r.RemoteAddr)
	if *debug {
		b, err := httputil.DumpRequest(r, true)
		if err != nil {
			l.Errorf("request ( %s %s ) from %s: error dumping request: %s", r.Method, r.URL, r.RemoteAddr, err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		l.Debugf("request ( %s %s ) from %s: %s", r.Method, r.URL, r.RemoteAddr, string(b))
	}
	if r.Method != http.MethodPost {
		l.Errorf("rejected request ( %s %s ) from %s: bad method", r.Method, r.URL, r.RemoteAddr)
		bye(w, r)
		return
	}
	post := jreq{
		//		Template:  `hello {{.client}}, it's {{timestamp}}`,
		//		Data:      struct{ client string }{r.RemoteAddr},
		Templates: make(map[string]string),
	}
	multipart := strings.Contains(r.Header.Get("content-type"), "multipart")
	var err error
	if multipart {
		mr, err := r.MultipartReader()
		if err != nil {
			l.Errorf("request ( %s %s ) from %s: error in request.MultipartReader: %s", r.Method, r.URL, r.RemoteAddr, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			switch pname := part.FormName(); pname {
			case "template":
				buf := new(bytes.Buffer)
				_, err = io.Copy(buf, part)
				if err != nil {
					l.Errorf("request ( %s %s ) from %s: error reading part %s request.MultipartReader: %s", r.Method, r.URL, r.RemoteAddr, pname, err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				post.Template = buf.String()
			case "data":
				buf := new(bytes.Buffer)
				_, err = io.Copy(buf, part)
				if err != nil {
					l.Errorf("request ( %s %s ) from %s: error reading part %s request.MultipartReader: %s", r.Method, r.URL, r.RemoteAddr, pname, err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				d := buf.Bytes()
				err = json.Unmarshal(d, &post.Data)
				if err != nil {
					post.Data = string(d)
				}
			case "templates":
				buf := new(bytes.Buffer)
				_, err = io.Copy(buf, part)
				if err != nil {
					l.Errorf("request ( %s %s ) from %s: error reading part %s request.MultipartReader: %s", r.Method, r.URL, r.RemoteAddr, pname, err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if buf.Len() > 0 {
					post.Templates[part.FileName()] = buf.String()
				}

			}
		}
	} else {
		if err = json.NewDecoder(r.Body).Decode(&post); err != nil {
			l.Warningf("error processing request ( %s %s ) from %s: json.Decode: %s", r.Method, r.URL, r.RemoteAddr, err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(jresp{
				Status: http.StatusBadRequest,
				Error:  err.Error(),
			})
			return
		}
	}
	tpl, err := build(*unsafe, "post", post.Template, post.Templates)
	if err != nil {
		l.Errorf("request ( %s %s ) from %s: error building template.Template: %s", r.Method, r.URL, r.RemoteAddr, err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(jresp{
			Status: http.StatusBadRequest,
			Error:  err.Error(),
		})
		return
	}
	j, _ := json.MarshalIndent(post, "", "  ")
	fmt.Println(string(j))
	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, post.Data); err != nil {
		l.Warningf("error processing request ( %s %s ) from %s: tpl.Execute: %s", r.Method, r.URL, r.RemoteAddr, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(jresp{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		})
		return
	}
	s := buf.String()
	fmt.Println("have", s)
	w.WriteHeader(http.StatusOK)
	if (buf.Len() < (1 << 20)) && multipart {
		_, err = fmt.Fprintf(w, "%s\n%s", htmlHead, htmlArticle(buf.String()))
	} else {
		_, err = io.Copy(w, buf)
	}
	if err != nil {
		l.Errorf("error sending response to request ( %s %s ) from %s: %s", r.Method, r.URL, r.RemoteAddr, err)
	} else {
		l.Infof("processed request ( %s %s ) from %s", r.Method, r.URL, r.RemoteAddr)
	}
}

func uiPage(w http.ResponseWriter, r *http.Request) {
	l.Noticef("new request ( %s %s ) from %s", r.Method, r.URL, r.RemoteAddr)
	if *debug {
		b, _ := httputil.DumpRequest(r, true)
		l.Debugf("request ( %s %s ) from %s: %s", r.Method, r.URL, r.RemoteAddr, string(b))
	}
	if r.Method != http.MethodGet {
		l.Errorf("rejected request ( %s %s ) from %s: bad method", r.Method, r.URL, r.RemoteAddr)
		bye(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprintf(w, "%s\n%s", htmlHead, htmlForm)
	if err != nil {
		l.Errorf("error writing response to request ( %s %s ) from %s: %s", r.Method, r.URL, r.RemoteAddr, err)
	}
}

const (
	htmlHead = `
<!DOCTYPE html>
<meta name="viewport" charset="utf-8" content="width=device-width, initial-scale=1">
<meta http-equiv="Cache-Control" content="no-cache, no-store, must-revalidate" />
<meta http-equiv="Pragma" content="no-cache" />
<meta http-equiv="Expires" content="0" />
<style type="text/css">
Body {
	display:flex;
	font-family: Calibri, Helvetica, sans-serif;
}
.container {
	margin:auto;
    position: absolute;
    top: 50%;
    left: 50%;
	background: #eee;
    -moz-transform: translateX(-50%) translateY(-50%);
    -webkit-transform: translateX(-50%) translateY(-50%);
    transform: translateX(-50%) translateY(-50%);
	border-radius: 3px;
	border: 5px solid;
	box-shadow: 0 1px 2px rgba(0, 0, 0, .1);
	display: inline-block;
	text-align: center;
}
textarea {
	display: inline-block;
	min-width: 30em;
	min-height: 20em;
	overflow: auto;
	resize: both;
}
#formItem label {
    display: inline-block;
	margin:auto;
	position:absolute;
}
p { white-space: pre-line; }
</style>
<title>xinput render</title>
`
	htmlForm = `
<body>
<div class="container">
<form method="post" action="render" enctype="multipart/form-data">
    <p>
		<label for="text">TEMPLATE</label>
		<textarea class="text" name="template" id="template">hello {{.client}}, it's {{timestamp}}</textarea>
	</p>
	<p>
		<label for="text">DATA</label>
		<textarea class="text" name="data" id="data">{"client": "meeee"}</textarea>
	</p>
	<p>
		<input type="file" id="templates" name="templates" accept="text/*" multiple>
		<input type="submit" class="submit" value="Submit" />
		<input type="reset" value="Reset">
	</p>
</form>
</div>
</body>
</html>
`
)

func htmlArticle(text string) string {
	if text == "" {
		text = "aooo"
	}
	return fmt.Sprintf(
		"\n%s%s%s\n%s\n%s",
		`<div class="container">
<article class="all-browsers">
<pre><code>`,
		text,
		`</code></pre>
</article>`,
		`<a href="javascript:history.back()">back</a>`,
		`</div>
</body>
</html>`,
	)
}

func bye(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://www.youtube.com/watch?v=dQw4w9WgXcQ?autoplay=1", http.StatusPermanentRedirect)
}
