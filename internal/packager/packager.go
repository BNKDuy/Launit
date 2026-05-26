package packager

import "fmt"

type Packager interface {
	BuildAndPackage(sourcePath string) (zipPath string, err error)
}

func New(runtime string) (Packager, error) {
	switch runtime {
	case "go":
		return &GoPackager{}, nil
	default:
		return nil, fmt.Errorf("unsupported runtime: %s", runtime)
	}
}
