#### introduction
- 利用elasticsearch对个人pocket存档进行搜索的工具, 包括同步pocket文档，拉取最新pocket文档

#### objective
- 解决pocket文档的搜索问题

#### target user
- me or other software engineers

#### search specifications
- 支持tags，title，excerpt的精确和模糊搜索
- 搜索优先级tags > title > excerpt

#### get start
> 确保本地搭好elasticsearch
- clone the repo the run the `go build`

#### usage
``` shell
./pfs -help # 查看帮助
./pfs -init # pocket授权
./pfs -sync # 同步pocket数据到es中
./pfs -rmIndex # 删除es的中index
./pfs -search -text {search input} # 搜索，exact match term
./pfs -fzsearch -text {search input} # 搜索，wildcard match
./pfs -search -tag tag1 -tag tag2 # 按照tag搜索(多个tag是与的关系)
```
