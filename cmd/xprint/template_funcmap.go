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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

var tplFuncMap template.FuncMap = template.FuncMap{
	"tojson": func(in interface{}) (string, error) {
		b, err := json.Marshal(in)
		if err != nil {
			return "", err
		}
		return string(b), nil
	},
	"fromjson": func(in string) (map[string]interface{}, error) {
		var out map[string]interface{}
		if err := json.Unmarshal([]byte(in), &out); err != nil {
			return nil, err
		}
		return out, nil
	},
	"toyaml": func(in interface{}) (string, error) {
		b, err := yaml.Marshal(in)
		if err != nil {
			return "", err
		}
		return string(b), nil
	},
	"fromyaml": func(in string) (map[string]interface{}, error) {
		var out map[string]interface{}
		if err := yaml.Unmarshal([]byte(in), &out); err != nil {
			return nil, err
		}
		return out, nil
	},
	"b64enc": func(in interface{}) (string, error) {
		var b []byte
		switch t := in.(type) {
		case string:
			b = make([]byte, base64.StdEncoding.EncodedLen(len(t)))
			base64.StdEncoding.Encode(b, []byte(t))
		case []byte:
			b = make([]byte, base64.StdEncoding.EncodedLen(len(t)))
			base64.StdEncoding.Encode(b, t)
		default:
			return "", fmt.Errorf("can only work with string or []byte, not %T", t)
		}
		return string(b), nil
	},
	"b64dec": func(in interface{}) ([]byte, error) {
		var b []byte
		var n int
		var err error
		switch t := in.(type) {
		case string:
			b = make([]byte, base64.StdEncoding.DecodedLen(len(t)))
			n, err = base64.StdEncoding.Decode(b, []byte(t))
		case []byte:
			b = make([]byte, base64.StdEncoding.DecodedLen(len(t)))
			n, err = base64.StdEncoding.Decode(b, t)
		default:
			return nil, fmt.Errorf("can only work with string or []byte, not %T", t)
		}
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	},
	"hexenc": func(in interface{}) (string, error) {
		switch t := in.(type) {
		case string:
			return hex.EncodeToString([]byte(t)), nil
		case []byte:
			return hex.EncodeToString(t), nil
		default:
			return "", fmt.Errorf("can only work with string or []byte, not %T", t)
		}
	},
	"hexdec": func(in interface{}) ([]byte, error) {
		var b []byte
		var n int
		var err error
		switch t := in.(type) {
		case string:
			b, err = hex.DecodeString(t)
		case []byte:
			b = make([]byte, hex.DecodedLen(len(t)))
			n, err = hex.Decode(b, t)
		default:
			return nil, fmt.Errorf("can only work with string or []byte, not %T", t)
		}
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	},
	"gzip": func(in interface{}) ([]byte, error) {
		var todo []byte
		switch t := in.(type) {
		case string:
			todo = []byte(t)
		case []byte:
			todo = t
		}
		out := new(bytes.Buffer)
		gzw, err := gzip.NewWriterLevel(out, gzip.BestCompression)
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(gzw, bytes.NewBuffer(todo))
		if err != nil {
			return nil, err
		}
		if err = gzw.Flush(); err != nil {
			return nil, err
		}
		if err = gzw.Close(); err != nil {
			return nil, err
		}
		return out.Bytes(), nil
	},
	"gunzip": func(in []byte) ([]byte, error) {
		gzr, err := gzip.NewReader(bytes.NewBuffer(in))
		if err != nil {
			return nil, err
		}
		out := new(bytes.Buffer)
		_, err = io.Copy(out, gzr)
		if err != nil {
			return nil, err
		}
		if err = gzr.Close(); err != nil {
			return nil, err
		}
		return out.Bytes(), nil
	},
	"rawfile": func(in string) ([]byte, error) {
		f, err := os.Open(in)
		if err != nil {
			return nil, err
		}
		return ioutil.ReadAll(f)
	},
	"textfile": func(in string) (string, error) {
		f, err := os.Open(in)
		if err != nil {
			return "", err
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			return "", err
		}
		return string(data), nil
	},
	"writefile": func(in interface{}, fpath string) (string, error) {
		f, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY, os.FileMode(0600))
		if err != nil {
			return "", err
		}
		var buf *bytes.Buffer
		switch t := in.(type) {
		case string:
			buf = bytes.NewBuffer([]byte(t))
		case []byte:
			buf = bytes.NewBuffer(t)
		default:
			return "", fmt.Errorf("can only work with string or []byte, not %T", t)
		}
		_, err = io.Copy(f, buf)
		return f.Name(), err
	},
	"string": func(in interface{}) (string, error) {
		switch t := in.(type) {
		case string:
			return t, nil
		case int:
			return strconv.Itoa(t), nil
		case bool:
			return strconv.FormatBool(t), nil
		case []byte:
			return string(t), nil
		default:
			return "", fmt.Errorf("string: can only work with int, bool and []byte, not %T", t)
		}
	},
	"encrypt": func(in interface{}, b64key string, aad string) ([]byte, error) {
		var ptxt []byte
		switch t := in.(type) {
		case string:
			ptxt = []byte(t)
		case []byte:
			ptxt = t
		default:
			return nil, fmt.Errorf("can only work with string or []byte, not %T", t)
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
		return append(ctxt, iv...), nil
	},
	"decrypt": func(ctxt []byte, b64key string, aad string) ([]byte, error) {
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
		var ptxt []byte
		ptxt, err = aead.Open(ptxt, ctxt[l-ns:], ctxt[:l-ns], []byte(aad))
		if err != nil {
			return nil, err
		}
		return ptxt, nil
	},
	"env":        env,
	"cmd":        cmd,
	"random":     random,
	"join":       strings.Join,
	"split":      strings.Split,
	"trimprefix": strings.TrimPrefix,
	"trimsuffix": strings.TrimSuffix,
	"lower":      strings.ToLower,
	"upper":      strings.ToUpper,
	"pathbase":   filepath.Base,
	"pathext":    filepath.Ext,
}

func env(in string, or ...string) string {
	if !*unsafe {
		panic(*unsafe)
	}
	if v, ok := os.LookupEnv(in); ok {
		return v
	}
	if len(or) > 0 {
		return or[0]
	}
	return ""
}

func cmd(prog string, args ...string) (string, error) {
	if !*unsafe {
		panic(*unsafe)
	}
	x := exec.Command(prog, args...)
	outbuf, errbuf := new(bytes.Buffer), new(bytes.Buffer)
	x.Stderr = errbuf
	x.Stdout = outbuf
	err := x.Run()
	if err != nil {
		return "", err
	}
	if errbuf.Len() > 0 {
		err = fmt.Errorf("%s error: %s", prog, errbuf.String())
	}
	return outbuf.String(), err
}

func random(size int) []byte {
	buf := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		panic(err)
	}
	return buf[:size]
}
