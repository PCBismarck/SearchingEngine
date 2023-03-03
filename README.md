# SearchingEngine
一个简单的搜索引擎
参考了 [gofound](https://github.com/sea-team/gofound) 的实现<br/>
使用leveldb作为数据库以及倒排索引的存储<br/>
目前仅支持**查询**功能，数据库的更新需要使用**init_levelDB/load_doc_test.go**来根据文档库生成新的数据库
## 倒排索引的创建
基本逻辑为：
1. 读取文档存入leveldb数据库中
2. 对存入数据库中的文档使用gojieba的提取功能提取文档中的30个关键字
3. 以2中提取出的关键字作为索引充当新的key，将文档的id集合作为value，构建倒排索引数据
