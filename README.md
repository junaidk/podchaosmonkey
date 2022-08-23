## Description
Golang program which runs inside the K8s cluster, interacts with the
kube apiserver, and deletes one pod at random in a particular namespace on a schedule.

## Build
run `make build` to build binary package

binary will be placed in `build/bin/` dir

## Build Docker image
run `make image` to build container image.

## Deploy to K8s
update image name in `k8s/deployment` if needed.

update env variables to configure the programme behaviour.

run `make deploy-workload` to deploy test workload (nginx deployments)

then run `make deploy` to deploy the programme to k8s cluster.

## Program Parameters
Following parameters can be provided as Env or run time flags to change behaviour of the program.

```
LABELS      = label selector with kubectl label... syntax to select pods to delete.
NAME_SPACE  = namespace to monitor and delete pods (default "default").
SCHEDULE    = schedule to run the process, in duration string format e.g 10s, 1h (default "10s").
KUBE_CONFIG = path to kube config file, if empty will use in cluster config.
```
Program will use `MY_POD_NAME` env var to get its self pod name, so that it does not delete itself.

It will only delete pods which are in running state.

## Enhancement
- We can enhance the behaviour by adding ability to pass higher construct names like Deployment, Statefullset etc
and only delete pods for a specific deployment object.
- Add validation on runtime config.
- Add namespace allow/deny list to delete pods from.