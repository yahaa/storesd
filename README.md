### storesd
thanos query file store sd sidecar

### 功能

 以 sidecar 方式与 thanos query 在同一个 pod 中通过共享 emptyDir 目录协同工作，为 thanos query 定时更新 `--store.sd-files=/etc/sd/stores.yaml` 文件，从而实现 thanos query 跨集群的服务发现

`注: 需要多个 k8s 集群 pod 网络互通`

### 配置文件

```yaml
outputPath: /etc/sd
srvAddr: :8081
syncTargets:
  - kubeConfigPath: /etc/kube/config-qa
    cluster: cluster-qa
    namespace: monitoring
	services: 
    - service: thanos-store-gateway
      portName: grpc
```

### 输出文件

文件格式参考 [prometheus file sd](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#file_sd_config)
```yaml
- targets:
  - 10.224.5.138:10901
  - 10.224.9.76:10901
  - 10.224.3.86:10901
  - 10.224.9.206:10901
```

