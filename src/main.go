package main

import (
	"fmt"
	"wechat-sdk/jssdk"
)

func main()  {
	jssdk1 := jssdk.NewJssdk("wx5d648fda319bed0c", "643a09b180876a550e20cea6d32c8095",nil)
	//token, e := jssdk1.GetAccessToken()
	//fmt.Println(token, e)
	//token2, e2 := jssdk1.GetJsapiTicket()
	//fmt.Println(token2, e2)
	fmt.Println(jssdk1.GetSignPackage("http://www.jiangyang.me"))
}