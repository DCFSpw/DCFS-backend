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