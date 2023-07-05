## ProxyTemplate

- `selectors` (required, repeated)

    List of Dataplane selectors.

- `conf` (required)

    Configuration for ProxyTemplate

    Child properties:    
    
    - `imports` (optional, repeated)
    
        List of imported profiles.
        +optional    
    
    - `resources` (optional, repeated)
    
        List of raw xDS resources.
        +optional    
    
    - `modifications` (optional, repeated)
    
        List of config modifications
## ProxyTemplateSource

- `name` (optional)

    Name of a configuration source.
    +optional

- `profile` (optional)

    Profile, e.g. `default-proxy`.
    +optional

    Child properties:    
    
    - `name` (optional)
    
        Profile name.    
    
    - `params` (optional)
    
        Profile params if any.
        +optional

- `raw` (optional)

    Raw xDS resources.
    +optional

    Child properties:    
    
    - `resources` (optional, repeated)
    
        List of raw xDS resources.
        +optional
## ProxyTemplateProfileSource

- `name` (optional)

    Profile name.

- `params` (optional)

    Profile params if any.
    +optional
## ProxyTemplateRawSource

- `resources` (optional, repeated)

    List of raw xDS resources.
    +optional
## ProxyTemplateRawResource

- `name` (required)

    The resource's name, to distinguish it from others of the same type of
    resource.

- `version` (required)

    The resource level version. It allows xDS to track the state of individual
    resources.

- `resource` (required)

    xDS resource.

