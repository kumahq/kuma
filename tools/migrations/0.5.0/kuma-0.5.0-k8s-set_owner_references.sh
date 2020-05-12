#!/usr/bin/env bash

function patchResources {
    tuplesParam=("$@")

    for i in "${tuplesParam[@]}" 
    do
    :
        arr=($i)
        name="${arr[0]}"
        namespace="${arr[1]}"
        mesh="${arr[2]}"
        kind="${arr[3]}"
        meshUID=$(kubectl get mesh $mesh -o jsonpath='{.metadata.uid}')
        echo name=$name ns=$namespace mesh=$mesh kind=$kind meshUID=$meshUID


        patch=$(cat<<EOF
[
    {
        "op": "add",
        "path": "/metadata/ownerReferences",
        "value": [
            {
            "apiVersion": "kuma.io/v1alpha1",
            "blockOwnerDeletion": true,
            "controller": true,
            "kind": "Mesh",
            "name": "$mesh",
            "uid": "$meshUID"
            }
        ]
    }
]
EOF
)
        kubectl patch $kind $name -n$namespace --type='json' -p "$patch"
    done
}

SAVEIFS=$IFS   # Save current IFS
IFS=$'\n'      # Change IFS to new line
faultinjections=($(kubectl get faultinjections.kuma.io -A -o custom-columns=Name:.metadata.name,Namespace:.metadata.namespace,Mesh:.mesh,Kind:.kind | awk 'NR>1{print $1,$2,$3,$4}'))
healthchecks=($(kubectl get healthchecks.kuma.io -A -o custom-columns=Name:.metadata.name,Namespace:.metadata.namespace,Mesh:.mesh,Kind:.kind | awk 'NR>1{print $1,$2,$3,$4}'))
proxytemplates=($(kubectl get proxytemplates.kuma.io -A -o custom-columns=Name:.metadata.name,Namespace:.metadata.namespace,Mesh:.mesh,Kind:.kind | awk 'NR>1{print $1,$2,$3,$4}'))
trafficlogs=($(kubectl get trafficlogs.kuma.io -A -o custom-columns=Name:.metadata.name,Namespace:.metadata.namespace,Mesh:.mesh,Kind:.kind | awk 'NR>1{print $1,$2,$3,$4}'))
trafficpermissions=($(kubectl get trafficpermissions.kuma.io -A -o custom-columns=Name:.metadata.name,Namespace:.metadata.namespace,Mesh:.mesh,Kind:.kind | awk 'NR>1{print $1,$2,$3,$4}'))
trafficroutes=($(kubectl get trafficroutes.kuma.io -A -o custom-columns=Name:.metadata.name,Namespace:.metadata.namespace,Mesh:.mesh,Kind:.kind | awk 'NR>1{print $1,$2,$3,$4}'))
traffictraces=($(kubectl get traffictraces.kuma.io -A -o custom-columns=Name:.metadata.name,Namespace:.metadata.namespace,Mesh:.mesh,Kind:.kind | awk 'NR>1{print $1,$2,$3,$4}'))
IFS=$SAVEIFS

patchResources "${faultinjections[@]}"
patchResources "${healthchecks[@]}"
patchResources "${proxytemplates[@]}"
patchResources "${trafficlogs[@]}"
patchResources "${trafficpermissions[@]}"
patchResources "${trafficroutes[@]}"
patchResources "${traffictraces[@]}"

# Delete DataplaneInsights
kubectl delete dataplaneinsights --all -A
