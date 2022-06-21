# 工具组件离线包制作说明文档

文档提供 caas4.1 中所需工具组件的 rpm 离线压缩包制作说明，根据实际需求，分别基于 x86_64，arm 架构平台下制作。

## docker 离线包

### 环境准备

```bash
CentOS 7.7: 要求系统刚安装完成，防止因安装其他服务导致 rpm 依赖包被提前装好，导致 rpm 无法完全下载.
```

### 基于 x86_64 平台制作示例

制作机器：10.0.0.1
检验机器：10.0.0.2

```bash
## 设置 yum 源。制作机器上执行
$ yum install -y epel-release.noarch

$ yum install -y yum-utils

$ yum-config-manager --add-repo=https://download.docker.com/linux/centos/docker-ce.repo

$ yum-config-manager --enable docker-ce-nightly

$ yum makecache fast

# 创建目录。制作机器上执行
$ mkdir ~/docker && cd ~/docker

# 下载 rpm 文件。制作机器上执行
$ yumdownloader --resolve docker-ce

# 查看。制作机器上执行
$ ls
# 特别说明，根据所在机器的包关系依赖，下载到本地的 rpm 文件数量可能会不一样，若数量少于下述文件数量，请找一台刚安装完成的 centeros 7.7 进行制作，防止 rpm 文件缺失
[root@lb-master-1 docker]# ls
audit-libs-python-2.8.5-4.el7.x86_64.rpm
docker-ce-19.03.12-3.el7.x86_64.rpm
policycoreutils-2.5-34.el7.x86_64.rpm
checkpolicy-2.5-8.el7.x86_64.rpm
docker-ce-cli-19.03.12-3.el7.x86_64.rpm
policycoreutils-python-2.5-34.el7.x86_64.rpm
containerd.io-1.2.13-3.2.el7.x86_64.rpm
libcgroup-0.41-21.el7.x86_64.rpm
python-IPy-0.75-6.el7.noarch.rpm
container-selinux-2.119.2-1.911c772.el7_8.noarch.rpm
libsemanage-python-2.5-14.el7.x86_64.rpm
setools-libs-3.3.8-4.el7.x86_64.rpm


# 压缩。制作机器上执行
$ tar cvzf ~/docker.tar.gz *

# 检验。制作机器上执行
$ scp mkdir docker root@[检验机器ip]:/root/
# 解压 tar 包。检验机器上执行
$ mkdir docker
$ tar xvf docker.tar.gz -C ~/docker

# 安装 rpm 包。检验机器上执行
$ cd docker
$ rpm -ivh --replacefiles --replacepkgs *.rpm

# 启动 docker。检验机器上执行
$ systemctl enable docker.service
$ systemctl start docker.service

# 运行容器。检验机器上执行
$ docker run --name nginx-test -p 8082:80 -d nginx

# 查看容器是否运行。检验机器上执行
$ docker ps
CONTAINER ID        IMAGE               COMMAND                  CREATED             STATUS              PORTS                  NAMES
a88477553375        nginx               "/docker-entrypoint.…"   17 seconds ago      Up 16 seconds       0.0.0.0:8082->80/tcp   nginx-test

```
