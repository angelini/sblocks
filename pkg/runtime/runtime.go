package runtime

import (
	"github.com/angelini/sblocks/pkg/core"
)

type Runtime struct {
	Name     string
	Public   bool
	FreeSize int
	Labels   map[string]string
	Revision core.Revision
}

func NewRuntime(name string, public bool, freeSize int, labels map[string]string, revision core.Revision) Runtime {
	labels["sb_runtime"] = name
	return Runtime{
		Name:     name,
		Public:   public,
		FreeSize: freeSize,
		Labels:   labels,
		Revision: revision,
	}
}
