# 微信SDK合集

### 微信jssdk-go版本

> 可自定义token存储方式，实现jssdk.Storage接口即可
> 默认在当前目录以json文件形式存储
```go
sdk := jssdk.NewJssdk("your appId", "your appSecret",nil)
token, e := sdk.GetAccessToken()
token2, e2 := sdk.GetJsapiTicket()
```

