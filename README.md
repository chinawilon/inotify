#日志监控工具

> 主要需求就是解决监控日志中特定的关键字进行告警通知！
> 
## 配置
```json
{
  "dirPath": "/Users/mac/code/go/inotify",
  "errorKey": "error",
  "noticeTitle": "监控：管理后台出现error，请及时关注!",
  "dingdingAPI": "https://oapi.dingtalk.com/robot/send?access_token=220afa846d61ae5cc022033df758aa8507252574e66e1956c4dfd016ce411751"
}
```

- `dirPath` 监控的日志目录路径，必须绝对路径
- `errorKey` 特定关键字
- `noticeTitle` 告警标题
- `dingdingAPI` 钉钉API接口

## 用法
1. 下载对应系统的inotify-xxx执行文件
2. 编辑inotify.json配置文件，你想要监听的目录
3. 执行notify-xxx + 配置文件

## 举例
```shell
inotify-xxx $PWD/inotify.json
```