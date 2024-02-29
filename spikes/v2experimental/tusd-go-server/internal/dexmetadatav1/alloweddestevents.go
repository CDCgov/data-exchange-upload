package dexmetadatav1

// TODO: line-up names with json names `json`

type CopyTarget struct {
	Target string 
} // .copyTarget

type ExtEvents struct {
	Name string
	DefinitionFileName string
	CopyTargets []CopyTarget
} // .extEvents

type AllowedDestAndEvents struct {
	destinationId string
	extEvents ExtEvents
} // .AllowedDestAndEvents