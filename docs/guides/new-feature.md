# New feature

## Design

If this is a bigger feature, consider writing a proposal, so we can avoid a situation that PR needs to be reimplemented.

When designing a new feature, consider the following questions

* How does it work on Kubernetes?
* How does it work on Universal?
* How does it work with multi-zone (hybrid) deployment?
* Is it backwards compatible?
* Does it affect projects that are based on Kuma?

## Implementation

* If possible, write E2E test that verifies the behavior
* If a new behavior is introduced or the old is changed, write [docs](https://github.com/kumahq/kuma-website/).
* Does it need changes in the [GUI](https://github.com/kumahq/kuma-gui)?

## Deployment

* How is deployed on Universal?
* How is deployed on Kubernetes with kumactl?
* How is deployed on Kubernetes with HELM?
