package output

type Format string

const (
	TableFormat Format = "table"
	YAMLFormat  Format = "yaml"
	JSONFormat  Format = "json"
)
