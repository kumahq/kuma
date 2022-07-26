# Support additional Retry options from Envoy

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/4494

## Context and Problem Statement

We currently set some defaults for the retryOn field for HTTP in our Retry policy. Some users may need additional flexibility so we should plumb through some extra fields from Envoy.

## Decision Drivers

* User request (additional flexibility)

## Considered Options

1. ProxyTemplate
2. Adding fields to Retry policy

## Decision Outcome

Chosen option: 2 because we want to provide a lower friction experience to add these fields.

### Policy Changes

The change the policy would add a `retryOn` field to expose additional configuration options beyond our defaults.

```yaml
spec:
  conf:
    http:
      backOff:
        baseInterval: 0.025s
        maxInterval: 0.250s
      numRetries: 5
      perTryTimeout: 16s
      retryOn:
      - all_5xx
      - gateway_error
```

### Changes to default / existing behavior

- This addition _would not_ change existing behavior in the case that the `http.retriableStatusCodes` field _is not_ specified.
- This addition _would_ change the existing behavior in the case that the `http.retriableStatusCodes` field _is_ specified. Previously specifying this field would set the `retryOn` field to a default of `"connect-failure,refused-stream,retriable-status-codes"`. Following this change we will add _only_ `retriable-status-codes` (as specifying them doesn't make sense otherwise), and leave the additional conditions to be specified by the user (which would be possible now) if desired.

### Potential Implementation Notes

- Due to this field being an enum, we cannot use `5xx` as a field name, so must prefix it (`all_5xx` above) then post 'convert' to the Envoy-native field name.
- We already use `retriable_status_codes` as a field to specify which codes should be retried, however that is also one of the enum options in `retryOn`. We also need to have a different field name here and post-process back to the Envoy-native field name.

### Positive Consequences <!-- optional -->

* Improve flexibility for configuring Retry policies.

### Negative Consequences <!-- optional -->

* Additional complexity of policy?