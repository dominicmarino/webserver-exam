# Project Title

Webserver Coding assessment for RedHat Senior Observability Engineer

## Description

This project implements a simple webserver that stores HTTP text bodies in memory. It supports the following operations:
* Upload object to the service
  * Request: PUT /objects/{bucket}/{objectID}
* Download an object from the service
  * Request: GET /objects/{bucket}/{objectID}
* Delete an object from the service
  * Request: DELETE /objects/{bucket}/{objectID}

Objects, which are just HTTP request bodies, are deduplicated on a per bucket basis. This means that two buckets can have the same data in them,
but one bucket only ever has one instance of a particular HTTP request body in it. Implementation details and caveats are below.

## Getting Started

### Dependencies

A minimal prometheus server returning go and process metrics is implemented using the prometheus libraries at github.com/prometheus/client_golang/prometheus

### Executing program

#### General execution
Run the program by executing 
```
go run webserver.go
```
Optionally, you can specify the port to listen for requests by using the `--port` flag
```
go run webserver.go --port 9090
```
By default, it uses port `8080`. As mentioned earlier, there is a prometheus server presenting metrics included in this application. You can access it on port `2112`. 

#### Using cURL 
The webserver is primarly accessed by using HTTP requests. Supported commands are GET, PUT, and DELETE. All three commands require a path format of `/objects/{bucket}/{id}`.  
No other paths will be accepted and will return a `404`.

#### Metrics
Prometheus metrics can be accessed by running
```
curl -X GET http://localhost:2112/metrics
```

### Testing
#### Unit tests
A suite of unit tests is provided in `webserver_test.go`. To execute them, in the directory the code is in simply run
```
go test
```

#### Curl tests
A set of basic curl commands is located at [Curl commands for testing](docs/curl.md)

### Further documentation
[AI Usage in this project](docs/ai.md)

[Data Structure description and assumptions](docs/architecture.md)


## Version History
* 0.1
    * Initial Release

## Contact

Dominic Marino

dominic `dot` marino `at` gmail `dot` com
