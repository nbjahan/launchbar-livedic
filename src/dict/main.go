package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"reflect"
	"strings"

	"time"

	"github.com/DHowett/go-plist"
	sjson "github.com/bitly/go-simplejson"
	. "github.com/nbjahan/go-launchbar"
)

var InDev string

var pb *Action
var start = time.Now()

var funcs = map[string]Func{
	"openDictionary": func(c *Context) {
		word := c.Input.FuncArg()
		if pb.IsControlKey() {
			word = c.Input.String()
		}
		if pb.IsShiftKey() {
			exec.Command("osascript", "-e", fmt.Sprintf(`tell application "LaunchBar"
       perform action "Paste in Frontmost Application" with string "%s"
       end tell`, word)).Start()
		} else {
			exec.Command("open", "dict://"+url.QueryEscape(word)).Start()
		}
	},
}

func init() {
	pb = NewAction("Live Dictionary", ConfigValues{
		"actionDefaultScript": "dict",
		"debug":               false,
		"limit":               10,
		"autoupdate":          true,
	})
	pb.Config.Set("indev", InDev != "")
}

func main() {
	pb.Init(funcs)

	width := float64(300)
	LBPlistPath := os.ExpandEnv("$HOME/Library/Preferences/at.obdev.LaunchBar.plist")
	if fd, err := os.Open(LBPlistPath); err == nil {
		defer fd.Close()
		var pl map[string]interface{}
		decoder := plist.NewDecoder(fd)
		if err := decoder.Decode(&pl); err == nil && pl["LaunchBarWindowWidth"] != nil {
			width = float64(reflect.ValueOf(pl["LaunchBarWindowWidth"]).Float())
		} else if err != nil {
			pb.Logger.Println(err)
		}
	} else {
		pb.Logger.Println(err)
	}

	if InDev != "" {
		pb.Logger.Printf("in:\n%s\n", pb.Input.Raw())
	}

	in := pb.Input

	var i *Item
	v := pb.NewView("main")
	q := strings.TrimSpace(in.String())
	definitions := lookup(q, int(pb.Config.GetInt("limit")))

	if q != "" && len(definitions) == 0 {
		i = v.NewItem(in.String())
		i.SetIcon("com.apple.Dictionary")
		i.Run("openDictionary", q)
	}
	for _, row := range definitions {
		word := row[0]
		def := row[1]
		pos := strings.Index(def, "▶")
		if pos != -1 {
			def = def[pos+len("▶"):]
		}

		fields := strings.Fields(def)
		maxChars := int(width / 7)
		totalLen := 0
		parts := make([]string, 0, len(fields))
		for _, field := range fields {
			l := len([]rune(field))
			if totalLen+l+len("…") < maxChars {
				parts = append(parts, field)
				totalLen += l + 1
			} else {
				parts = append(parts, "…")
				break
			}
		}
		def = strings.Join(parts, " ")

		i = v.NewItem(word)
		i.SetSubtitle(def)
		i.SetIcon("DictionaryOn")
		i.Run("openDictionary", word)
	}
	if len(definitions) > 0 {
		if definitions[0][0] != q {
			i = v.NewItem(q)
			i.SetIcon("DictionaryOff")
			i.Run("openDictionary", q)
			i.SetSubtitle(q)
		} else {
			v.Items[0].SetIcon("com.apple.Dictionary")
		}

		// i = v.NewItem("⌃+Enter to open Dictionary.app")
		// i.SetIcon("LinkArrowTemplate")
		// i.SetAction("")

		// i = v.NewItem("⇧+Enter to paste in frontmost application")
		// i.SetIcon("at.obdev.LaunchBar:CopyActionTemplate")
		// i.SetAction("")
	}

	out := pb.Run()

	if false && InDev != "" {
		nice := out
		js, err := sjson.NewJson([]byte(out))
		if err == nil {
			b, err := js.EncodePretty()
			if err == nil {
				nice = string(b)
			}
		}
		pb.Logger.Println("out:", string(nice))
	}

	fmt.Println(out)
}
