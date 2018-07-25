SimpleReverseProxy is a basic reverse proxy server that can overwrite string patterns in the response body and serve static files. Can also modify request leaving proxy and response. This was written primarily for simulating remote systems while developing and most importantly for fun..

## Run instructions
```go run main.go```

## Run Demo

1. Run test server (server behind proxy) ```go run ./test/test_server.go```

2. Start reverse proxy 
```go run main.go --config=config.json``` 

3. Look at the response in the only route in test_server.go

4. Visit [http://localhost:7575/hello](http://localhost:7575/hello)

5. Notice that the response body has been modified with a find and replace specified in config.json

6. Visit localhost:7575/static/{file-name-in-dir-test} to access static files

## Run tests
```go test``` or ```go test -v```

## Configuration
```
{
  "proxy-host": "localhost:9595", // server behind proxy
  "port": ":7575", // port of proxy server
  "regex-find-replace": { 
    "a":"A", // keys are the regex pattern and values are the value that will be used to replace 
    "b":"B", // simple pattern
    "c":"C", 
    "d":"D",
    "e":"E",
    "f":"F",
    "g":"G"
    "p(x*)q": "T" // more advanced regex pattern
  },
  "static-dir-url-root": "/static", // url root of static files host by proxy
  "static-dir-root": "./test" // location of static files hosted by proxy
}
```
