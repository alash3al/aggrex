package vm

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"core.globals"

	"github.com/clbanning/mxj"
	"github.com/go-resty/resty"
	"github.com/robertkrimen/otto"
	_ "github.com/robertkrimen/otto/underscore"
)

// VM .
type VM struct {
	AllowedHosts []string
	MaxExecTime  int64
	Request      *http.Request
	LastError    interface{}
}

// New .
func New(v VM) *VM {
	return &v
}

// Exec .
func (v *VM) Exec(script string) (interface{}, error) {
	defer func() {
		if err := recover(); err != nil {
			v.LastError = err
			log.Println("[Panic]", err)
		}
	}()

	script = `var collect = _; utils.collect = collect;` + script

	vm := otto.New()
	vm.Interrupt = make(chan func(), 1)

	time.AfterFunc(time.Duration(v.MaxExecTime)*time.Second, func() {
		vm.Interrupt <- func() {
			panic(fmt.Sprintf("reached the maximum execution time (%d sec)", v.MaxExecTime))
		}
	})

	queryParams := map[string]string{}
	if v.Request != nil {
		for k, v := range v.Request.URL.Query() {
			if len(v) < 1 {
				continue
			}
			queryParams[k] = v[0]
		}
	}

	headers := map[string]string{}
	if v.Request != nil {
		for k, v := range v.Request.Header {
			if len(v) < 1 {
				continue
			}
			headers[k] = v[0]
		}
	}

	var inBody interface{}

	if v.Request != nil {
		if v.Request.Method == "POST" {
			json.NewDecoder(v.Request.Body).Decode(&inBody)
			if v.Request.Body != nil {
				v.Request.Body.Close()
			}
		}
	}

	vm.Set("exports", map[string]interface{}{})
	vm.Set("fetch", v.funcFetch)
	vm.Set("globals", globals.DBHandler.GlobalsGet())

	if v.Request != nil {
		vm.Set("request", map[string]interface{}{
			"uri":         v.Request.URL.RequestURI(),
			"proto":       v.Request.Proto,
			"host":        v.Request.Host,
			"remote_addr": v.Request.RemoteAddr,
			"query":       queryParams,
			"headers":     headers,
			"body":        inBody,
		})
	} else {
		vm.Set("request", map[string]interface{}{})
	}

	vm.Set("utils", map[string]interface{}{
		"btoa": func(s string) string {
			return base64.StdEncoding.EncodeToString([]byte(s))
		},
		"atob": func(s string) string {
			b, _ := base64.StdEncoding.DecodeString(s)
			return string(b)
		},
		"uniqid": func(l int) string {
			b := make([]byte, l)
			rand.Read(b)
			return fmt.Sprintf("%x", b)
		},
		"md5": func(s string) string {
			return fmt.Sprintf("%x", md5.Sum([]byte(s)))
		},
		"sha256": func(s string) string {
			return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
		},
		"sha512": func(s string) string {
			return fmt.Sprintf("%x", sha512.Sum512([]byte(s)))
		},
		"bcrypt": func(s string) string {
			b, _ := bcrypt.GenerateFromPassword([]byte(s), 9)
			return string(b)
		},
		"bcryptCheck": func(h, s string) bool {
			return bcrypt.CompareHashAndPassword([]byte(h), []byte(s)) == nil
		},
		"fetch": v.funcFetch,
	})

	vm.Set("cron", map[string]interface{}{
		"list":  globals.DBHandler.CronsGet,
		"set":   globals.DBHandler.CronsSet,
		"unset": globals.DBHandler.CronsUnset,
	})

	val, err := vm.Eval(script)
	if err != nil {
		return val, err
	}

	exports, err := vm.Get("exports")
	if err != nil {
		return nil, err
	}

	return exports.Export()
}

// funcFetch .
func (v *VM) funcFetch(args map[string]interface{}) map[string]interface{} {
	target, proxy, method := "", "", ""
	redirCount := 5
	headers := map[string]string{}

	if args["url"] == nil {
		return map[string]interface{}{
			"statusCode": 500,
			"headers":    nil,
			"size":       0,
			"body":       nil,
			"error":      "empty url",
		}
	}

	target = args["url"].(string)

	parsedURL, err := url.Parse(target)
	if err != nil {
		return map[string]interface{}{
			"statusCode": 500,
			"headers":    nil,
			"size":       0,
			"body":       nil,
			"error":      err.Error(),
		}
	}

	if !v.isAllowedHost(parsedURL.Host, v.AllowedHosts) {
		return map[string]interface{}{
			"statusCode": 500,
			"headers":    nil,
			"size":       0,
			"body":       nil,
			"error":      "The requested host isn't allowed",
		}
	}

	if args["method"] != nil {
		method = args["method"].(string)
	} else {
		method = "GET"
	}

	if args["redirects"] != nil {
		redirCount = args["redirects"].(int)
	}

	if args["headers"] != nil {
		hdrs := args["headers"].(map[string]interface{})
		for k, v := range hdrs {
			headers[k] = v.(string)
		}
	}

	if args["proxy"] != nil {
		proxy = args["proxy"].(string)
	}

	body := args["body"]

	client := resty.New()
	client.SetTimeout(time.Duration(v.MaxExecTime) * time.Second)

	if proxy != "" {
		client.SetProxy(proxy)
	}

	client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(redirCount))

	resp, err := client.R().SetHeaders(headers).SetBody(body).Execute(method, target)
	if err != nil {
		return map[string]interface{}{
			"statusCode": resp.StatusCode(),
			"headers":    resp.Header(),
			"size":       resp.Size(),
			"body":       nil,
			"error":      err.Error(),
		}
	}

	var respBody interface{}

	if strings.Contains(strings.ToLower(resp.Header().Get("Content-Type")), "application/json") {
		err := json.Unmarshal(resp.Body(), &respBody)
		if err != nil {
			respBody = string(resp.Body())
		}
	} else if m, err := mxj.NewMapXml(resp.Body()); err == nil {
		respBody = m
	} else {
		respBody = string(resp.Body())
	}

	return map[string]interface{}{
		"statusCode": resp.StatusCode(),
		"headers":    resp.Header(),
		"size":       resp.Size(),
		"body":       respBody,
		"error":      nil,
	}
}

// if the specified host is in allowed hosts
func (v *VM) isAllowedHost(host string, allowed []string) bool {
	if len(allowed) < 1 || allowed[0] == "" {
		return true
	}
	for _, v := range allowed {
		if host == strings.TrimSpace(v) {
			return true
		}
	}
	return false
}
