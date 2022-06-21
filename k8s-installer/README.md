# k8s installer

k8s installer 是一个 k8s 集群生命周期管理工具，他分为 2 个组件: 服务端(server)、客户端(client)

## 支持的 cpu 架构

- x86_64(经过测试)
- aarch64(未经过测试)

## 支持的操作系统

- centos-7.8(经过测试)
- centos-7.7(经过测试)
- centos-7.6(经过测试)
- 其他centos-7.x(未经过测试)

## 设计文档

详见 <http://172.16.30.21/openshift_origin/caas-documentation/blob/master/doc/mec-caas-4-1/Design.md>

## Getting Start

### 编译源码（限 Linux 系统）

```bash
git clone git@172.16.30.21:openshift_origin/k8s-installer.git
cd k8s-installer
make all
```

检查是否有漏网的 rpm

```bash
rpmList=$(grep -r -E -i "\b\S+\.rpm" * | grep -v -E "^\S+\s*//" | grep -v swagger | grep -v k8s-installer-deps | grep -v schema/upgrade.go | grep -v Binary | grep -v docs | awk '{print $3}' | awk -F \" '{print $2}')
for y in $rpmList;do grep $y hack/deps/k8s-installer-deps.json > /dev/null || echo "==$y==";done
```

## 本地调试

