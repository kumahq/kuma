package vips

type List map[string]string

type ChangeHandler func(list List)

func (vips List) Append(other List) {
	for k, v := range other {
		vips[k] = v
	}
}
