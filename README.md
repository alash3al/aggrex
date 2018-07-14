AggreX
=======
> Just a scriptable APIs aggregator using the simple `javascript` syntax

```bash
curl -d "\
var continent = fetch({url: 'http://country.io/continent.json'}); \
var names = fetch({url: 'http://country.io/names.json'}); \
var exports = {
	c: continent.body,
	n: names.body
};\
" localhost:6030
```

Features
========
- Using a simple `javascript` interpreter to execute your requests
- Using the [`underscore.js`](http://underscorejs.org) library for helper functions
- Stored procedures`!`
- Procedures search engine`!`
- Lightweight & High concurrent request dispatcher
- Built using `Golang`

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
  -admin-token string
        the admin secret token (default "bbh42d0186vj9c87qdc0")
  -allowed-hosts all are allowed
        the allowed hosts, empty means all are allowed
  -http string
        the http listen address (default ":6030")
  -index string
        the database index (default "/home/alash3al/.aggrex")
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
➜  ~ curl -d "exports.example = fetch({url: 'http://country.io/names.json'})" localhost:6030
```

Javascript API
==============
> as well as the basic javascript keywords/objects there are two things `fetch()` and `_` (underscore)

#### # `fetch(options)`
```javascript
var resp = fetch({
    url: "http://localhost",    // the url
    method: "GET",              // the http method
    headers: {                  // the request headers
        "key": "value"
    },
    proxy: "",                  // the proxy to be used
    redirects: 5,               // the maximum redirects count
    body: ""                    // the request body (anything to be sent i.e 'string', 'object' ... etc)
})

var statusCode = resp.statusCode
var headers = resp.headers
var size = resp.size
var body = resp.body

// you must fill the exports variable, because this is the main var used as a response
exports.example = body
```

### # `utils.btoa(string)`
### # `utils.atob(string)`
### # `utils.uniqid(integer length)`
### # `utils.md5(string)`
### # `utils.sha256(string)`
### # `utils.sha512(string)`
### # `utils.bcrypt(string)`
### # `utils.bcryptCheck(string hashed, string real)`
### # `utils.fetch(Object args)`
### # `utils.collect()`, `underscore.js`

### # `requests.headers` `Object`
### # `requests.query` `Object`
### # `requests.body` `Object`
### # `requests.host` `String`
### # `requests.proto` `String`
### # `requests.remote_addr` `String`
### # `requests.uri` `String`

### # `cron.set(string key, string interval, closure job)`
### # `cron.unset(string key)`
### # `cron.list() Object`

### # `globals` `Object`


RESTful API [![Run in Postman](https://run.pstmn.io/button.svg)](https://app.getpostman.com/run-collection/dac8c42fcce004c6c7e8)
=============
> Goto the postman [documenter]](https://documenter.getpostman.com/view/2408647/aggrex/RW1aJfJ8)

Contribution
============
Weclome :)

Author
=========
I'm [Mohammed Al Ashaal](http://github.com/alash3al), a Gopher ;)