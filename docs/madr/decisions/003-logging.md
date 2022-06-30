# Logging standard

* Status: accepted

## Context and Problem Statement

Seamless troubleshooting is a feature. We should keep more attention to the logging in Kuma.

## Considered Options

* Adopting consistent logging standard.

## Decision Outcome

Chosen option: "Adopting consistent logging standard", because it's the only option.

### Positive Consequences

* Better UX for users.
* Better troubleshooting means both users and maintainers spend less time debugging issues.

### Negative Consequences

* More effort to try to follow the guidelines.

## Pros and Cons of the Options

### Adopting consistent logging standard

#### Convention

Ideally, all those points are linted.

1) Logs are lowercase. Can be uppercase when you start with proper noun.
_GOOD_
```go
log.Info("something has happened")
log.Info("Envoy is starting")
```
_BAD_
```go
log.Info("Something has happened")
```

2) Do not include `.` at the end of the sentence if this is the only sentence in the log
_GOOD_
```go
log.Info("something has happened")
log.Info("something has happened. Restarting the process")
```
_BAD_
```go
log.Info("something has happened.")
```

3) Either log or return an error. Don't do both, unless you want to log more details on DEBUG level.
_GOOD_
```go
if err != nil {
	return errors.Wrap(err, "more info")
}
...
if err != nil {
	log.Error(err, "more info")
	return nil
}
...
if err != nil {
	log.V(1).Info("much more details about the error")
    return errors.Wrap(err, "more info")
}
```
_BAD_
```go
if err != nil {
    log.Error(err, "more info")
	return errors.Wrap(err, "more info")
}
```

4) Use constant message and structured logging

The first argument to the logger should be a constant message.

_GOOD_
```go
log.Info("dp connected", "name", dpName)
```
_BAD_
```go
log.Info(fmt.Sprintf("dp %s connected", dpName))
log.Info(fn(dpName)))
```

5) Carry the same context to multiple logs using WithValues
_GOOD_
```go
log := logger.WithValues("proxy", dpName, "mesh", mesh)
log.Info("connected")
log.Info("reconciled")
log.Info("disconnected")
```
_BAD_
```go
logger.Info("connected", proxy", dpName, "mesh", mesh)
logger.Info("reconciled", proxy", dpName, "mesh", mesh)
logger.Info("disconnected", proxy", dpName, "mesh", mesh)
```

6) Be mindful about logger name. Try to not repeat the information.
_GOOD_
```go
var log = core.Log.WithName("xds-server")
log.Info("starting")
```
_BAD_
```go
var log = core.Log.WithName("xds-server")
log.Info("starting XDS server")
```

7) Don't use ellipsis
_GOOD_
```go
log.Info("starting")
```
_BAD_
```go
log.Info("starting...")
```

#### Levels

We have 3 log levels:
* `ERROR` used like that `log.Error(err, "msg")`
* `INFO` used like that `log.Info("msg")`
* `DEBUG` used like that `log.V(1).Info("msg")`
Logging on higher levels is redundant since the log will end up in `DEBUG` level anyway.

What to log and how to pick a proper level for logs.

**ERROR**
ERROR should inform about failures of the system.

_Examples:_
* XDS reconciliation failed.
* There is connection error to Postgres / Kubernetes API.

Ideally, ERROR should be used for issues that require immediate attention.
If you execute an action that can potentially fail, and you add retries,
consider logging retry attempts on INFO and ERROR only after all retries failed.

These logs are meant to be read by the user so try to make it human-readable.
It should include both human-readable message and details for engineers to understand the cause of the problem.  

**INFO**
INFO should inform about significant events in the system.
Significant events usually change the state of the system or other systems.
_Examples:_
* The resource is updated
* Server is started / terminated
* A component is started / stopped
* Leader is picked / dropped
* Sidecar is injected
* Proxy joined the mesh
* Proxy is reconfigured

These logs are meant to be read by the user so try to make it human-readable.
It should not be required to be familiar with Kuma implementation to understand the INFO logs.

_GOOD_
```go
log.Info("proxy reconfigured", ...)
```
_BAD_
```go
log.Info("OnStreamResponse", ...)
```

**DEBUG**
DEBUG should print information that can help troubleshoot the issue.
Here we can log actions that do not modify the system.
_Examples:_
* GET request to the API Server
* We generated the Envoy configuration, but it's the same so nothing significant happens
* We triggered XDS generation because resources have changed, but it does not mean that proxy will be reconfigured.
* Reconciliation process is run on interval or event. User should not care until it actually does something important.
* `else conditions`. For example, "pod already has Kuma sidecar. Skipping the injection"
* Content of the update. 
  For example, on the INFO we may log that "user X modified resource Y", but logging the whole resource is heavy.
  We can log the content of the resource on DEBUG level.

These logs are meant to be read by Kuma maintainers and users that are ok with looking at Kuma implementation to understand them.

**WARNING**
We don't have a warning level in Kuma.
Warning could be a level to inform a user about potential misuse of the system like:
* Using deprecated options
* Warning about usage of autogenerated self-signed certs.

If you really need this level, prefix INFO level with `[WARNING]`.
Consider other mechanism of informing a user about it, it's easy to skip the log by accident.

When still in doubt:
* log on INFO but with separate logger name, so the user may disable this log.
  _Note: controlling the log level on logger name is not yet available._
* emit a metric with generalized information, so even if user disable the log, they still have some data.

#### Log or emit a metric

Metrics and logs are closely related, yet they have different characteristics. 
Logs give you the best precision. You know exactly what happened at specific time.
Logs usually give you more information than metrics. For example "User X accessed endpoint /xyz".
Metrics are more generalized so "Someone accessed endpoint /xyz".
In majority of cases, (structured) logging is superior, but it comes at cost of maintain the infrastructure to collect them.

For example, you may want to log every GET request to your server.
With systems that take a significant load, you may easily DDoS your logging system.
With metrics, you can increase counters and let the scraper fetch them every X second.

If you want to introduce a log that can be noisy, consider adding a generalized metric.
This way even if the user disables this log, they still have some data to troubleshoot. 

#### Global variable or pass the logger

In Kuma, we can either declare logger globally in the package
```go
var logger = core.Log.WithName("xds")
```
or pass it explicitly to the component
```go
NewComponent(core.Log.WithName("xds"))
```

The library code should ALWAYS receive the logger. The user of the library should decide whether they want to log or not.
With the current Kuma codebase, the main library packages are `pkg/core` and `pkg/util`.
For the rest, rely on your own judgement whether the code can be reused or not.
