# ddd

# proto
```shell
cd dgo
```
## 生成领域事件的通用格式
```shell
protoc -I=. --go_out=. pb/event.proto
```

## 生成示例中的具体事件
```shell
protoc -I=. --go_out=. examples/pb/event.proto
```
