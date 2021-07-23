package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
	"text/template"
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
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
	var post jreq
	var err error
	if strings.Contains(r.Header.Get("content-type"), "multipart") {
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
				err = json.Unmarshal(buf.Bytes(), &post.Data)
				if err != nil {
					l.Errorf("request ( %s %s ) from %s: error reading part %s json.Unmarshal: %s", r.Method, r.URL, r.RemoteAddr, pname, err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			case "templates":
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
	tpl := template.New("post").Funcs(buildFuncMap(*unsafe))
	for n, t := range post.Templates {
		l.Debugf("request ( %s %s ) from %s: parsing post.Templates[%s] = %s", r.Method, r.URL, r.RemoteAddr, n, t)
		tpl, err = tpl.New(n).Parse(t)
		if err != nil {
			l.Warningf("error processing request ( %s %s ) from %s: tpl.New: %s", r.Method, r.URL, r.RemoteAddr, err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(jresp{
				Status: http.StatusBadRequest,
				Error:  err.Error(),
			})
			return
		}
	}
	if post.Template != "" {
		tpl, err = tpl.Parse(post.Template)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(jresp{
				Status: http.StatusBadRequest,
				Error:  err.Error(),
			})
			return
		}
	} else {
		l.Errorf("request ( %s %s ) from %s: empty template", r.Method, r.URL, r.RemoteAddr)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
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
	w.WriteHeader(http.StatusOK)
	if _, err = io.Copy(w, buf); err != nil {
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
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<!DOCTYPE html>
<style>
.container {
    position: absolute;
    top: 50%;
    left: 50%;
    -moz-transform: translateX(-50%) translateY(-50%);
    -webkit-transform: translateX(-50%) translateY(-50%);
    transform: translateX(-50%) translateY(-50%);
}
</style>
<html>
<head>
<meta charset="utf-8">
<title>xinput render</title>
</head>
<body>
<div class="container">
<form method="post" action="render" enctype="multipart/form-data">
    <p>
		<label for="text">Template:</label>
		<textarea class="text" name="template" id="template"></textarea>
	</p>
	<p>
		<label for="text">Data:</label>
		<textarea class="text" name="data" id="data"></textarea>
	</p>
	<p><input type="submit" class="submit" value="Submit" /></p>
</form>
</div>
</body>
</html>
`))
}
