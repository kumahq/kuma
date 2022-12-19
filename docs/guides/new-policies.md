## How to generate a new Kuma policy

Use the tool:

```shell
go run ./tools/policy-gen/bootstrap/... --name CaseNameOfPolicy
```

The output of the tool will tell you where the important files are!

## Add plugin name to the configuration

Enabled plugin configuration is in `pkg/plugins/policies/policies.go`. Plugins name is equals to `KumactlArg` in file `zz_generated.resource.go`. It's important to place the plugin in the correct place because the order of executions is important.
