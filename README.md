# 注意
- 不要按安装步骤里的目录，按我这个结构来，后面的计算sha256以及写入config.yaml也要对应的改变路径
先git clone把项目下载到任意位置，然后，新建/etc/hids文件夹
/etc/hids下两个文件夹：agent（这个是直接把agent目录copy来的）、plugin（这个要新建）

```
-hids
---agent(目录结构按我这个来，原先的driver/、support/放到plugin文件夹下，journal_watch可以不要)
-----config.yaml(需要自己新建)
-----common
-----log
-----config
-----transport
-----health
-----main.go
-----go.mod
-----go.sum
---plugin
-----driver
-------hids_driver-latest.ko(这个内核模块在LKM下，要先编译好，编译前修改Makefile第一行 MODULE_NAME为hids_driver-latest,编译完复制到这里)
-------driver(这个二进制文件要先在当前目录下用rust编译，然后默认在plugin/driver/target/release下，你需要复制到这里)
-----LKM
-----support(还需要在那个项目的github上找到flexi_logger下载到这个文件夹下，原先是没有的)
```

- /etc/hids/plugin/driver/template.toml 修改socket_path: "/etc/hids/agent/plugin.sock"
/etc/hids/plugin/drive/src/config.rs 修改pub const SOCKET_PATH: &str = "/etc/hids/agent/plugin.sock"

- shasum -a 256 /etc/hids/plugin/driver/driver
echo "plugins: [{name: hids_driver,version: 1.5.0.0,path: /etc/hids/plugin/driver/driver,sha256: }]" > /etc/hids/agent/config.yaml

- 保证运行./agent前要先rmmod hids_driver-latest，否则会报错

# Honeypot

#### 介绍
蜜罐监控--By lyx.

#### 软件架构
软件架构说明


#### 安装教程

1.  xxxx
2.  xxxx
3.  xxxx

#### 使用说明

1.  xxxx
2.  xxxx
3.  xxxx

#### 参与贡献

1.  Fork 本仓库
2.  新建 Feat_xxx 分支
3.  提交代码
4.  新建 Pull Request


#### 特技

1.  使用 Readme\_XXX.md 来支持不同的语言，例如 Readme\_en.md, Readme\_zh.md
2.  Gitee 官方博客 [blog.gitee.com](https://blog.gitee.com)
3.  你可以 [https://gitee.com/explore](https://gitee.com/explore) 这个地址来了解 Gitee 上的优秀开源项目
4.  [GVP](https://gitee.com/gvp) 全称是 Gitee 最有价值开源项目，是综合评定出的优秀开源项目
5.  Gitee 官方提供的使用手册 [https://gitee.com/help](https://gitee.com/help)
6.  Gitee 封面人物是一档用来展示 Gitee 会员风采的栏目 [https://gitee.com/gitee-stars/](https://gitee.com/gitee-stars/)
>>>>>>> d92b3937b48a5a6904656e408b1e63777ac486b0
