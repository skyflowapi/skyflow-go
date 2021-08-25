# Description
skyflow-go is the Skyflow SDK for the Go programming language.

## Usage

### Service Account Token Generation
This go module is used to generate service account tokens from service account credentials file which is downloaded upon creation of service account. The token generated from this module is valid for 60 minutes and can be used to make API calls to vault services as well as management API(s) based on the permissions of the service account.

```go
package main
    
    import (
    	"fmt"
    	saUtil "github.com/skyflowapi/skyflow-go/service-account/util"
    )
    
    func main() {
	token, err := saUtil.GenerateToken("<path_to_sa_credentials_file>")
    	if err != nil {
    		panic(err)
    	}
    
    	fmt.Printf("token %v", *token)
    }
```