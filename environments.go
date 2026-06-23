package manager

// Environment selects a named babelforce host. Target other (e.g. per-customer or non-production)
// hosts with Options.BaseURL.
type Environment int

const (
	// Production targets https://services.babelforce.com.
	Production Environment = iota
)

func resolveBaseURL(Environment) string {
	return "https://services.babelforce.com"
}
