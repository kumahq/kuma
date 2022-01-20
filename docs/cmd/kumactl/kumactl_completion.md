## kumactl completion

Output shell completion code for bash, fish or zsh

### Synopsis


Outputs kumactl shell completion for the given shell (bash, fish or zsh).
This depends on the bash-completion package.  Example installation instructions:
# for bash users
	$ kumactl completion bash > ~/.kumactl-completion
	$ source ~/.kumactl-completion

# for zsh users
	% kumactl completion zsh > /usr/local/share/zsh/site-functions/_kumactl
	% autoload -U compinit && compinit
# or if zsh-completion is installed via homebrew
    % kumactl completion zsh > "${fpath[1]}/_kumactl"

# for fish users
	% kumactl completion fish > ~/.config/fish/completions/kumactl.fish

Additionally, you may want to output the completion to a file and source in your .bashrc
Note for zsh users: [1] zsh completions are only supported in versions of zsh >= 5.2


### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
      --api-timeout duration   the timeout for api calls. It includes connection time, any redirects, and reading the response body. A timeout of zero means no timeout (default 1m0s)
      --config-file string     path to the configuration file to use
      --log-level string       log level: one of off|info|debug (default "off")
  -m, --mesh string            mesh to use (default "default")
      --no-config              if set no config file and config directory will be created
```

### SEE ALSO

* [kumactl](kumactl.md)	 - Management tool for Kuma
* [kumactl completion bash](kumactl_completion_bash.md)	 - Output shell completions for bash
* [kumactl completion fish](kumactl_completion_fish.md)	 - Output shell completions for fish
* [kumactl completion zsh](kumactl_completion_zsh.md)	 - Output shell completions for zsh

