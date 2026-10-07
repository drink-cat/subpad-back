
已有表，使用gin开发对应的API，增删改查等。
path 加前缀 /api 
method 只使用 GET POST 
API的请求、响应，都要用 驼峰写法。


返回的通用结构：
type BaseResp struct {
    Code int64
    Message string 
    Data any 
}

用户管理：
登录后，返回jwtToken。会话保持。
gin增加 JwtFilter 。查用户表，放入gin上下文。

域名解析：
gin增加 DomainFilter 
如果域名格式类似 foods.launch.o1.local，则把 foods 拆出。用brand字段，查subpad表，放入gin上下文。















