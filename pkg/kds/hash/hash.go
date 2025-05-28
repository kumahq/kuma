package hash

import (
	"encoding/hex"
	"fmt"
	"hash/fnv"

	"k8s.io/apimachinery/pkg/util/rand"
	apimachineryvalidation "k8s.io/apimachinery/pkg/util/validation"
	k8s_strings "k8s.io/utils/strings"
)

type HashedNameFn func(mesh, name string, opts ...Option) string

const (
	hashLength = 1 + 16 // 1 dash plus 8-byte hash that is encoded with hex

	K8sNameLengthLimit253 = apimachineryvalidation.DNS1123SubdomainMaxLength
	K8sNameLengthLimit63  = apimachineryvalidation.DNS1035LabelMaxLength
)

type options struct {
	nameLengthLimit        int
	additionalValuesToHash []string
}

type Option func(*options)

func WithNameLengthLimit(limit int) Option {
	return func(o *options) {
		o.nameLengthLimit = limit
	}
}

func WithAdditionalValuesToHash(values ...string) Option {
	return func(o *options) {
		o.additionalValuesToHash = append(o.additionalValuesToHash, values...)
	}
}

func HashedName(mesh, name string, opts ...Option) string {
	o := &options{
		nameLengthLimit: K8sNameLengthLimit253,
	}

	for _, opt := range opts {
		opt(o)
	}

	return addSuffix(name, hash(append([]string{mesh, name}, o.additionalValuesToHash...)), o.nameLengthLimit)
}

func addSuffix(name, hash string, nameLengthLimit int) string {
	shortenName := k8s_strings.ShortenString(name, nameLengthLimit-hashLength)
	return fmt.Sprintf("%s-%s", shortenName, hash)
}

func hash(ss []string) string {
	hasher := fnv.New64a()
	for _, s := range ss {
		_, _ = hasher.Write([]byte(s))
	}
	b := []byte{}
	b = hasher.Sum(b)

	return rand.SafeEncodeString(hex.EncodeToString(b))
}

func DoNothingHashedName(mesh, name string, opts ...Option) string {
	return name
}
