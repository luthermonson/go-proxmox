package types

type Routes []Route
type Route struct {
	MatchHeader map[string]string
	MatchType   string
	Method      string
	Path        string
	Reply       int
	BodyString  string
	JSON        interface{}
}

type Config struct {
	TestURI        string
	Node           string
	VirtualMachine string
	Version        func(Config)
}
