# 飞书开放接口教程示例代码

## 概述
- 主要为飞书IM域中的开放文档示例

## 运行环境
- 建议Golang 1.5及以上

## 使用主要事项
- conf/config.go 文件中替换应用的appID和secret
- 如果在测试环境运行，需要替换URL域名
- 创建群和邀请人入群，请替换UserA，UserB，UserC为应用可见用户的openID，获取应用的openID方法详见[如何获得UserID、OpenID和UnionID](https://open.feishu.cn/document/home/user-identity-introduction/how-to-get)
- 发送消息中，如果需要使用图片，需要上传图片，可将文件放在upload目录下
- 监听消息需要启动服务

## 问题反馈
有任何问题请联系飞书官方客服
