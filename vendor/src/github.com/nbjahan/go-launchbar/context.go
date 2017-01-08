package launchbar

import "log"

// Context is a dependency that is available in Matcher, Runner, Renderer func
type Context struct {
	Action *Action     // points to the LaunchBar action
	Config *Config     // the Config object
	Cache  *Cache      // the Cache object
	Self   *Item       // the item that is accessing the context
	Input  *Input      // the user input
	Logger *log.Logger // Logger is used to log to Action.SupportPath() + '/error.log'
}
