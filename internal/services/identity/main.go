package main

import (
	"context"
	"fmt"

	"github.com/Nerzal/gocloak/v14"
)

func main() {
	client := gocloak.NewClient("http://localhost:8100")
	ctx := context.Background()
	token, err := client.LoginClient(ctx, "2hand-shop-backend", "uH2Uf3jfqi7nD87t9ZIXKBxs0kEPLAMy", "2hand-shop")
	if err != nil {
		fmt.Printf("===============err1:%+v", err)

		return
	}
	fmt.Println("========token:%v", token.AccessToken)

	user, err := client.GetUserByID(ctx, token.AccessToken, "2hand-shop", "1cff1d10-b9c5-4cc3-baee-578a57984e2c")
	if err != nil {
		fmt.Printf("===============err2:%+v", err)
		return
	}
	fmt.Printf("===============user:%+v", *user)
	if user == nil {
		return
	}

	if user.Attributes == nil {
		user.Attributes = make(map[string][]string)
	}
	user.Attributes["internalUserId"] = []string{"test1232131"}

	err = client.UpdateUser(ctx, token.AccessToken, "2hand-shop", *user)
	if err != nil {
		fmt.Printf("===============err3:%+v", err)
		return
	}
}
