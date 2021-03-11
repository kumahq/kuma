package vips

type List map[string]string

func (vips List) Append(other List) {
	for k, v := range other {
		vips[k] = v
	}
}
