# Golang service account token
Provides service account token when given a credentials file

## Usage

```go
package main
    
    import (
    	"fmt"
    	tokenProvider "github.com/skyflowapi/golang-sa-token"
    )
    
    func main() {
		token, err := tokenProvider.GetToken("<path_to_sa_credentials_file")
    	if err != nil {
    		panic(err)
    	}
    
    	fmt.Printf("token %v", *token)
    }
```