1. run etcd localhost: <https://github.com/etcd-io/etcd/releases/tag/v3.4.16>
1. copy config-server.yaml to ~/.k8s-installer

    ```yaml
    disable-lazy-operation-log: true
    api-server:
    api-port: 9090
    enable-tls: false
    #jwt-expired-after-hours: 8
    jwt-expired-after-hours: 1
    jwt-sign-method: HS256
    jwt-sign-string: |-
        LS0tLS1CRUdJTiBPUEVOU1NIIFBSSVZBVEUgS0VZLS0tLS0KYjNCbGJuTnphQzFyWlhrdGRqRUFB
        QUFBQkc1dmJtVUFBQUFFYm05dVpRQUFBQUFBQUFBQkFBQUJGd0FBQUFkemMyZ3RjbgpOaEFBQUFB
        d0VBQVFBQUFRRUFyYkdsenMrZHUwZmI4NE5IbU8wa1VtRU9iZEFkSisvaGhsZ3QzTFU0WWxBUXRO
        cUY4VUw3Clp5VEY3cDdTcFB0UnFRY1FyUTJzbE8zQmEydEttZFRhcUdIOWF2YVpWRm5sOXU4Vmc4
        NG02NGdoTjhuTlI4QnN0QjFBTjkKbDl1K1BQYTR5TFF5WkY1dzZaSVA0Vy9RV2tkaHRxYWE5WkZR
        Sk5qeW55N3FZc1B5QTRPODI4dFNOaldRSHU2YUNVK1RiTgpLRU1sUnB2RmhWMStmOTJxWGpkalBi
        YS9DTys0bUhNNTMvby9VeW5TSlZ2eU9WUE4vdGErN29GQUhydnFLZXFYbFR5eEx0CnZveGNVWHpv
        ZzNlY1F1NGdRaFFLSzAzSWZ4aG9LSjNaYnA0VGtxdm9HUGJ0S2NmYStIZ0Z1eXFncTI3T1dKYSta
        Q2oweEMKclFLQUtpQWo3UUFBQThnOFNSNmtQRWtlcEFBQUFBZHpjMmd0Y25OaEFBQUJBUUN0c2FY
        T3o1MjdSOXZ6ZzBlWTdTUlNZUQo1dDBCMG43K0dHV0MzY3RUaGlVQkMwMm9YeFF2dG5KTVh1bnRL
        aysxR3BCeEN0RGF5VTdjRnJhMHFaMU5xb1lmMXE5cGxVCldlWDI3eFdEemlicmlDRTN5YzFId0d5
        MEhVQTMyWDI3NDg5cmpJdERKa1huRHBrZy9oYjlCYVIyRzJwcHIxa1ZBazJQS2YKTHVwaXcvSURn
        N3pieTFJMk5aQWU3cG9KVDVOczBvUXlWR204V0ZYWDUvM2FwZU4yTTl0cjhJNzdpWWN6bmYrajlU
        S2RJbApXL0k1VTgzKzFyN3VnVUFldStvcDZwZVZQTEV1MitqRnhSZk9pRGQ1eEM3aUJDRkFvclRj
        aC9HR2dvbmRsdW5oT1NxK2dZCjl1MHB4OXI0ZUFXN0txQ3JiczVZbHI1a0tQVEVLdEFvQXFJQ1B0
        QUFBQUF3RUFBUUFBQVFBK1F1T3drbk56NG5wUmU4bDYKWStjVk1IMC9sODRidHIwY3J4Y2hla1JQ
        Mld0anFNRkNqa1FYNFBLaWFvUVBaNWNLQStKU1pnaHJDaDYvSnFLREtlMkhWagpqRTBzaDdtQTM2
        eWhEb1FrbHBQRTdMOUthRkJkRHhiMXJKcWtpTHhVbGd2K3hia2FpVS9vS2RkUGRBazNrMGJQZGtF
        dHJYCjBRK0VOZ0ZDMG9ZaHlnOVZJSDVFQVdHOEgxZWRaOTVRRGsyZmJaYmJkdmJKUExJRjJEeWNr
        ZlBLUmpZRkZKa2paV3hzcDMKUzJ4cXpkaE1IblNvZ0RWZ0ptalhKQ00vUTZZdTdBU1NvVTN0OFNl
        VThtY1J0ZitDZUM1VnNNbkt3MFZNdjNmNnlCcmhObwprWGNsR0p0NVZrT3Y5ZlpXc3ZmUHprRFY4
        M0RrOUhYVWNwZS9SNlYwVWJCeEFBQUFnQVA1QkNkbURKZ28vYkxONTZuZUJnCnoxNkpSdnBCbGFX
        aENNTmZFNWUzVVZnQkpPWXRQRnpEN3NrSWZOVzRFdXQ3Rm94clJRUjBpenhHcGFuUTNRNitmNHR3
        MW0KVi9ScGY5ZGdaV0dpS3dsR2gwMHlDQkx3Y2dPaXl5cUpNbllIaWJ1U1ZXOU1KanFSaE53T05C
        NGtWczhhVUxEVWZVNWkvLwpKS2FhTkVrZmFMQUFBQWdRRGg3WlZUVjBHQ0lxK0Uwd24yckJwNlhx
        RzA4dTZMTVpwbURZMFFXR3BvZElaWjEvazBmbzRFCngxN1RNZ3ZqNlkvYlBWZkl3TXVCS0ovNk5W
        OVFsUmJydm5GYTd4elVGUENoWFRZNTVNNDFGaTJ0VDJNS2N0STA4VHlDNk8KSlFRYUJCNTlURXJx
        NlRQclZWLzVySFptSVB2TEtKY1hibTNIUHlybmpLMGxGbmt3QUFBSUVBeE5BMG9EMS8xQk5PMTh0
        YgozUENiZjZwdDVDeitQRitxNlZlb082U2N1T3VEL3hHNWQ0RC83ZnEweHNWVEVlTTB3UWFTYVVK
        c3BQeC93VVhyMmZzWXRFCjNMUkh4cGp5bDdJV2NTOE14cFhNRWE5RDl6Y04wMTdmZHd5Y2tMWCtI
        akdnWGlpQW1vTUE5UTlHSXU2VHFkQWZ1S1VtYTkKV2tzNmhGNUVRem9BZG44QUFBQU9jbTl2ZEVC
        emFXMXZibk10Y0dNQkFnTUVCUT09Ci0tLS0tRU5EIE9QRU5TU0ggUFJJVkFURSBLRVktLS0tLQo=
    resource-server-cidr: ""
    resource-server-filepath: /usr/share/k8s-installer/resource
    resource-server-port: 8079
    cache:
    cache-runtime: no-cache
    etcd:
    etcd-auth-mode: none
    #etcd-auth-mode: tls
    etcd-ca-file-path: /etc/etcd/etcd-client-ca.crt
    etcd-cert-file-path: /etc/etcd/etcd-client.crt
    #etcd-endpoints: https://etcd-node1:2379 | https://etcd-node2:2379 | https://etcd-node3:2379
    etcd-endpoints: http://localhost:2379
    etcd-key-file-path: /etc/etcd/etcd-client.key
    etcd-password: ""
    etcd-username: ""
    host-requirement:
    check-kernel-version: false
    check-os-family: true
    kernel-major: 3
    kernel-sub: 10
    kernel-ver: 0
    support-os-family: centos,ubuntu
    support-os-family-version: 7,
    log:
    log-level: 6
    message-queue:
    auth-mode: basic
    ca-cert-path: ""
    cert-path: ""
    cluster-port: 9890
    key-path: ""
    message-queue-group-name: caas-node-status-report-queue
    message-queue-leader: 127.0.0.1
    message-queue-node-status-report-in-subject: caas-node-status-report-subject
    message-queue-server-address: 127.0.0.1
    message-queue-server-listening-on: 0.0.0.0
    message-queue-subject-suffix: k8s-install
    password: password
    port: 9889
    time-out-seconds: 10
    username: username
    server-id: server-MTcyLjE2LjIwNi4yMDc=
    server-message-timeout-config:
    task-basic-config-timeout: 5
    task-cri-timeout: 300
    task-kubeadm-init-first-control-plane-timeout: 300
    task-kubeadm-join-control-plane-timeout: 300
    task-kubeadm-join-worker-timeout: 300
    task-kubectl-timeout: 5
    task-prepare-offline-resource: 0
    task-rename-hostname-timeout: 10
    task-vip-timeout: 60
    signal-port: 9091
    system-info-pubkey: AQ3ONZ55HMRBWFHVT6MPV5WAGZCAALCK3VSV3VO5PJI5OW4MPBZ6EN26WBBSD5MNS4R3LJT3ACCSGFXTZXFTHN2A5RFOA6DR5DFZRYT6UODBHT5RYKIYSDE372JMEMPDDM5I3XUGGCEFIZJQX6CRH3ZXA6SQ====
    is-test-node: true
    report-to-kube-caas: false
    coredns-config:
    notify-mode: now
    ```
