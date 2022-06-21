package plugable

type IPlugAble interface {
	IsEnable() bool
	GetName() string
	GetStatus() string
	SetDependencies(map[string]IPlugAble)
	GetDependencies() map[string]IPlugAble
	CheckDependencies() (error,[]string)
	GetLicenseLabel() uint16
}