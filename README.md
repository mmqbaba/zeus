# zeus公共库

# 代码生成工具 tools/bin/gen-zeus

## 生成二进制文件

```bash
go build -o tools/bin/ ./tools/gen-zeus
```

## 将工具添加系统环境变量中
* linux : export PATH=$PATH:$GOPATH/src/zeus/tools/bin
* windows : PATH=%PATH%;%GOPATH%/src/zeus/tools/bin