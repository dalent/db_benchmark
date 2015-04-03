# db_benchmark
测试redis和mongo等的速度
用法
```
 go run benchmark.go -d mongo -u  mgo1:8888 -t 2
```
表示用两个线程，每个线程插入和读取1万次，的时间结果为
```
mode w
threads: 2
per thread: 10000
total time:3.879012986s

mode r
threads: 2
per thread: 10000
total time:3.338706411s
```

同样可以测试redis，因为redis的url不包括passwd信息所以职能控制台提供例如
```
go run benchmark.go --d redis -u  mgo2:6379 -t 2 -p passwd 
```
结果和上边的类似。
