package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty"
	"github.com/robertkrimen/otto"
	_ "github.com/robertkrimen/otto/underscore"
)

var (
	flagHTTPAddr     = flag.String("http", ":6030", "the http listen address")
	flagAllowedHosts = flag.String("allowed-hosts", "", "the allowed hosts, empty means `all are allowed`")
	flagMaxExecTime  = flag.Int64("max-exec-time", 5, "max execution time of each script in seconds")
	flagMaxBodySize  = flag.Int64("max-body-size", 2, "max body size in MB")
)

var (
	lastPanic interface{}
)

func main() {
	flag.Parse()
	http.ListenAndServe(*flagHTTPAddr, http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println("[Panic]", err)
			}
		}()
		res.Header().Set("Content-Type", "application/json; charset=UTF-8")
		body, err := ioutil.ReadAll(io.LimitReader(req.Body, *flagMaxBodySize*1024))
		defer req.Body.Close()
		if err != nil {
			res.WriteHeader(422)
			res.Write([]byte(`{"success": false, "error": "` + err.Error() + `"}`))
			return
		}
		queryParams := map[string]string{}
		for k, v := range req.URL.Query() {
			if len(v) < 1 {
				continue
			}
			queryParams[k] = v[0]
		}
		headers := map[string]string{}
		for k, v := range req.Header {
			if len(v) < 1 {
				continue
			}
			headers[k] = v[0]
		}
		value, err := execJS(string(body), *flagMaxExecTime, map[string]interface{}{
			"req": map[string]interface{}{
				"uri":     req.URL.RequestURI(),
				"query":   queryParams,
				"headers": headers,
			},
		})
		if err != nil {
			res.Write([]byte(`{"success": false, "error": "` + err.Error() + `"}`))
			return
		}
		exported, err := value.Export()
		if exported == nil {
			errMsg := lastPanic
			if errMsg == "" {
				errMsg = "unexpected input"
			}
			exported = map[string]interface{}{"success": false, "error": errMsg}
		} else {
			exported = map[string]interface{}{
				"success": true,
				"result":  exported,
			}
		}
		j, _ := json.Marshal(exported)
		res.Write(j)
	}))
}

// exec the specified js script
func execJS(script string, maxTime int64, vars map[string]interface{}) (otto.Value, error) {
	vm := otto.New()
	vm.Interrupt = make(chan func(), 1)

	vm.Set("context", vars)
	vm.Set("exports", map[string]interface{}{})
	vm.Set("request", jsFunctionRequest)

	defer func() {
		if err := recover(); err != nil {
			lastPanic = err
			log.Println("[Panic]", err)
		}
	}()

	go func() {
		time.Sleep(time.Duration(maxTime) * time.Second)
		vm.Interrupt <- func() {
			panic(fmt.Sprintf("reached the maximum execution time (%d sec)", *flagMaxExecTime))
		}
	}()

	val, err := vm.Eval(fmt.Sprintf("%s", script))
	if err != nil {
		return val, err
	}

	exports, err := vm.Get("exports")
	return exports, err
}

// the request function that will be added to the js context
func jsFunctionRequest(args map[string]interface{}) map[string]interface{} {
	var target, method string
	var headers map[string]string

	if args["url"] == nil {
		return nil
	}

	target = args["url"].(string)

	parsedURL, err := url.Parse(target)
	if err != nil {
		return nil
	}
	if !isAllowedHost(parsedURL.Host) {
		return nil
	}

	if args["method"] != nil {
		method = args["method"].(string)
	} else {
		method = "GET"
	}

	if args["headers"] != nil {
		headers = args["headers"].(map[string]string)
	}

	body := args["body"]

	resp, err := resty.R().SetHeaders(headers).SetBody(body).Execute(method, target)
	if err != nil {
		return nil
	}

	var respBody interface{}
	if strings.Contains(strings.ToLower(resp.Header().Get("Content-Type")), "application/json") {
		err := json.Unmarshal(resp.Body(), &respBody)
		if err != nil {
			respBody = string(resp.Body())
		}
	} else {
		respBody = string(resp.Body())
	}

	return map[string]interface{}{
		"statusCode": resp.StatusCode(),
		"headers":    resp.Header(),
		"size":       resp.Size(),
		"body":       respBody,
	}
}

// if the specified host is in allowed hosts
func isAllowedHost(host string) bool {
	if *flagAllowedHosts == "" {
		return true
	}
	for _, v := range strings.Split(*flagAllowedHosts, ",") {
		if host == strings.TrimSpace(v) {
			return true
		}
	}
	return false
}
