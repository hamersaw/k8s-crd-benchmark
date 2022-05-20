# usage
## installation
    k3d cluster create -p "30000:30000@server:0" -p "32000:32000@server:0" --image rancher/k3s:v1.21.8-k3s2 -v /dev/mapper/nvme0n1p3_crypt:/dev/mapper/nvme0n1p3_crypt flyte

    kubectl create -f k8s/service-account.yaml
    kubectl create -f k8s/prometheus.yaml
    kubectl create -f k8s/grafana.yaml

# update vs patch
https://blog.atomist.com/kubernetes-apply-replace-patch/

                    50
patch single        8.257, 8.261, 8.235
patch all           0.067, 0.102, 0.039
update single       8.266, 8.238, 8.239
update all          0.055, 0.042, 0.038



sum(rate(rest_client_request_duration_seconds_bucket{job="apiserver"}[$__rate_interval])) by (verb)
sum(rate(rest_client_request_duration_seconds_bucket{job="apiserver"}[$__rate_interval])) by (verb)
sum(rate(apiserver_request_duration_seconds_bucket{job="apiserver"}[$__rate_interval])) by (verb)

https://grafana.com/grafana/dashboards/15761


[better k3d setup?](https://www.linkedin.com/pulse/setup-your-personal-kubernetes-cluster-k3s-k3d-suren-raju)

    sum(rate(apiserver_request_duration_seconds_bucket{resource="flyteworkflows"}[1m])) by (verb)

    apiserver_current_inqueue_requests
    apiserver_current_inflight_requests
