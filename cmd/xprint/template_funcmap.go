// Copyright (c) 2019-2021 SILVANO ZAMPARDI, All rights reserved.
// This source code license can be found in the LICENSE file in the root directory of this source tree.

package main

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	log "github.com/szampardi/msg"
	"golang.org/x/term"
	"gopkg.in/yaml.v2"
)

type fn struct {
	_fn         interface{} `json:"-"`
	Description string      `json:"description"`
	Signature   string      `json:"function"`
	Unsafe      bool        `json:"unsafe"`
}

/* src/text/template/funcs.go
func builtins() FuncMap {
	return FuncMap{
		"and":      and,
		"call":     call,
		"html":     HTMLEscaper,
		"index":    index,
		"slice":    slice,
		"js":       JSEscaper,
		"len":      length,
		"not":      not,
		"or":       or,
		"print":    fmt.Sprint,
		"printf":   fmt.Sprintf,
		"println":  fmt.Sprintln,
		"urlquery": URLQueryEscaper,

		// Comparisons
		"eq": eq, // ==
		"ge": ge, // >=
		"gt": gt, // >
		"le": le, // <=
		"lt": lt, // <
		"ne": ne, // !=
	}
}
*/

var templateFnsInfo = map[string]fn{
	"add": {
		add,
		"add value $2 to map or slice $1, map needs $3 for value's key in map",
		reflect.TypeOf(add).String(),
		false,
	},
	"b64dec": {
		b64dec,
		"base64 decode",
		reflect.TypeOf(b64dec).String(),
		false,
	},
	"b64enc": {
		b64enc,
		"base64 encode",
		reflect.TypeOf(b64enc).String(),
		false,
	},
	"cmd": {
		cmd,
		"execute a command on local host",
		reflect.TypeOf(cmd).String(),
		true,
	},
	"decrypt": {
		decrypt,
		"decrypt data with AES_GCM: $1 ctxt, $2 base64 key, $3 AAD",
		reflect.TypeOf(decrypt).String(),
		false,
	},
	"encrypt": {
		encrypt,
		"encrypt data with AES_GCM: $1 ptxt, $2 base64 key, $3 AAD",
		reflect.TypeOf(encrypt).String(),
		false,
	},
	"env": {
		env,
		"get environment vars, optionally use a placeholder value $2",
		reflect.TypeOf(env).String(),
		true,
	},
	"fromgob": {
		fromgob,
		"gob decode",
		reflect.TypeOf(fromgob).String(),
		false,
	},
	"fromjson": {
		fromjson,
		"json decode",
		reflect.TypeOf(fromjson).String(),
		false,
	},
	"fromyaml": {
		fromyaml,
		"yaml decode",
		reflect.TypeOf(fromyaml).String(),
		false,
	},
	"gunzip": {
		_gunzip,
		"extract GZIP compressed data",
		reflect.TypeOf(_gunzip).String(),
		false,
	},
	"gzip": {
		_gzip,
		"compress with GZIP",
		reflect.TypeOf(_gzip).String(),
		false,
	},
	"hexdec": {
		hexdec,
		"hex decode",
		reflect.TypeOf(hexdec).String(),
		false,
	},
	"hexenc": {
		hexenc,
		"hex encode",
		reflect.TypeOf(hexenc).String(),
		false,
	},
	"http": {
		_http,
		"HEAD|GET|POST, url, body(raw), headers",
		reflect.TypeOf(_http).String(),
		true,
	},
	"join": {
		strings.Join,
		"strings.Join",
		reflect.TypeOf(strings.Join).String(),
		false,
	},
	"lower": {
		strings.ToLower,
		"strings.ToLower",
		reflect.TypeOf(strings.ToLower).String(),
		false,
	},
	"pathbase": {
		filepath.Base,
		"filepath.Base",
		reflect.TypeOf(filepath.Base).String(),
		false,
	},
	"pathext": {
		filepath.Ext,
		"filepath.Ext",
		reflect.TypeOf(filepath.Ext).String(),
		false,
	},
	"random": {
		random,
		"generate a $1 sized []byte filled with bytes from crypto.Rand",
		reflect.TypeOf(random).String(),
		false,
	},
	"rawfile": {
		rawfile,
		"read raw bytes from a file",
		reflect.TypeOf(rawfile).String(),
		true,
	},
	"split": {
		strings.Split,
		"strings.Split",
		reflect.TypeOf(strings.Split).String(),
		false,
	},
	"string": {
		stringify,
		"convert int/bool to string, retype []byte to string (handle with care)",
		reflect.TypeOf(stringify).String(),
		false,
	},
	"textfile": {
		textfile,
		"read a file as a string",
		reflect.TypeOf(textfile).String(),
		true,
	},
	"togob": {
		togob,
		"gob encode",
		reflect.TypeOf(togob).String(),
		false,
	},
	"tojson": {
		tojson,
		"json encode",
		reflect.TypeOf(tojson).String(),
		false,
	},
	"toyaml": {
		toyaml,
		"yaml encode",
		reflect.TypeOf(toyaml).String(),
		false,
	},
	"trimprefix": {
		strings.TrimPrefix,
		"strings.TrimPrefix",
		reflect.TypeOf(strings.TrimPrefix).String(),
		false,
	},
	"trimsuffix": {
		strings.TrimSuffix,
		"strings.TrimSuffix",
		reflect.TypeOf(strings.TrimSuffix).String(),
		false,
	},
	"upper": {
		strings.ToUpper,
		"strings.ToUpper",
		reflect.TypeOf(strings.ToUpper).String(),
		false,
	},
	"userinput": {
		userinput,
		"get interactive user input (needs a terminal), if $2 bool is provided and true, term.ReadPassword is used. $1 is used as hint",
		reflect.TypeOf(userinput).String(),
		true,
	},
	"writefile": {
		writefile,
		"store data to a file (append if it already exists)",
		reflect.TypeOf(writefile).String(),
		true,
	},
}

