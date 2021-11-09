# tcp-client

拥有以下的特点：

1. 与任意TCP服务端建立长链接进行同步交互
2. 实现类似HTTP的请求-响应模型
3. 支持处理服务端请求的handler
4. 支持异常断连的退避重试
5. 兼容protobuf、json等序列化协议
6. 通过记录包长，避免了TCP协议的分包、粘包问题
7. 实现服务端热切换的快速响应
8. 单侧覆盖率90%+


如果有用，欢迎点🌟支持，QQ185734549
