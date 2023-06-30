package scopes

type HasScopeRequirements interface {
	RequiredScopes() []string
}
