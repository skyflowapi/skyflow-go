# Description
This go module is used to generate service account tokens from service account credentials file which is downloaded upon creation of service account. The token generated from this module is valid for 60 minutes and can be used to make API calls to vault services as well as management API(s) based on the permissions of the service account.

## Usage

```go
package main
    
    import (
    	"fmt"
    	tokenProvider "github.com/skyflowapi/skyflow-go"
    )
    
    func main() {
	token, err := tokenProvider.GetToken("<path_to_sa_credentials_file>")
    	if err != nil {
    		panic(err)
    	}
    
    	fmt.Printf("token %v", *token)
    }
```