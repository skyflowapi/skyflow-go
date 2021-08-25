package main
    
    import (
    	"fmt"
    	tokenProvider "github.com/skyflowapi/golang-sa-token"
		// "github.com/dgrijalva/jwt-go"
    )
    
    func main() {
		token, err := tokenProvider.GetToken("/home/sakethk/Downloads/credentials8.json")
    	if err != nil {
    		panic(err)
    	}
    
    	fmt.Printf("token %v", "")
    }