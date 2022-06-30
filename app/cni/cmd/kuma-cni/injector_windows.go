package main

import "errors"

func Inject(netns string, intermediateConfig *IntermediateConfig) error {
	return errors.New("only implemented on linux")
}
