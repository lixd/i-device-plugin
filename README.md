# i-device-plugin
k8s device-plugin demo.

* i-device-plugin 将会新增资源 `lixueduan.com/gopher`。
* 将会扫描获取 `/etc/gophers` 目录下的文件作为对应的设备。
* 将设备分配给 Pod 后，会在 Pod 中新增环境变量`Gopher=$deviceId`


## 构建镜像
```bash
make build-image
```

## 部署
使用 DaemonSet 来部署 `i-device-plugin`，以便其能运行到集群中的所有节点上。

```bash
kubectl apply -f deploy/daemonset.yaml
```
检测 Pod 运行情况

```bash
[root@test ~]# kubectl -n kube-system get po
i-device-plugin-vnw6z            1/1     Running   0          17s
```

## 测试

新增设备,在该 Demo 中，把 /etc/gophers 目录下的文件作为设备，因此我们只需要到 /etc/gophers 目录下创建文件，模拟有新的设备接入即可。
```bash
mkdir /etc/gophers

touch /etc/gophers/g1
```
查看 device plugin pod 日志,可以正常感知到设备
```bash
I0719 14:01:00.308599       1 device_monitor.go:70] fsnotify device event: /etc/gophers/g1 CREATE
I0719 14:01:00.308986       1 device_monitor.go:79] find new device [g1]
I0719 14:01:00.309017       1 device_monitor.go:70] fsnotify device event: /etc/gophers/g1 CHMOD
I0719 14:01:00.309141       1 api.go:32] device update,new device list [g1]
```
查看 node capacity 信息，能够看到新增的资源
```bash
[root@test ~]# kubectl get node argo-1 -oyaml|grep  capacity -A 7
  capacity:
    cpu: "4"
    ephemeral-storage: 20960236Ki
    hugepages-1Gi: "0"
    hugepages-2Mi: "0"
    lixueduan.com/gopher: "1"
    memory: 8154984Ki
    pods: "110"
```

创建 Pod 申请该资源
```bash
kubectl apply -f deploy/test-pod.yaml
```
Pod 启动成功

```bash
[root@test ~]# kubectl get po
NAME         READY   STATUS    RESTARTS   AGE
gopher-pod   1/1     Running   0          27s
```

之前分配设备是添加 Gopher=xxx  这个环境变量，现在看下是否正常分配

```bash
[root@test ~]# kubectl exec -it gopher-pod -- env|grep Gopher
Gopher=g1
```

ok,环境变量存在，可以看到分配给该 Pod 的设备是 g1。