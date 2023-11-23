# Deprecate standalone mode

* Status: accepted

## Context and Problem Statement

We want to simplify our deployments model to only have Zone CP and Global CP.
The goal is not to get rid of standalone altogether, but to change the naming.

## Decision Drivers

* Simplify UX for (new) users. Simpler docs. Fewer things to wrap your head around.
* Provide better migration path from standalone to multizone, as standalone won't exist.

## Considered Options

* Deprecate standalone

## Decision Outcome

Chosen option: "Deprecate standalone".

## Pros and Cons of the Options

### Deprecate standalone

If you want to run the "standalone" you can execute the following command
```
KUMA_MODE=zone kuma-cp run 
```
We will change the default `KUMA_MODE` value to `zone`. So in the end, you would run this the same way as before which is
```
kuma-cp run
```
HELM and kumactl install adds `KUMA_MODE=standalone` by default, we can also change the default to `zone`.

The current way of running Zone CP will not change.

The current difference of running standalone and zone is providing zone name and KDS address.

Current differences between standalone and zone CP:
* Cannot apply policies on Zone CP directly.
* GUI is disabled on Zone CP
* Cluster ID is synced from Global on Zone CP. Cluster ID is used for analytics
* Zone CP connects to Global CP to pull policies
* Default resources (Mesh, policies etc.) are disabled on Zone CP, because those are pulled from Global.
* Standalone does not finalize (gc/finalizer) zone ingresses
* Mesh/Service Insights are not computed on Zone CP
* GatewayAPI is not supported on Zone CP
* On Zone CP, Zone Token public key is synced from Global.

With deprecating standalone we still need to differentiate between zone running on its own and zone connected to the Global CP.
To do this, instead of relying on `KUMA_MODE` `standalone` or `zone`, we will check if KDS address was provided.

Existing standalone and future Zone CP without KDS connection will work the same way.

Changes between existing Zone CP and upcoming Zone CP with KDS connection:
* You can access GUI

#### Visibility changes

Currently, Zone tab is missing in "standalone" GUI.
We create Zone object and ZoneInsights with statistic of KDS only on Global CP. Those objects are not synced to Zone CP.
We want to also create Zone object for Zone CP and create ZoneInsights with KDS statistics from Zone perspective.
Then we can show on this screen whether the Zone is connected to Global or not.

#### Transitioning from Zone CP to Zone CP with KDS connection

Let's assume we first start using Kuma with Zone CP without KDS connection.

Then we switch the zone to connect to the Global CP. What will happen is:
* Policies that are synced from Global CP are replaced on Zone CP
  It's important to copy policies from Zone CP to Global CP first.
  However, this will be easier once we support applying policies on Zone CP, because instead of replacing policies from Global, we will just sync existing policies to Global.
* We block applying policies on Zone CP. This will change once we support apply policies on Zone CP.
* Secrets that are synced from Global CP are replaced on Zone CP
  It's important to copy secrets from Zone CP to Global CP first.
  This includes signing keys and CAs.
* You can no longer issue zone tokens on Zone CP.
* You can no longer apply Gateway API resources. All Kuma resources created from GAPI are replaced
  This is the case until we implement applying policies on Zone CP.

Ideally we should create a tool to save relevant policies from Zone CP and apply them to Global CP.
This however won't be a part of initial implementation of this MADR.
