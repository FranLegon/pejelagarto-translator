//go:build !obfuscated

package obfuscation

const (
	projectName  = "pejelagarto-translator"
	scriptSuffix = "pejelagarto-get-requirements"
)

// ProjectName returns the project name
func ProjectName() string {
	return projectName
}

// ScriptSuffix returns the script suffix
func ScriptSuffix() string {
	return scriptSuffix
}
