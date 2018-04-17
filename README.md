AggreX
=======
> Just a scriptable APIs aggregator using the simple `javascript` syntax

```bash
curl -d "\
var continent = request({url: 'http://country.io/continent.json'}); \
var names = request({url: 'http://country.io/names.json'}); \
var exports = {
	c: continent.body,
	n: names.body
};\
" localhost:6030
```
Features
========
1- Using a simple `javascript` interpreter to execute your requests
2- Using the (`underscore.js`)[http://underscorejs.org] library for helper functions
3- Lightweight & High concurrent request dispatcher
4- Built using `Golang`

Why
====
> I wanted a genaric way to call multiple endpoints from the browser without the ajax hell so I won't reduce the page load performance, as well as I don't want to create a customized script to aggregate the endpoints for me, I need it to be genaric.

Installation
=============
> Just goto the [releases](https://github.com/alash3al/aggrex/releases) page and download yours

Usage
=====

#### CLI Flags
```bash
➜  ~ aggrex -h
Usage of aggrex:
  -allowed-hosts all are allowed
    	the allowed hosts, empty means all are allowed
  -http string
    	the http listen address (default ":6030")
  -max-body-size int
    	max body size in MB (default 2)
  -max-exec-time int
    	max execution time of each script in seconds (default 5)

```

#### Run
```bash
➜  ~ aggrex
```

#### Example
```bash
➜  ~ curl -d 'exports.example = request({url: 'http://country.io/names.json'})' localhost:6030
```

Javascript API
==============
> as well as the basic javascript keywords/objects there are two things `request()` and `_` (underscore)

#### #request(options)
```javascript
var resp = request({
    url: "http://localhost",    // the url
    method: "GET",              // the http method
    headers: {                  // the request headers
        "key": "value"
    },
    body: ""                    // the request body (anything to be sent i.e 'string', 'object' ... etc)
})

var statusCode = resp.statusCode
var headers = resp.headers
var size = resp.size
var body = resp.body

// you must fill the exports variable, because this is the main var used as a response
exports.example = body
```

Contribution
============
Any contribution/suggestion is welcomed

Author
=========
I'm [Mohammed Al Ashaal](http://github.com/alash3al), a Gopher ;)