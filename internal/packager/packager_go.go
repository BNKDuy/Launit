package packager

type GoPackager struct {
}

var _ (Packager) = (*GoPackager)(nil)

func (p *GoPackager) BuildAndPackage(sourcePath string) (zipPath string, err error) {
	return "", nil
}
