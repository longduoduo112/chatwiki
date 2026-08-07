### 说明文件

1.此目录是运行本地文档转换任务的目录,内容不跟随git仓库走

2.目录结构

```
clawbot/working_dir/query-local-docs/convert_docs.py(转换脚本)
clawbot/working_dir/query-local-docs/<task_batch>/*.*(任务临时文件,结束后会删除)
```