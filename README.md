#日志监控工具

> 主要需求就是解决监控日志中特定的关键字进行告警通知！

## 配置
```json
[
    {
      "dirPath": "/Users/mac/code/go/inotify/log",
      "filterFile": "\\.log$",
      "errorKey": "error|fatal",
      "excludeKey": "23000\\]",
      "notifyTypes": {
        "dingding": {
          "title": "监控：管理后台出现error，请及时关注!",
          "api": "https://oapi.dingtalk.com/robot/send?access_token="
        },
        "shell": {
          "command": "echo"
        }
      }
    }
]

```
#### 注意是数组，支持一次定义多个监控目录和相关的配置
- `dirPath` 监控的日志目录路径，必须绝对路径
- `filterFile` 指定特定的文件，支持正则表达式
- `errorKey` 错误关键字，支持正则表达式
- `excludeKey` 排除那些关键字，支持正则表达式
- `notifyTypes` 通知方式，支持多个，目前支持`dingding/shell`
    - `shell` 注意执行的shell这里会把`changedFile`和`newContent`会当作参数传入
      - `changedFile` 修改的文件
      - `newContent` 修改的内容

## 用法
1. 下载对应系统的inotify-xxx执行文件
2. 编辑inotify.json配置文件，你想要监听的目录
3. 执行notify-xxx + 配置文件

## 举例
```shell
inotify-xxx $PWD/inotify.json
```