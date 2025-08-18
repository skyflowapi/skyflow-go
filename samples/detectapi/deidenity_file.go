/*
Copyright (c) 2022 Skyflow, Inc.
*/

package main

import (
	"context"
	"fmt"

	"github.com/skyflowapi/skyflow-go/v2/client"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

func main() {
	vaultConfig1 := common.VaultConfig{VaultId: "d381d995d003445b90081add46e2317f", ClusterId: "qhdmceurtnlz", Env: common.DEV, Credentials: common.Credentials{Token: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2MiOiJiOTYzZTcxMjFkZDY0YTZkOTY0MGQ3ZTNlNGNjODdhNyIsImF1ZCI6Imh0dHBzOi8vbWFuYWdlLWJsaXR6LnNreWZsb3dhcGlzLmRldiIsImV4cCI6MTc1NTU0NjgwMSwiaWF0IjoxNzU1NDYwNDAxLCJpc3MiOiJzYS1hdXRoQG1hbmFnZS1ibGl0ei5za3lmbG93YXBpcy5kZXYiLCJqdGkiOiJlNDAyNTJmZGQ5YTg0NDFjOWFmYjQzNzUxNTc5MDEyYyIsInN1YiI6ImJlNTliYjc4MjA3ZTQ5MjE5OTRhMGJhNGNjMDkxMjg5In0.Exopu1luaWeeECy5tWuaQAxuvyZKqupDXZgmH7J8IGbeQM2udqEpjCXriIt58Q0bQVcj8I_mzc-KXZ-tM3e0omoQX1Pm8oM5OyAxJ7mqTJyF9BpnvNqnAPkMhI3UjMY9u86ef2UiV8V7QfNTeWPsNuTjINRIXxChVCFvhTA9Y2TxlqGvl5AuCEnTT1xwEXtZAlqBX_TF8475xlP-cLMcXHdX7NmTmqZbOFDVmJ4Lj_BWiVyLKqQz_jJFZR3r51WlAR9Ujt6pW7dRws8Ypb-gBueqbLwGwZxJYK8ySdRsmnmBqZnkFf_XerEJfj5FwBB8ug9PtkGZHoadHbfdatWNFQ"}}
	// vaultConfig2 := common.VaultConfig{VaultId: "rd1c901b696b44a0a0f46dee1e3c10ec", ClusterId: "qhdmceurtnlz", Env: common.DEV, Credentials: common.Credentials{Token: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2MiOiJiOTYzZTcxMjFkZDY0YTZkOTY0MGQ3ZTNlNGNjODdhNyIsImF1ZCI6Imh0dHBzOi8vbWFuYWdlLWJsaXR6LnNreWZsb3dhcGlzLmRldiIsImV4cCI6MTc1NTA3MTgzMywiaWF0IjoxNzU0OTg1NDMzLCJpc3MiOiJzYS1hdXRoQG1hbmFnZS1ibGl0ei5za3lmbG93YXBpcy5kZXYiLCJqdGkiOiJkZTMwMDRmOTZkNTc0NDZjOTY0NTZhZmIwYWQ3ODZlNyIsInN1YiI6ImJlNTliYjc4MjA3ZTQ5MjE5OTRhMGJhNGNjMDkxMjg5In0.VodmCjqB0YtINIXYuW_Juzy4ZBT_dTVJx1tgfuBA1bDzXHdsayEjvhTQLBUQXHuCVztbdaziAU3scxrNuB5ERMUh5B9VO3rJ3Bj9RJca9Lfwzq-8FQfewQ1XlnfIOjXOazrtfyVBRiPsJ4FoQToRlyaBmifM6Urli1A01ZXVHGj4Y1NnnfOgpcbGFVvlTgFas0lm2WyyL03I-ombHYRJtuZqrBORIo1hzshKR8HzpUkzJLgRW7MQYkvAujlSbXYlzw2xWHW89Ywumy_1OiziCNHPWXYCEbFyrdk7KBZlvg8RWnpq1vfB38mP6pPoNTHP7kB7JzswaT1iqVQDHUYeyA"}}
	var arr []common.VaultConfig
	arr = append(arr, vaultConfig1)
	skyflowInstance, err := client.NewSkyflow(
		client.WithVaults(arr...),
		// client.WithCredentials(common.Credentials{}), // pass credentials if not provided in vault config
		client.WithLogLevel(logger.DEBUG),
	)
	if err != nil {
		fmt.Println(*err)
	} else {
		service, serviceErr := skyflowInstance.Detect("d381d995d003445b90081add46e2317f")
		if serviceErr != nil {
			fmt.Println(*serviceErr)
		} else {
			ctx := context.TODO()
			fmt.Println(service)
			res, err := service.DeidentifyFile(ctx, common.DeidentifyFileRequest{
				FileInput: common.FileInput{
					FilePath: "/home/raushan.gupta/image/demo-folder/card4.jpeg",
				},
				OutputDirectory: "/home/raushan.gupta/image/gen-image",
				// MaskingMethod: common.MaskingMethod{
				// 	common.BLUR
				// },
			})
			if err != nil {
				fmt.Println(*err)
			} else {
				fmt.Println(res)

			}

		}
	}

}
