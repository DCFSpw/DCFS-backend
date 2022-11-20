# DCFS-backend

## Documentation
Code should be written in a self-documenting way using the `godoc` binary.

### 1.  Godoc installation
Pull up your terminal and paste the following:

```go install golang.org/x/tools/cmd/godoc@latest```

### 2. Serve docs
```godoc --http=:6060```

Access the documentation on `http://localhost:6060/pkg/dcfs`

### 3. Code writing guidelines
- place comments above public functions with appropriate info:
    
  ```
  // FuncName - short description (one line)
  // <one line empty>
  // <more details if needed>
  // <one line empty if details provided
  // params:
  // - param1 - ...
  // ...
  // <one line empty>
  // return type:
  //   <return type>
  // <one line empty>
  // exceptions:
  // - ...
  func (r *return_type) FuncName(params...) {...}
  ```
  
- make sure that if commenting a method the struct is public as well (godoc will not create documentation for private structs)

## Unit Testing

### Guidelines <TBD>

### Run
You can just run the Unit Test coverage report by simply:
```
go test ./...
```
This will just run all the tests from all the available test files in the project dir.
Given the fact that all the UTs are located in <DCFS>/test/unit, the UTs should be run that way:
```
go test -v ./test/unit
```
(the `-v` flag means `verbose`).

### Test coverage
To generate the test coverage, you should do the following:
```
go test -cover -coverpkg "./models" -v ./test/unit
```
This will just generate coverage percentage in the terminal.

To get a nicely formatted HTML report do the following:
```
go test -cover -coverpkg "./models" -v -coverprofile cover.out ./test/unit
go tool cover -html cover.out -o cover.html
```
Afterwards delete the files `cover.out` and `cover.html`. Please be mindful **NOT** to add them to git.

### GoConvey UI
You can also use the GoConvey HTTP UI. Unfortunately, given the fact that the coverage settings can't be configured no coverage data from this UI will be accurate. Additionally this way is extremely slow.