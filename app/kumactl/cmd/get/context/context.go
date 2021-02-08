package context

type GetContext struct {
	Args struct {
		OutputFormat string
	}
}

type ListContext struct {
	Args struct {
		Size   int
		Offset string
	}
}
