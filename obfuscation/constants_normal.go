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

// ShouldOpenBrowser returns whether the browser should auto-open
func ShouldOpenBrowser() bool {
	return true
}

// NotObfuscated returns whether this is a non-obfuscated build
func NotObfuscated() bool {
	return true
}
