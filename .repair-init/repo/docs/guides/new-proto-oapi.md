## How to add new proto resource to openapi schema

### Manual

1. Add the type in `AdditionalProtoTypes` in [tools/resource-gen/pkg/generator/main.go](https://github.com/kumahq/kuma/blob/e279bba108307837bbe966f032c5e95039759372/tools/resource-gen/pkg/generator/main.go#L462)
2. Run `make generate/oas docs/generated/openapi.yaml`
3. Commit the changes

### Recording

Zoom clip: https://konghq.zoom.us/rec/share/GhqZq9VM4cnKc_DCsGSiZ_6BbwFFmMQzyWvlpHhc5qPATZrYasOYx1-7B-rYmglU.k1YpevHxpOeTxqjO
Passcode: z=KN36.6

### Example

You can look at these PRs to see how it works:
- https://github.com/kumahq/kuma/pull/13757/files
- https://github.com/kumahq/kuma/pull/13479/files#diff-67e3391f5bf7ad9d0f014ebe4a1efa7789da467fa233708bbbbe019ed12162bfR454 
