package dep

type (
	DepPackage map[string]string
	DepArch    map[string]DepPackage
	DepVersion map[string]DepArch
	DepMap     map[string]DepVersion
)
