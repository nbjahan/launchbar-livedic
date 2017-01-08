package launchbar

import "encoding/json"

// Func represents a generic type to pass functions.
type Func interface{}

// FuncMap represents a predefined map of functions to execute with Item.Run
type FuncMap map[string]Func

// AlwaysMatch is a Matcher func that always returns true.
var AlwasMatch = func() bool { return true }

// NeverMatch is a Matcher func that always returns false.
var NeverMatch = func() bool { return false }

// MatchIfTrueFunc is Matcher func that returns true if a value of passed argument is true.
var MatchIfTrueFunc = func(b bool) func() bool { return func() bool { return b } }

// MatchIfFalseFunc is Matcher func that returns true a value of passes argument is false.
var MatchIfFalseFunc = func(b bool) func() bool { return func() bool { return !b } }

// ShowViewFunc is a Runner func that shows the specified view.
var ShowViewFunc = func(v string) func(*Context) { return func(c *Context) { c.Action.ShowView(v) } }

// Item represents the LaunchBar item
type Item struct {
	View     *View
	item     *item
	match    Func // Matcher func
	run      Func // Runner func
	render   Func // Renderer func
	children []item
}

// NewItem initialize and returns a new Item
func NewItem(title string) *Item {
	return &Item{
		item: &item{
			Title: title,
			Data:  make(map[string]interface{}),
		},
	}
}

func newItem(item *item) *Item {
	return &Item{item: item}
}

type item struct {
	// Standard fields
	Title                  string  `json:"title,omitempty"`
	Subtitle               string  `json:"subtitle,omitempty"`
	URL                    string  `json:"url,omitempty"`
	Path                   string  `json:"path,omitempty"`
	Icon                   string  `json:"icon,omitempty"`
	QuickLookURL           string  `json:"quickLookURL,omitempty"`
	Action                 string  `json:"action,omitempty"`
	ActionArgument         string  `json:"actionArgument,omitempty"`
	ActionReturnsItems     bool    `json:"actionReturnsItems,omitempty"`
	ActionRunsInBackground bool    `json:"actionRunsInBackground,omitempty"`
	ActionBundleIdentifier string  `json:"actionBundleIdentifier,omitempty"`
	Children               []*item `json:"children,omitempty"`

	// Custom fields
	ID       int                    `json:"x-id,omitempty"`
	Order    int                    `json:"x-order,omitempty"`
	FuncName string                 `json:"x-func,omitempty"`
	FuncArg  string                 `json:"x-funcarg,omitempty"`
	Arg      string                 `json:"x-arg,omitempty"`
	Data     map[string]interface{} `json:"x-data,omitempty"`
}

// SetTitle sets the Item's title.
func (i *Item) SetTitle(title string) *Item { i.item.Title = title; return i }

// SetSubtitle sets the Item's subtitle that appears below or next to the title.
func (i *Item) SetSubtitle(subtitle string) *Item { i.item.Subtitle = subtitle; return i }

// SetURL sets the Item's URL. When the user selects the item and hits Enter, this URL is opened.
func (i *Item) SetURL(url string) *Item { i.item.URL = url; return i }

// SetPath sets the absolute path of a file or folder the item represents. If
// icon is not set, LaunchBar automatically uses an item that represents the
// path.
func (i *Item) SetPath(path string) *Item { i.item.Path = path; return i }

// SetIcon sets the  icon for the item. This is a string that is interpreted
// the same way as CFBundleIconFile in the action’s Info.plist.
//  http://www.obdev.at/resources/launchbar/developer-documentation/action-info-plist.html#info-plist-CFBundleIconFile
func (i *Item) SetIcon(icon string) *Item { i.item.Icon = icon; return i }

// SetQuickLookURL sets the URL to be shown by the QuickLook panel when the
// user hits ⌘Y on the item. This can by any URL supported by QuickLook,
// including http of file URLs. Items that have a path property automatically
// support QuickLook and do not need to set this property too.
func (i *Item) SetQuickLookURL(qlurl string) *Item { i.item.QuickLookURL = qlurl; return i }

// SetAction sets the name of an action that should be run when the user
// selects this item and hits Enter.  This is the name of a script file
// inside the action bundle’s Scripts folder, including the file name
// extension.
//
// The argument for the action depends on the value of actionArgument.
func (i *Item) SetAction(action string) *Item { i.item.Action = action; return i }

// SetActionArgument sets the argument to pass to the action.
//
// When the user selects this item and hits Enter and the item has an action
// set, this is the argument that gets passed to that action as a string. If
// this key is not present, the whole item is passed as an argument as a JSON
// string
func (i *Item) SetActionArgument(arg string) *Item { i.item.ActionArgument = arg; return i }

// SetActionBundleIdentifier sets the identifier of an action that should be
// run when the user selects this item and hits enter
func (i *Item) SetActionBundleIdentifier(s string) *Item { i.item.ActionBundleIdentifier = s; return i }

// SetActionRunsInBackground sets the action to be run in background
//
// See
// http://www.obdev.at/resources/launchbar/developer-documentation/action-info-plist.html#info-plist-LBRunInBackground
// for more detail.
func (i *Item) SetActionRunsInBackground(b bool) *Item { i.item.ActionRunsInBackground = b; return i }

// SetActionReturnsItems specifies that selecting and executing the item (as
// specified by the action key) will return a new list of items. If this is set
// to true, the item will have a chevron on the right side indicating that the
// user can navigate into it and doing so causes action to be executed.
func (i *Item) SetActionReturnsItems(b bool) *Item { i.item.ActionReturnsItems = b; return i }

// SetChildren sets an array of items.
func (i *Item) SetChildren(items *Items) *Item { i.item.Children = items.getItems(); return i }

// SetMatch sets the matcher func of this item. This func determines that if
// the item should be visible or not.
//
// Example:
//   func(c *Context) bool { return c.Action.IsControlKey() }
func (i *Item) SetMatch(fn Func) *Item { i.match = fn; return i }

// SetRun sets the runner func of this item. This func is optional and will be
// run when the user selects the item and hit Enter. Runner func output is
// optional, if it returns Items those items will be displayed to the user.
//
// Example:
//   func(c *Context) { c.Action.ShowView("main") }
func (i *Item) SetRun(fn Func) *Item { i.run = fn; return i }

// SetRender sets the renderer func of this item. This func will be executed
// each time the user enters a key and can be used to update the values based
// on the user input.
// Example:
//  func(c *Context) { c.Self.SetSubtitle("") }
func (i *Item) SetRender(fn Func) *Item { i.render = fn; return i }

// SetOrder sets the order of the item. The Items are ordered by their creation time.
func (i *Item) SetOrder(n int) *Item { i.item.Order = n; return i }

// Item returns an underlying LaunchBar item that can be passed around in json format.
func (i *Item) Item() *item { return i.item }

// Done returns the pointer to the Item's view, used for chaining the item creation.
func (i *Item) Done() *View { return i.View }

// Run sets the predefined func (see FuncMap) to be run with the optional
// arguments when the user selects this item and hit Enter
func (i *Item) Run(f string, args ...interface{}) *Item {
	i.item.FuncName = f
	var ok bool
	var s string
	if len(args) == 1 {
		if s, ok = args[0].(string); ok {
			i.item.FuncArg = s
		}
	}
	if len(args) > 1 || !ok {
		b, err := json.Marshal(args)
		if err == nil {
			i.item.FuncArg = string(b)
		}
	}
	return i
}
