# How to generate a new Kuma policy

1. Create a new directory for the policy in `pkg/plugins/policies`. Example:
    ```shell
    mkdir -p pkg/plugins/policies/donothingpolicy
    ```

2. Create a proto file for new policy in `pkg/plugins/policies/donothingpolicy/api/v1alpha1`. For example 
donothingpolicy.proto:
    ```protobuf
    syntax = "proto3";

    package kuma.plugins.policies.donothingpolicy.v1alpha1;
    
    import "mesh/options.proto";
    option go_package = "github.com/kumahq/kuma/pkg/plugins/policies/donothingpolicy/api/v1alpha1";
    
    import "mesh/v1alpha1/selector.proto";
    import "config.proto";
    
    option (doc.config) = {
    type : Policy,
    name : "DoNothingPolicy",
    file_name : "donothingpolicy"
    };
    
    // DoNothingPolicy defines permission for traffic between dataplanes.
    message DoNothingPolicy {
    
    option (kuma.mesh.resource).name = "DoNothingPolicyResource";
    option (kuma.mesh.resource).type = "DoNothingPolicy";
    option (kuma.mesh.resource).package = "mesh";
    option (kuma.mesh.resource).kds.send_to_zone = true;
    option (kuma.mesh.resource).ws.name = "donothingpolicy";
    option (kuma.mesh.resource).ws.plural = "donothingpolicies";
    option (kuma.mesh.resource).allow_to_inspect = true;
    
    // List of selectors to match dataplanes that are sources of traffic.
    repeated kuma.mesh.v1alpha1.Selector sources = 1 [ (doc.required) = true ];
    // List of selectors to match services that are destinations of traffic.
    repeated kuma.mesh.v1alpha1.Selector destinations = 2 [ (doc.required) = true ];
   
    message Conf {
      bool enableDoNothing = 1;
    }
    
    Conf conf = 3;

    }
    ```

3. Call `make generate/policy/<POLICY_NAME>`. Example:
   ```shell
   make generate/policy/donothingpolicy
   ```
   
4. Add import to `pkg/core/bootstrap/plugins.go`:
   ```go
   _ "github.com/kumahq/kuma/pkg/plugins/policies/donothingpolicy"
   ```

5. Update Helm chart with a new CRD:
   ```shell
   make generate/helm/donothingpolicy
   ```
   Also, today it's required to update `cp-rbac.yaml` manually, automation is yet to come.