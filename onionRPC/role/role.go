package role

type RoleConfig struct {
	ListenAddr        string
	TracingServerAddr string
	TracingIdentity   string
}

type Role struct {
	SessionKeys map[string]string
	FatalError  chan error
}

func InitRole() Role {
	return Role{
		SessionKeys: make(map[string]string),
		FatalError:  make(chan error),
	}
}
