package main

/*
import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"text/template"
)

type jresp struct {
	Status  int         `json:"status"`
	Results interface{} `json:"results,omitempty"`
	Error   error       `json:"error,omitempty"`
}

func renderServer(w http.ResponseWriter, r *http.Request) {
	l.Noticef("new request ( %s %s ) from %s", r.Method, r.URL, r.RemoteAddr)
	if *debug {
		b, _ := httputil.DumpRequest(r, true)
		l.Debugf("request ( %s %s ) from %s: %s", r.Method, r.URL, r.RemoteAddr, string(b))
	}
	if r.Method != http.MethodPost {
		l.Errorf("rejected request ( %s %s ) from %s: bad method", r.Method, r.URL, r.RemoteAddr)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		l.Errorf("error processing request ( %s %s ) from %s: invalid form data: %s", r.Method, r.URL, r.RemoteAddr, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	j, _ := json.MarshalIndent(r.Form, "", "  ")
	fmt.Println(string(j))
	var postTemplates map[string]string
	if err := json.NewDecoder(r.Body).Decode(&postTemplates); err != nil {
		l.Errorf("error processing request ( %s %s ) from %s: invalid body: %s", r.Method, r.URL, r.RemoteAddr, err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(jresp{
			Status: http.StatusBadRequest,
			Error: fmt.Errorf(
				`must submit a map[key]value JSON object where key is a string identifier for the template and value is the base64-encoded template itself. a special map["data"] field (also base64 encoded) may be provided, it will not be considered a template but as a data object to apply to the templates in template.Execute`,
			),
		})
		return
	}
	var postedData interface{}
	if v, ok := postTemplates["data"]; ok {
		b, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			l.Warningf("error decoding map[data] from request ( %s %s ) from %s: b64dec: %s", r.Method, r.URL, r.RemoteAddr, err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(jresp{
				Status: http.StatusBadRequest,
				Error:  err,
			})
			return
		}
		postedData = string(b)
		delete(postTemplates, "data")
	}
	if len(postTemplates) < 1 {
		l.Infof("request ( %s %s ) from %s: no content", r.Method, r.URL, r.RemoteAddr)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	tpl := template.New("").Funcs(buildFuncMap(*unsafe))
	for n, t := range postTemplates {
		b, err := base64.RawStdEncoding.DecodeString(t)
		if err != nil {
			l.Warningf("error decoding templates from request ( %s %s ) from %s: b64dec: %s", r.Method, r.URL, r.RemoteAddr, err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(jresp{
				Status: http.StatusBadRequest,
				Error:  err,
			})
			return
		}
		tpl, err = tpl.New(n).Parse(string(b))
		if err != nil {
			l.Warningf("error processing request ( %s %s ) from %s: tpl.New: %s", r.Method, r.URL, r.RemoteAddr, err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(jresp{
				Status: http.StatusBadRequest,
				Error:  err,
			})
			return
		}
	}
	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, postedData); err != nil {
		l.Warningf("error processing request ( %s %s ) from %s: tpl.Execute: %s", r.Method, r.URL, r.RemoteAddr, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(jresp{
			Status: http.StatusInternalServerError,
			Error:  err,
		})
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err := io.Copy(w, buf)
	if err != nil {
		l.Warningf("error sending response to request ( %s %s ) from %s: %s", r.Method, r.URL, r.RemoteAddr, err)
	}
	l.Infof("processed request ( %s %s ) from %s", r.Method, r.URL, r.RemoteAddr)
}

//http.Error()

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
Body {
 display:flex;
 font-family: Calibri, Helvetica, sans-serif;
 background-color: #303030;
}
h1 {text-align: center;}
p {text-align: center;}
div {text-align: center;}
</style>
<html>
    <head>
    <title>xprint render page</title>
    </head>
    <body>
      <form enctype="multipart/form-data" action="/render" method="POST">
		<textarea id="template">main template</textarea>
		<textarea id="data">data</textarea>
		<input type="file" id="templates" name="templates" multiple>
		<input type="submit">render</button>
	  </form>
    </body>
</html>
`))
}
*/