1. 注入 license

    ```bash
    etcdctl  del "" --from-key=true
    etcdctl  put /roles/1 "{\"id\":\"1\",\"name\":\"role1\",\"function\":134217727}"
    #etcdctl  --endpoints=https://etcd-node1:2379 --cert=/etc/etcd/etcd-client.crt --key=/etc/etcd/etcd-client.key --cacert=/etc/etcd/etcd-client-ca.crt  put /roles/1 "{\"id\":\"1\",\"name\":\"role1\",\"function\":16777215}"
    etcdctl  put /users/admin "{\"id\":\"1\",\"username\":\"admin\",\"password\":\"123\",\"roles\":[{\"id\":\"1\"}]}"
    etcdctl  put /license/ "{\"license\":\"FAKE-LICENSE\"}"
    ```
1. VSCode remote-ssh 调试，可以从 login 测试起（login 绕过了 license 检查，见 LicenseValid 方法）

    ```bash
    curl -i -s -X POST http://localhost:9090/api/core/v1/login     --header 'Content-Type: application/json'     --data '{
        "username": "admin",
        "password": "123"
    }'
    ```

## 玩起来

本实验目的在于通过1台 server 来部署 2 套 3 master 的 k8s 集群，总共需要有 7 台机器

- server 主机(4c 8G) X 1 -- 部署 server 二进制文件和 etcd 数据库
- client 主机(2c 4G) X 6 -- 部署 client 二进制文件

