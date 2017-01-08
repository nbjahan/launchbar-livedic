// Package launchbar is a package to quickly write LaunchBar v6 actions like a pro
//
// For example check :
//   https://github.com/nbjahan/launchbar-pinboard
//   https://github.com/nbjahan/launchbar-spotlight
package launchbar

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/DHowett/go-plist"
	"github.com/bitly/go-simplejson"
	"github.com/codegangsta/inject"
)

type infoPlist map[string]interface{}

// Action represents a LaunchBar action
type Action struct {
	inject.Injector // Used for dependency injection
	Config          *Config
	Cache           *Cache
	Input           *Input
	Logger          *log.Logger
	name            string
	views           map[string]*View
	items           []*Item
	context         *Context
	funcs           *FuncMap
	info            infoPlist
}

// NewAction creates an empty action, ready to populate with views
func NewAction(name string, config ConfigValues) *Action {
	a := &Action{
		Injector: inject.New(),
		name:     name,
		views:    make(map[string]*View),
		items:    make([]*Item, 0),
	}

	// config
	if _, found := config["actionDefaultScript"]; !found {
		panic("you should specify 'actionDefaultScript' in the config")
	}
	defaultConfig := ConfigValues{
		"debug":      false,
		"autoUpdate": true,
	}
	for k, v := range config {
		defaultConfig[k] = v
	}
	a.Config = NewConfigDefaults(a.SupportPath(), defaultConfig)

	a.Cache = NewCache(a.CachePath())
	fd, err := os.OpenFile(path.Join(a.SupportPath(), "error.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0644)
	if err != nil {
		fd = os.Stderr
	}
	a.Logger = log.New(fd, "", 0)
	c := &Context{
		Action: a,
		Config: a.Config,
		Cache:  a.Cache,
		Logger: a.Logger,
	}
	a.context = c
	a.Map(c)

	data, err := ioutil.ReadFile(path.Join(a.ActionPath(), "Contents", "Info.plist"))
	if err != nil {
		a.Logger.Println(err)
		panic(err)
	}
	_, err = plist.Unmarshal(data, &a.info)
	if err != nil {
		a.Logger.Println(err)
		panic(err)
	}
	return a
}

// Init parses the input
func (a *Action) Init(m ...FuncMap) *Action {
	a.funcs = &FuncMap{}
	if m != nil {
		*a.funcs = m[0]
	}

	in := NewInput(a, os.Args[1:])
	a.Input = in
	a.context.Input = in

	// TODO: needs good documentation
	if in.hasFunc && in.Item.Item().FuncName == "update" {
		updateFn := Func(update)
		if fn, ok := (*a.funcs)[in.Item.item.FuncName]; ok {
			updateFn = fn
		}

		vals, err := a.Invoke(updateFn)

		if err != nil {
			a.Logger.Fatalln(err)
		}
		if len(vals) == 0 {
			a.Logger.Fatalln("update function should return a value")
		}
		out, ok := vals[0].Interface().(string)
		if !ok {
			a.Logger.Fatalf("update function: expected string got: %#v", vals[0].Interface())
		}
		if a.InDev() {
			a.Logger.Println(out)
		}
		json, err := simplejson.NewJson([]byte(out))
		if err != nil {
			a.Logger.Fatalf("update function should return a valid json string: %q (%v)", out, err)
		}
		if _, ok := json.CheckGet("error"); !ok {
			a.Logger.Fatalf("update function bad output: %q (%s)", out, "missing 'error'")
		}
		e, err := json.Get("error").String()
		if err != nil {
			a.Logger.Fatalf("update function bad output: %q (%s)", out, "'error' is not string")
		}

		if e == "" {
			_, hasDownload := json.CheckGet("download")
			if !hasDownload {
				a.Logger.Fatalf("update function bad output: %q (%s)", out, "missing 'download'")
			}

			_, hasVersion := json.CheckGet("version")
			if !hasVersion {
				a.Logger.Fatalf("update function bad output: %q (%s)", out, "missing 'version'")
			}

			var changelog string
			_, hasChangelog := json.CheckGet("changelog")
			if !hasChangelog {
				changelog = ""
			}
			changelog, err = json.Get("changelog").String()
			if err != nil {
				a.Logger.Fatalf("update function bad output: %q (%s)", out, "'changelog' is not string")
			}

			download, err := json.Get("download").String()
			if err != nil {
				a.Logger.Fatalf("update function bad output: %q (%s)", out, "'download' is not string")
			}

			version, err := json.Get("version").String()
			if err != nil {
				a.Logger.Fatalf("update function bad output: %q (%s)", out, "'version' is not string")
			}

			a.Cache.Set("lastUpdate", time.Now(), 7*24*time.Hour)
			a.Cache.Set("updateInfo", map[string]string{
				"version":   version,
				"download":  download,
				"changelog": changelog,
			}, 7*24*time.Hour)

		} else {
			_, hasDesc := json.CheckGet("description")
			if !hasDesc {
				a.Logger.Fatalf("update function bad output: %q (%s)", out, "missing 'description'")
			}
			desc, err := json.Get("description").String()
			if err != nil {
				a.Logger.Fatalf("update function bad output: %q (%s)", out, "'description' is not string")
			}
			a.Logger.Println(e, ":", desc)
		}

		os.Exit(0)
	}

	return a
}

// Run returns the compiled output of views. You must call Init first
func (a *Action) Run() string {

	// Creating item to handle update
	i := a.GetView("main").NewItem("")
	i.Item().ID = -1
	i.SetOrder(9999)
	// i.SetSubtitle("Hold ⌃ to ignore")
	i.SetIcon("/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/ToolbarDownloadsFolderIcon.icns")
	i.SetActionRunsInBackground(false)
	i.SetActionReturnsItems(true)
	i.SetRender(func(c *Context) {
		oldversion := c.Action.Version()
		var updateInfo map[string]string
		if _, err := c.Cache.Get("updateInfo", &updateInfo); err != nil {
			return
		}
		newversion := Version(updateInfo["version"])
		c.Self.SetTitle(fmt.Sprintf("New Version Available: v%s (I'm v%s)", newversion, oldversion))
	})
	i.SetMatch(func(c *Context) bool {
		oldversion := c.Action.Version()
		var updateInfo map[string]string
		if _, err := c.Cache.Get("updateInfo", &updateInfo); err != nil {
			return false
		}
		newversion := Version(updateInfo["version"])
		return oldversion.Less(newversion)
	})
	i.SetRun(func(c *Context) *Items {
		oldversion := c.Action.Version()
		var updateInfo map[string]string
		if _, err := c.Cache.Get("updateInfo", &updateInfo); err != nil {
			return nil
		}
		newversion := Version(updateInfo["version"])
		if !oldversion.Less(newversion) {
			return nil
		}
		items := NewItems()
		items.Add(NewItem(fmt.Sprintf("Download %s", path.Base(updateInfo["download"]))).SetURL(updateInfo["download"]))
		for _, line := range strings.Split(updateInfo["changelog"], "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			items.Add(NewItem(line).SetIcon("at.obdev.LaunchBar:ContentsTemplate"))
		}
		homepage := ""
		if desc := a.info["LBDescription"]; desc != nil {
			if web := desc.(map[string]interface{})["LBWebsite"]; web != nil {
				if s, ok := web.(string); ok {
					homepage = s
				}
			}
		}
		if homepage != "" {
			items.Add(NewItem("Open Homepage").SetURL(homepage))
		}
		return items
	})

	in := a.Input
	if in.IsObject() {
		if in.hasFunc {
			// I'm not sure!
			a.context.Self = in.Item
			if fn, ok := (*a.funcs)[in.Item.item.FuncName]; ok {
				vals, err := a.Invoke(fn)
				if err != nil {
					a.Logger.Fatalln(err)
				}
				if len(vals) > 0 {
					if vals[0].Interface() != nil {
						s := ""
						switch res := vals[0].Interface().(type) {

						case Items:
							s = res.Compile()
						case string:
							s = res
						case *View:
							s = res.Compile()
						case *Items:
							s = res.Compile()
						}
						return s
					}
				}
				return ""
			}
		} else {
			if item := a.GetItem(in.Item.item.ID); item != nil {
				a.context.Self = item
				if item.run != nil {
					vals, err := a.Invoke(item.run)
					if err != nil {
						a.Logger.Fatalln(err)
					}
					if len(vals) > 0 {
						if vals[0].Interface() != nil {
							var s string
							if out, ok := vals[0].Interface().(Items); ok {
								s = out.Compile()
							} else {
								s = vals[0].Interface().(*Items).Compile()
							}
							return s
						}
					}
					return ""
				}
			}
		}
	}
	// TODO: if a.GetView(view) == nil inform the developer
	view := a.Config.GetString("view")

	if view == "" {
		view = "main"
	}

	if view == "main" {
		// check for updates
		checkForUpdates := false
		updateLink := ""
		if v := a.info["LBDescription"].(map[string]interface{})["LBUpdate"]; v != nil {
			updateLink = v.(string)
		}
		// lastUpdate := a.Config.GetInt("lastUpdate")
		var lastUpdate time.Time
		if updateLink != "" {
			// TODO: Watch this, IsControlKey, IsOptionKey does not work in LB6102
			if a.IsShiftKey() && a.IsOptionKey() {
				// TODO: notify the user
				a.Logger.Println("Force update.")
				checkForUpdates = true
			} else if a.Config.GetBool("autoUpdate") {
				if _, err := a.Cache.Get("lastUpdate", &lastUpdate); err == nil || err == ErrCacheDoesNotExists {
					if lastUpdate.Before(time.Now().AddDate(0, 0, -1)) {
						checkForUpdates = true
					}
				}
			}
		}
		if checkForUpdates {
			if a.InDev() {
				a.Logger.Println("Checking for update...")
				out, _ := exec.Command(os.Args[0], `{"x-func":"update"}`).CombinedOutput()
				if s := strings.TrimSpace(string(out)); s != "" {
					a.Logger.Println(s)
				}
			} else {
				exec.Command(os.Args[0], `{"x-func":"update"}`).Start()
			}
		}
	}

	w := a.GetView("*")
	out := a.GetView(view).Join(w).Compile()
	return out
}

// ShowView reruns the LaunchBar with the specified view.
//
// Use this when your LiveFeedback is enabled and you want to show another view
func (a *Action) ShowView(v string) {
	a.Config.Set("view", v)
	exec.Command("osascript", "-e", fmt.Sprintf(`tell application "LaunchBar"
       remain active
       perform action "%s"
       end tell`, a.name)).Start()
}

// NewView created a new view ready to populate with Items
func (a *Action) NewView(name string) *View {
	v := &View{a, name, make(Items, 0)}
	a.views[name] = v
	return v
}

// GetView returns a View if the view is not defined returns nil
func (a *Action) GetView(v string) *View {
	view, ok := a.views[v]
	if ok {
		return view
	}
	return nil
}

// GetItem return an Item with its ID. Returns nil if not found.
func (a *Action) GetItem(id int) *Item {
	for _, item := range a.items {
		if item.Item().ID == id {
			return item
		}
	}
	return nil
}

// InDev returns true if the config.indev is true
func (a *Action) InDev() bool { return a.Config.GetBool("indev") }

// Info.plist variables

// Varsion returns Action version specified by CFBundleVersion key in Info.plist
func (a *Action) Version() Version { return Version(a.info["CFBundleVersion"].(string)) }

// LaunchBar provided variabled

// ActionPath returns the absolute path to the .lbaction bundle.
func (a *Action) ActionPath() string { return os.Getenv("LB_ACTION_PATH") }

// CachePath returns the absolute path to the action’s cache directory:
//  ~/Library/Caches/at.obdev.LaunchBar/Actions/Action Bundle Identifier/
//
// The action’s cache directory can be used to store files that can be recreated
// by the action itself, e.g. by downloading a file from a server again.
//
// Currently, this directory’s contents will never be touched by LaunchBar,
// but it may be periodically cleared in a future release.
// When the action is run, this directory is guaranteed to exist.
func (a *Action) CachePath() string { return os.Getenv("LB_CACHE_PATH") }

// Supportpath returns the The absolute path to the action’s support directory:
//  ~/Library/Application Support/LaunchBar/Action Support/Action Bundle Identifier/
//
// The action support directory can be used to persist user data between runs of
// the action, like preferences. When the action is run, this directory is
// guaranteed to exist.
func (a *Action) SupportPath() string { return os.Getenv("LB_SUPPORT_PATH") }

// IsDebug returns the value corresponds to LBDebugLogEnabled in the action’s Info.plist.
func (a *Action) IsDebug() bool { return os.Getenv("LB_DEBUG_LOG_ENABLED") == "true" }

// Launchbarpath returns the path to the LaunchBar.app bundle.
func (a *Action) LaunchBarPath() string { return os.Getenv("LB_LAUNCHBAR_PATH") }

// ScriptType returns the type of the script, as defined by the action’s Info.plist.
//
// This is either “default”, “suggestions” or “actionURL”.
//
// See http://www.obdev.at/resources/launchbar/developer-documentation/action-programming-guide.html#script-types for more information.
func (a *Action) ScriptType() string { return os.Getenv("LB_SCRIPT_TYPE") }

// IsCommandKey returns true if the Command key was down while running the action.
func (a *Action) IsCommandKey() bool { return os.Getenv("LB_OPTION_COMMAND_KEY") == "1" }

// IsOptionKey returns true if the Alternate (Option) key was down while running the action.
func (a *Action) IsOptionKey() bool { return os.Getenv("LB_OPTION_ALTERNATE_KEY") == "1" }

// IsShiftKey returns true if the Shift key was down while running the action.
func (a *Action) IsShiftKey() bool { return os.Getenv("LB_OPTION_SHIFT_KEY") == "1" }

// IsControlKey returns true if the Control key was down while running the action.
func (a *Action) IsControlKey() bool { return os.Getenv("LB_OPTION_CONTROL_KEY") == "1" }

// IsBackground returns true if the action is running in background.
func (a *Action) IsBackground() bool { return os.Getenv("LB_OPTION_RUN_IN_BACKGROUND") == "1" }
