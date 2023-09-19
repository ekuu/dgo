# ddd

# proto

## 生成领域事件的通用格式
```shell
cd dgo
protoc -I=. --go_out=. pb/event.proto
```

## 生成示例中的具体事件
```shell
cd dgo/internal
protoc -I=. --go_out=. examples/pb/event.proto
```
