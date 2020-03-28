package project

var (
	description = "A simple JSON logs parser for Kubernetes Operators designed for effective debugging."
	gitSHA      = "n/a"
	name        = "gg"
	source      = "https://github.com/giantswarm/gg"
	version     = "n/a"
)

func Description() string {
	return description
}

func GitSHA() string {
	return gitSHA
}

func Name() string {
	return name
}

func Source() string {
	return source
}

func Version() string {
	return version
}
