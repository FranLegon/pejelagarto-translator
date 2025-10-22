//go:build obfuscated

package obfuscation

const (
	projectName  = "piper-server"
	scriptSuffix = "piper-get-requirements"
)

// ProjectName returns the project name
func ProjectName() string {
	return projectName
}

// ScriptSuffix returns the script suffix
func ScriptSuffix() string {
	return scriptSuffix
}

// ShouldOpenBrowser returns whether the browser should auto-open
func ShouldOpenBrowser() bool {
	return false
}

// Obfuscated returns whether this is an obfuscated build
func Obfuscated() bool {
	return true
}
