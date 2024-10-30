# ai-connect

## gpt镜像

### 配置文件

```json
{
  "chatgpt": {
    "mirror": {
      "address": ":443",
      "tls": {
        "enabled": true,
        "key": "./tls/key.pem",
        "cert": "./tls/cert.pem"
      },
      "tokens": {
        "9090": "eyJhbGciOi"
      }
    }
  }
}
```

### 使用

```shell
# 生成配置文件
cp example.json config.json
# 启动服务
ai-connect chatgpt --mirror
```

上面的配置文件的tokens参数是token映射，比如访问https://oai.253282.xyz?token=9090，会取映射的accessToken，如果
token参数是`eyJhbGciOi`认为是chatgpt的accessToken直接使用