func buildFuncMap(addUnsafe bool) template.FuncMap {
	m := make(template.FuncMap)
	for name, info := range templateFnsInfo {
		if !info.Unsafe || addUnsafe {
			m[name] = info._fn
		}
	}
	return m
}

type fnTrack struct {
	T      time.Time     `json:"time"`
	F      string        `json:"function"`
	Args   []interface{} `json:"args,omitempty"`
	Output interface{}   `json:"output,omitempty"`
	Err    error         `json:"error,omitempty"`
}

var (
	fnTrackChan chan (*fnTrack) = make(chan *fnTrack)
	trackWg     *sync.WaitGroup = &sync.WaitGroup{}
)

func trackUsage(_fn string, output interface{}, err error, args ...interface{}) {
	if *debug {
		trackWg.Add(1)
		fnTrackChan <- &fnTrack{
			T:      time.Now(),
			F:      _fn,
			Args:   args[:],
			Output: output,
			Err:    err,
		}
	}
}

func usageDebugger() {
	go func() {
		log.SetOutput(os.Stderr)
		for x := range fnTrackChan {
			j, _ := json.Marshal(x)
			log.Warningf(string(j))
			trackWg.Done()
		}
	}()
}

func _http(method, url string, body interface{}, headers map[string]string) (out *http.Response, err error) {
	method = strings.ToUpper(method)
	defer trackUsage("http", out, err, method, url, headers, body)
	var bodyr io.Reader
	switch t := body.(type) {
	case string:
		bodyr = bytes.NewBuffer([]byte(t))
	case []byte:
		bodyr = bytes.NewBuffer(t)
	case io.Reader:
		bodyr = t
	default:
		err = fmt.Errorf("invalid argument %T, supported types: io.Reader, string or []byte", t)
	}
	var req *http.Request
	req, err = http.NewRequest(method, url, bodyr)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	out, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func userinput(title string, hidden ...bool) (out string, err error) {
	defer trackUsage("userinput", &out, err, title, hidden[:])
	h := false
	if len(hidden) > 0 {
		h = hidden[0]
	}
	if h {
		log.Noticef("(%s) input secret now, followed by newline", title)
		b, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return "", err
		}
		out = string(b)
	} else {
		log.Noticef("(%s) reading input, CTRL-D to stop", title)
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		out = strings.TrimSpace(string(b))
	}
	return out, err
}

func togob(in interface{}) (out []byte, err error) {
	defer trackUsage("togob", &out, err, in)
	buf := new(bytes.Buffer)
	if err = gob.NewEncoder(buf).Encode(in); err != nil {
		return nil, err
	}
	out = buf.Bytes()
	return out, nil
}

