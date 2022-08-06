package api

type Config struct {
	AppName, AppVersion string
	DataPath            string //base data directory of zero trust services
	ApiPort             int
	Local               bool //if set, to bind the api to localhost:port (electron) or simply :port (web service) instead
}

type MonitorOptions struct {
	ProjectIDs []string
}

type SocketEndMessage struct {
	Message string
}