### 配置离线安装下的资源文件服务(可选-仅在离线环境下可用)

为了方便 server 二进制文件自带 resource server 的服务，默认他会启动在 10098 端口，你只需要把的 rpm 包放在 /tmp/k8s-installer/resource/1.18.6/centos/7/x86_64/package 下即可，这里注意 resource 下层的目录遵循规范 [k8s version]/[os family]/[os version]/[cpu arch]/package 的规则，你可以从这里下载到依赖的 [rpms](https://seafile.sh.caas.net/d/ea5ac7b7125b4be98701/)  注意开发中我们的 rpm 的更新可能没有跟上开发的节奏如果安装中出现包缺失， 欢迎通知我们:)

### 复制 server 二进制文件至 server 主机

```console
# 复制二进制文件到目标主机
$ scp [path]/k8s-installer/cmd/server/server [user]@[server ip]:/root
# 创建配置文件目录
$ mkdir /root/.k8s-installer/
```

### 复制 client 二进制文件至 client 主机

```console
# 复制二进制文件到目标主机
$ scp [path]/k8s-installer/cmd/client/client [user]@[client ip]:/root
# 创建配置文件目录
$ mkdir /root/.k8s-installer/
```

### 配置 server 的配置文件

在 server 主机 `/root/.k8s-installer/` 创建一个文件 `config-server.yaml` ，参见 etc 目录

### 配置 client 的配置文件

在所有的 client 主机上的 `/root/.k8s-installer/` 创建一个文件 `config-client.yaml`，参见 etc 目录

### 启动 server 守护进程

在 server 主机上运行以下命令

```console
# 启动 etcd 数据库
$ etcd # 你可以自行下载合适的版本 https://github.com/etcd-io/etcd/releases
$ cd /root
$ ./server
```

### 启动 client 守护进程

在 client 主机上运行以下命令

```console
$ cd /root
$ ./client
# 等待 1 分钟，让所有的节点汇报他们状态
```

### 获取 token

我们的 api 是需要通过 jwt 身份认证的，需要通过 login 方法传入合法用户和密码来获取 jwt ，问题在于没有使用部署脚本情况下我们需要预先创建 admin 帐号和一个 超级管理员的 角色

```console
# 使用 etcdctl 工具创建用户和角色

# tls
# etcdctl  --endpoints=https://[endpoint]:2379 --cert=/etc/etcd/etcd-client.crt --key=/etc/etcd/etcd-client.key --cacert=/etc/etcd/etcd-client-ca.crt  put /users/admin "{\"id\":\"1\",\"username\":\"admin\",\"password\":\"123\",\"roles\":[{\"id\":\"1\"}]}"

# no tls
$ etcdctl --endpoints=http://localhost:2379 put /users/admin "{\"id\":\"1\",\"username\":\"admin\",\"password\":\"123\",\"roles\":[{\"id\":\"1\"}]}"
# 目前代码中 131072 满足一个超级管理员的权限了
$ etcdctl --endpoints=http://localhost:2379 put /roles/1 "{\"id\":\"1\",\"name\":\"role1\",\"function\":262143}"

# 获取 token
$ curl --location --request POST 'http://localhost:9090/user/v1/login' \
--header 'Content-Type: application/json' \
--data-raw '{
    "username": "admin",
    "password": "123"
}'
# 返回结果
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjEiLCJ1c2VybmFtZSI6InNpbW9uIiwicGFzc3dvcmQiOiIxMjMiLCJyb2xlcyI6bnVsbH0.TYuoImyKeHGHZqSivb309gqnwjeyDys2tbUb296ETG4"
}
```

### 同时安装 2 套集群

在 server 主机上运行命令,注意下列的 node_id 都来自于 节点自动生产的，你可以通过每个 client 节点的 `/root/.k8s-installer/config-client.yaml` 中的 client-id 字段找到

```console
# 安装第一套集群

$ curl --location --request POST 'http://localhost:9090/cluster/v1' \
--header 'Content-Type: application/json' \
--header --header 'token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjEiLCJ1c2VybmFtZSI6InNpbW9uIiwicGFzc3dvcmQiOiIxMjMiLCJyb2xlcyI6bnVsbH0.TYuoImyKeHGHZqSivb309gqnwjeyDys2tbUb296ETG4' \
--data-raw '{
    "ClusterId": "cluster1",
    "cluster_name": "cluster1",
    "control_plane": {
        "service_v4_cidr": "10.96.0.0/16",
        "enable_ipvs": true
    },
    "container_runtime": {
        "container_runtime_type": "docker",
        "private_registry_address": "10.0.0.10",
        "private_registry_port": 4000,
        "private_registry_cert": "",
        "private_registry_key": "",
        "private_registry_ca": "",
        "cri_reinstall_if_already_install": false
    },
    "cni": {
        "cni_type": "calico",
        "pod_v4_cidr": "172.25.0.0/24",
        "pod_v6_cidr": "3001:db8::0/64",
        "calico": {
            "ipv4_lookup_method": "",
            "ipv6_lookup_method": "",
            "calico_mode": "overlay",
            "ipip_mode": "always",
            "enable_dual_stack": false
        },
        "flannel": {
            "pod_v4_cidr": "",
            "tunnel_interface_ip": ""
        },
        "multus": {
            "calico": {
                "pod_v4_cidr": "",
                "ipv4_lookup_method": "",
                "pod_v6_cidr": "",
                "ipv6_lookup_method": "",
                "calico_mode": "overlay",
                "ipip_mode": "always",
                "enable_dual_stack": false
            },
            "sriov_cni": {}
        }
    },
    "masters": [
        {
            "node-id": "client-MTkyLjE2OC4yLjE1NA=="
        },
                {
            "node-id": "client-MTkyLjE2OC4yLjE4MA=="
        },
        {
            "node-id": "client-MTkyLjE2OC4yLjQw"
        }
    ],
    "workers": [
        {
            "node-id": "client-MTkyLjE2OC4yLjE1NA=="
        }
    ],
    "ingresses": {
        "ingress-type": "nginx",
        "ingress-use-host-network": true,
        "ingress-nodes": [
            {
                "node-id": "client-MTkyLjE2OC4yLjE1NA=="
            },
            {
                "node-id": "client-MTkyLjE2OC4yLjE4MA=="
            },
                        {
                "node-id": "client-MTkyLjE2OC4yLjQw"
            }
        ]
    },
    "efk": {
            "enable": true,
            "namespace": "efk",
            "storage_class_name": "cinder-csi"
      },
    "cloud_providers":
        {
            "cloud_provider_type": "openstack",
            "cloud_provider_username": "wuwenxiang",
            "cloud_provider_password": "wuwenxiang",
            "auth_url": "http://test:5000/v3"
        },
    "ClusterOperationIDs": [
        "operation1"
    ]
}'


# 同时安装第二套集群

$ curl --location --request POST 'http://localhost:9090/cluster/v1' \
--header 'Content-Type: application/json' \
--header 'token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjEiLCJ1c2VybmFtZSI6InNpbW9uIiwicGFzc3dvcmQiOiIxMjMiLCJyb2xlcyI6bnVsbH0.TYuoImyKeHGHZqSivb309gqnwjeyDys2tbUb296ETG4' \
--data-raw '{
    "ClusterId": "cluster1",
    "cluster_name": "cluster1",
    "control_plane": {
        "service_v4_cidr": "10.96.0.0/16",
        "enable_ipvs": true
    },
    "container_runtime": {
        "container_runtime_type": "docker",
        "private_registry_address": "10.0.0.10",
        "private_registry_port": 4000,
        "private_registry_cert": "",
        "private_registry_key": "",
        "private_registry_ca": "",
        "cri_reinstall_if_already_install": false
    },
    "cni": {
        "cni_type": "calico",
        "pod_v4_cidr": "172.25.0.0/24",
        "pod_v6_cidr": "3001:db8::0/64",
        "calico": {
            "ipv4_lookup_method": "",
            "ipv6_lookup_method": "",
            "calico_mode": "overlay",
            "ipip_mode": "always",
            "enable_dual_stack": false
        },
        "flannel": {
            "pod_v4_cidr": "",
            "tunnel_interface_ip": ""
        },
        "multus": {
            "calico": {
                "pod_v4_cidr": "",
                "ipv4_lookup_method": "",
                "pod_v6_cidr": "",
                "ipv6_lookup_method": "",
                "calico_mode": "overlay",
                "ipip_mode": "always",
                "enable_dual_stack": false
            },
            "sriov_cni": {}
        }
    },
    "masters": [
        {
            "node-id": "client-MTkyLjE2OC4yLjgx"
        },
                {
            "node-id": "client-MTkyLjE2OC4yLjg1"
        },
        {
            "node-id": "client-MTkyLjE2OC4yLjc5"
        }
    ],
    "workers": [
        {
            "node-id": "client-MTkyLjE2OC4yLjgx"
        }
    ],
    "ingresses": {
        "ingress-type": "nginx",
        "ingress-use-host-network": true,
        "ingress-nodes": [
            {
                "node-id": "client-MTkyLjE2OC4yLjgx"
            },
            {
                "node-id": "client-MTkyLjE2OC4yLjg1"
            },
                        {
                "node-id": "client-MTkyLjE2OC4yLjc5"
            }
        ]
    },
    "efk": {
        "enable": true,
        "namespace": "efk",
        "storage_class_name": "cinder-csi"
    },
    "helm": {
            "enable": true,
            "helm_version": 3,
            },
    "gap": {
            "enable": true,
    },
    "cloud_providers":
        {
            "cloud_provider_type": "openstack",
            "cloud_provider_username": "wuwenxiang",
            "cloud_provider_password": "wuwenxiang",
            "auth_url": "http://test:5000/v3"
        },
    "ClusterOperationIDs": [
        "operation1"
    ]
}'

# 获取集群列表
curl --location \
--header 'token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjEiLCJ1c2VybmFtZSI6InNpbW9uIiwicGFzc3dvcmQiOiIxMjMiLCJyb2xlcyI6bnVsbH0.TYuoImyKeHGHZqSivb309gqnwjeyDys2tbUb296ETG4' \
 --request GET 'http://localhost:9090/cluster/v1'

# 获取指定集群信息
curl --location \
--header 'token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjEiLCJ1c2VybmFtZSI6InNpbW9uIiwicGFzc3dvcmQiOiIxMjMiLCJyb2xlcyI6bnVsbH0.TYuoImyKeHGHZqSivb309gqnwjeyDys2tbUb296ETG4' \
--request GET 'http://localhost:9090/cluster/v1/{cluster-id}'
```

### 检验安装

目前安装只实现了在线安装(要求能翻墙)和离线安装(需要准备rpm和镜像仓库)

```console
# 在 client 上运行
$ kubectl get nodes
# 输出
NAME                          STATUS   ROLES    AGE   VERSION
k8s-installer-client-test-1   Ready    master   45s   v1.18.5
k8s-installer-client-test-2   Ready    master   46s   v1.18.5
k8s-installer-client-test-3   Ready    master   74s   v1.18.5
```

## 关于 virtual kubelet 集成

目前 virtual kubelet 集成处于测试阶段，详细实现的设计文档请看 [caas-4.1实现virtual-kubelet的设计](http://172.16.30.21/openshift_origin/caas-documentation/blob/master/doc/mec-caas-5/addons/virtual_kubelet_implementation.md)

项目目前的进度:

1. 安装 virtual kubelet 节点在推集群的同时--已经完成
2. virtual kubelet caas provider--未完成

### 在安装集群时选择一个 worker 节点作为 virtual kubelet 节点

在安装集群时选择一个 worker 节点作为 virtual kubelet 节点的方法和安装集群没有任何区别，但是要注意的是，virtual kubelet 节点不允许是 master 节点，他只能是一个 worker 节点

```json
# 其他配置
    "workers": [
        {
            "node-id": "client-MTAuMC4wLjI0MA==",
            "use_virtual_kubelet": true,
            "virtual_kubelet": {
                "provider": "caas",
                "vk_provider_caas": {
                    "cpu_limit": "20",
                    "memory_limit": "100Gi",
                    "pod_limit": "20",
                    "caas_api_url": "self",
                    "caas_username": "admin",
                    "caas_password": "caas",
                    "caas_cluster_id": "self"
                }
            }
        }
    ],
# 其他配置
```

```cosnole
# 获取集群信息
$ kubectl get nodes
NAME                                                    STATUS   ROLES    AGE     VERSION
cluster-99ea622a-7d9e-46b8-b3fc-c7dad55e332a-master-0   Ready    master   3h13m   v1.18.6
cluster-99ea622a-7d9e-46b8-b3fc-c7dad55e332a-master-1   Ready    master   3h13m   v1.18.6
cluster-99ea622a-7d9e-46b8-b3fc-c7dad55e332a-master-2   Ready    master   3h13m   v1.18.6
cluster-99ea622a-7d9e-46b8-b3fc-c7dad55e332a-worker-0   Ready    agent    3h12m   v1.15.2-vk-v1.3.0-9-gbad60764-dev

# 创建 pod 在 vk 节点上
$ cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - image: nginx
    imagePullPolicy: Always
    name: nginx
    ports:
    - containerPort: 80
      name: http
      protocol: TCP
    - containerPort: 443
      name: https
  dnsPolicy: ClusterFirst
  nodeSelector:
    kubernetes.io/role: agent
    beta.kubernetes.io/os: linux
    type: virtual-kubelet
  tolerations:
  - key: virtual-kubelet.io/provider
    operator: Exists
  - key: azure.com/aci
    effect: NoSchedule
EOF

# 查看 pod
$ kubectl get pod
NAME    READY   STATUS    RESTARTS   AGE
nginx   1/1     Running   0          3h12m
```

## 动态添加节点

我们可以将未被其他集群占用的节点加入到集群当中，同时支持 virtual kubelet，直接请求 api, clusterId 代表要被添加节点的集群， body 中的参数表示要被添加的节点，需要注意的是要被添加节点的集群的状态必须要处在 cluster-running 状态方可操作，被添加的节点被需不能被其他节点占用

```console
# 其他配置
curl --location --request POST 'http://localhost:9090/cluster/v1/add_node/[clusterId]' \
--header 'token: [token]' \
--header 'Content-Type: application/json' \
--data-raw '[
    {
        "node-id": "client-MTAuMC4wLjI0MA=="
    },
    {
        "node-id": "client-MTAuMC4wLjIwMQ=="
    },
    {
        "node-id": "client-MTAuMC4wLjIwMg==", # go virtual kubelet node
        "use_virtual_kubelet": true,
        "virtual_kubelet": {
            "provider": "caas",
            "vk_provider_caas": {
                "cpu_limit": "20",
                "memory_limit": "100Gi",
                "pod_limit": "20",
                "caas_api_url": "self",
                "caas_username": "admin",
                "caas_password": "caas",
                "caas_cluster_id": "self"
            }
        }
    }
]'
# 其他配置
```

### 已知问题

三方库 [zcalusic/sysinfo](https://github.com/zcalusic/sysinfo) 在 commit [12f5bd936877453186ec046071a11799264eadc8](https://github.com/zcalusic/sysinfo/commit/12f5bd936877453186ec046071a11799264eadc8) 之前在 Linux 5.4+ 版本的内核上有严重 bug。