func fromgob(in interface{}) (out interface{}, err error) {
	defer trackUsage("fromgob", &out, err, in)
	var todo io.Reader
	switch t := in.(type) {
	case string:
		todo = bytes.NewBuffer([]byte(t))
	case []byte:
		todo = bytes.NewBuffer(t)
	case io.Reader:
		todo = t
	}
	if err = gob.NewDecoder(todo).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func tojson(in interface{}) (out string, err error) {
	defer trackUsage("tojson", &out, err, in)
	b, err := json.Marshal(in)
	if err != nil {
		return "", err
	}
	out = string(b)
	return out, nil
}

func fromjson(in interface{}) (out interface{}, err error) {
	defer trackUsage("fromjson", &out, err, in)
	switch t := in.(type) {
	case string:
		if err := json.Unmarshal([]byte(t), &out); err != nil {
			return nil, err
		}
	case []byte:
		if err := json.Unmarshal(t, &out); err != nil {
			return nil, err
		}
	case io.Reader:
		if err = json.NewDecoder(t).Decode(&out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func toyaml(in interface{}) (out string, err error) {
	defer trackUsage("toyaml", &out, err, in)
	b, err := yaml.Marshal(in)
	if err != nil {
		return "", err
	}
	out = string(b)
	return out, nil
}

func fromyaml(in interface{}) (out interface{}, err error) {
	defer trackUsage("fromyaml", &out, err, in)
	switch t := in.(type) {
	case string:
		if err := yaml.Unmarshal([]byte(t), &out); err != nil {
			return nil, err
		}
	case []byte:
		if err := yaml.Unmarshal(t, &out); err != nil {
			return nil, err
		}
	case io.Reader:
		if err = yaml.NewDecoder(t).Decode(&out); err != nil {
			return nil, err
		}
	default:
		err = fmt.Errorf("invalid argument %T, supported types: io.Reader, string or []byte", t)
		return nil, err
	}
	return out, nil
}

func b64enc(in interface{}) (out string, err error) {
	defer trackUsage("b64enc", &out, err, in)
	var b []byte
	switch t := in.(type) {
	case string:
		b = make([]byte, base64.StdEncoding.EncodedLen(len(t)))
		base64.StdEncoding.Encode(b, []byte(t))
		out = string(b)
	case []byte:
		b = make([]byte, base64.StdEncoding.EncodedLen(len(t)))
		base64.StdEncoding.Encode(b, t)
		out = string(b)
	case io.Reader:
		buf := new(bytes.Buffer)
		_, err = io.Copy(base64.NewEncoder(base64.RawStdEncoding, buf), t)
		if err != nil {
			return "", err
		}
		out = buf.String()
	default:
		err = fmt.Errorf("invalid argument %T, supported types: io.Reader, string or []byte", t)
		return "", err
	}
	return out, nil
}

func b64dec(in interface{}) (out []byte, err error) {
	defer trackUsage("b64dec", &out, err, in)
	var b []byte
	var n int
	switch t := in.(type) {
	case string:
		b = make([]byte, base64.StdEncoding.DecodedLen(len(t)))
		n, err = base64.StdEncoding.Decode(b, []byte(t))
		if err != nil {
			return nil, err
		}
		out = b[:n]
	case []byte:
		b = make([]byte, base64.StdEncoding.DecodedLen(len(t)))
		n, err = base64.StdEncoding.Decode(b, t)
		if err != nil {
			return nil, err
		}
		out = b[:n]
	case io.Reader:
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, base64.NewDecoder(base64.RawStdEncoding, t))
		if err != nil {
			return nil, err
		}
		out = buf.Bytes()
	default:
		err = fmt.Errorf("invalid argument %T, supported types: io.Reader, string or []byte", t)
		return nil, err
	}
	return out, nil
}

func hexenc(in interface{}) (out string, err error) {
	defer trackUsage("hexenc", &out, err, in)
	switch t := in.(type) {
	case string:
		out = hex.EncodeToString([]byte(t))
	case []byte:
		out = hex.EncodeToString(t)
	case io.Reader:
		buf := new(bytes.Buffer)
		_, err = io.Copy(hex.NewEncoder(buf), t)
		if err != nil {
			return "", err
		}
		out = buf.String()
	default:
		err = fmt.Errorf("invalid argument %T, supported types: string or []byte", t)
		return "", err
	}
	return out, nil
}

func hexdec(in interface{}) (out []byte, err error) {
	defer trackUsage("hexdec", &out, err, in)
	var b []byte
	var n int
	switch t := in.(type) {
	case string:
		b, err = hex.DecodeString(t)
		if err != nil {
			return nil, err
		}
		out = b[:n]
	case []byte:
		b = make([]byte, hex.DecodedLen(len(t)))
		n, err = hex.Decode(b, t)
		if err != nil {
			return nil, err
		}
		out = b[:n]
	case io.Reader:
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, hex.NewDecoder(t))
		if err != nil {
			return nil, err
		}
		out = buf.Bytes()
	default:
		err = fmt.Errorf("invalid argument %T, supported types: string or []byte", t)
		return nil, err
	}
	return out, nil
}

func _gzip(in interface{}) (out []byte, err error) {
	defer trackUsage("gzip", &out, err, in)
	var todo io.Reader
	switch t := in.(type) {
	case string:
		todo = bytes.NewBuffer([]byte(t))
	case []byte:
		todo = bytes.NewBuffer(t)
	case io.Reader:
		todo = t
	default:
		err = fmt.Errorf("invalid argument %T, supported types: io.Reader, string or []byte", t)
		return nil, err
	}
	buf := new(bytes.Buffer)
	gzw, err := gzip.NewWriterLevel(buf, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(gzw, todo)
	if err != nil {
		return nil, err
	}
	if err = gzw.Flush(); err != nil {
		return nil, err
	}
	if err = gzw.Close(); err != nil {
		return nil, err
	}
	out = buf.Bytes()
	return out, nil
}

func _gunzip(in interface{}) (out []byte, err error) {
	defer trackUsage("gunzip", &out, err, in)
	var todo io.Reader
	switch t := in.(type) {
	case string:
		// try to go on assuming its base64-encoded
		b, err := b64dec(t)
		if err != nil {
			return nil, err
		}
		todo = bytes.NewBuffer(b)
	case []byte:
		todo = bytes.NewBuffer(t)
	case io.Reader:
		todo = t
	default:
		err = fmt.Errorf("invalid argument %T, supported types: io.Reader or []byte", t)
		return nil, err
	}
	gzr, err := gzip.NewReader(todo)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, gzr); err != nil {
		return nil, err
	}
	if err = gzr.Close(); err != nil {
		return nil, err
	}
	out = buf.Bytes()
	return out, nil
}

func rawfile(in string) (out []byte, err error) {
	defer trackUsage("rawfile", &out, err, in)
	f, err := os.Open(in)
	if err != nil {
		return nil, err
	}
	out, err = ioutil.ReadAll(f)
	return out, err
}

func textfile(in string) (out string, err error) {
	defer trackUsage("textfile", &out, err, in)
	f, err := os.Open(in)
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	out = string(data)
	return out, nil
}

func writefile(in interface{}, fpath string) (out string, err error) {
	defer trackUsage("writefile", "", err, in, fpath)
	f, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.FileMode(0600))
	if err != nil {
		return "", err
	}
	var todo io.Reader
	switch t := in.(type) {
	case string:
		todo = bytes.NewBuffer([]byte(t))
	case []byte:
		todo = bytes.NewBuffer(t)
	case io.Reader:
		todo = t
	default:
		err = fmt.Errorf("invalid argument %T, supported types: io.Reader, string or []byte", t)
		return "", err
	}
	_, err = io.Copy(f, todo)
	return "", err
}

func stringify(in interface{}) (out string, err error) {
	defer trackUsage("string", &out, err, in)
	switch t := in.(type) {
	case string:
		out = t
	case int:
		out = strconv.Itoa(t)
	case bool:
		out = strconv.FormatBool(t)
	case []byte:
		out = string(t)
	default:
		err = fmt.Errorf("invalid argument %T, supported types: int, bool and []byte", t)
		return "", err
	}
	return out, nil
}

func encrypt(in interface{}, b64key string, aad string) (out []byte, err error) {
	defer trackUsage("encrypt", &out, err, in, b64key, aad)
	var ptxt []byte
	switch t := in.(type) {
	case string:
		ptxt = []byte(t)
	case []byte:
		ptxt = t
	case io.Reader:
		ptxt, err = consumeReader(t)
		if err != nil {
			return nil, err
		}
	default:
		err = fmt.Errorf("invalid argument %T, supported types: string or []byte", t)
		return nil, err
	}
	key, err := base64.StdEncoding.DecodeString(b64key)
	if err != nil {
		return nil, err
	}
	cb, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(cb)
	if err != nil {
		return nil, err
	}
	iv := random(aead.NonceSize())
	var ctxt []byte
	ctxt = aead.Seal(ctxt, iv, ptxt, []byte(aad))
	out = append(ctxt, iv...)
	return out, nil
}

func decrypt(in interface{}, b64key string, aad string) (out []byte, err error) {
	defer trackUsage("decrypt", &out, err, in, b64key, aad)
	var ctxt []byte
	switch t := in.(type) {
	case string: // try to go on assuming its base64-encoded
		ctxt, err = b64dec(t)
		if err != nil {
			return nil, err
		}
	case []byte:
		ctxt = t
	case io.Reader:
		ctxt, err = consumeReader(t)
		if err != nil {
			return nil, err
		}
	}
	key, err := base64.StdEncoding.DecodeString(b64key)
	if err != nil {
		return nil, err
	}
	cb, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(cb)
	if err != nil {
		return nil, err
	}
	ns := aead.NonceSize()
	l := len(ctxt)
	if l < ns {
		return nil, io.ErrUnexpectedEOF
	}
	out, err = aead.Open(out, ctxt[l-ns:], ctxt[:l-ns], []byte(aad))
	if err != nil {
		return nil, err
	}
	return out, nil
}

func add(in interface{}, value interface{}, key ...interface{}) (out interface{}, err error) {
	defer trackUsage("add", &out, err, in, value, key)
	switch t := in.(type) {
	case map[int]interface{}:
		if len(key) < 1 {
			return nil, fmt.Errorf("must provide a key for value to be added")
		}
		switch tk := key[0].(type) {
		case int:
			t[tk] = value
			out = t
			return t, nil
		default:
			err = fmt.Errorf("key for value to be added must be an int, not %T", tk)
			return nil, err
		}
	case map[string]interface{}:
		if len(key) < 1 {
			return nil, fmt.Errorf("must provide a key for value to be added")
		}
		switch tk := key[0].(type) {
		case string:
			t[tk] = value
			out = t
			return t, nil
		default:
			err = fmt.Errorf("key for value to be added must be a string, not %T", tk)
			return nil, err
		}
	case map[interface{}]interface{}:
		if len(key) < 1 {
			err = fmt.Errorf("must provide a key for value to be added")
			return nil, err
		}
		// should be safe?
		t[key[0]] = value
		out = t
		return t, nil
	case []interface{}:
		t = append(t, value)
		out = t
		return out, nil
	default:
		err = fmt.Errorf("invalid argument %T, supported types: slices and maps", t)
		return nil, err
	}
}

func env(in string, or ...string) (out string) {
	defer trackUsage("env", &out, nil, in, or)
	if v, ok := os.LookupEnv(in); ok {
		return v
	}
	if len(or) > 0 {
		return or[0]
	}
	return ""
}

func cmd(prog string, args ...string) (out string, err error) {
	defer trackUsage("cmd", &out, err, prog, args[:])
	x := exec.Command(prog, args...)
	outbuf, errbuf := new(bytes.Buffer), new(bytes.Buffer)
	x.Stderr = errbuf
	x.Stdout = outbuf
	err = x.Run()
	if err != nil {
		return "", err
	}
	if errbuf.Len() > 0 {
		err = fmt.Errorf("%s error: %s", prog, errbuf.String())
	}
	out = outbuf.String()
	return out, err
}

func random(size int) (out []byte) {
	defer trackUsage("random", &out, nil, size)
	out = make([]byte, size)
	_, err := io.ReadFull(rand.Reader, out)
	if err != nil {
		panic(err)
	}
	return out[:size]
}

func consumeReader(r io.Reader) ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, r)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
