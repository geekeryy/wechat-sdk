# 微信SDK合集

### 微信jssdk-go版本

#### 项目内容

介绍：
 1. 参照官方jssdk-php版本实现；
 2. 可自定义token存储方式，实现jssdk.Storage接口即可；
 3. 默认在当前目录以json文件形式存储；
 
 
示例：
```go
// 初始化jsskd对象
sdk := jssdk.NewJssdk("your appId", "your appSecret",nil)
// 获取AccessToken
token, e := sdk.GetAccessToken()
// 获取JsapiTicket
token2, e2 := sdk.GetJsapiTicket()
// 获取签名内容
signPackage, e3 := sdk.GetSignPackage("http://www.jiangyang.me")
```

