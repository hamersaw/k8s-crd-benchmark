# k8s-crd-benchmark
## overview
This is meant as a informal benchmark for the k8s API server in regards to creating and modifying CRDs.

# usage
## installation
    k3d cluster create -p "30000:30000@server:0" -p "32000:32000@server:0" --image rancher/k3s:v1.21.8-k3s2 -v /dev/mapper/nvme0n1p3_crypt:/dev/mapper/nvme0n1p3_crypt flyte

    kubectl create -f k8s/service-account.yaml
    kubectl create -f k8s/prometheus.yaml
    kubectl create -f k8s/grafana.yaml

# resources
## deploying prometheus / grafana on k8s
https://sysdig.com/blog/monitor-kubernetes-api-server/
https://medium.com/@gurpreets0610/deploy-prometheus-grafana-on-kubernetes-cluster-e8395cc16f91
https://blog.freshtracks.io/a-deep-dive-into-kubernetes-metrics-part-4-the-kubernetes-api-server-72f1e1210770
[metrics overview](https://docs.datadoghq.com/integrations/kube_apiserver_metrics/)
[grafana board](https://grafana.com/grafana/dashboards/15761)
## update vs patch
https://blog.atomist.com/kubernetes-apply-replace-patch/

# benchmarks
                    50
patch single        8.257, 8.261, 8.235
patch all           0.067, 0.102, 0.039
update single       8.266, 8.238, 8.239
update all          0.055, 0.042, 0.038
