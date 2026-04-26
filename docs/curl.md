```
curl -X PUT -d "hello world" http://localhost:8080/objects/bucket1/file1
```
-Status: 201 Created {"id":"file1"}

```
curl -X PUT -d "hello world" http://localhost:8080/objects/bucket1/file2
```
Status: 201 Created {"id":"file2"}

**Note that the content is the same**


Run after the above PUT:
```
curl -i http://localhost:8080/objects/bucket1/file1
```
Status: 200 OK {hello world}


Delete file 1
```
curl -X DELETE -i http://localhost:8080/objects/bucket1/file1
```
Status: 200 OK

**Note that file2 still exists because it is in a separate bucket**
```
curl -i http://localhost:8080/objects/bucket1/file2
```

Try to delete a nonexistant file
```
curl -i http://localhost:8080/objects/bucket1/imaginary-file
```
Status 400 Not Found