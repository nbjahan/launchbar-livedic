package launchbar

import (
	"encoding/json"
	"fmt"
	"sort"
)

// View represents collection of Items in LaunchBar
type View struct {
	Action *Action
	Name   string
	Items  Items
}

// NewItem creates an always matching Item that runs in background and adds it to the view.
func (v *View) NewItem(title string) *Item {
	i := &Item{View: v, item: &item{Title: title}}
	i.SetActionRunsInBackground(true).
		SetAction(i.View.Action.Config.GetString("actionDefaultScript")).
		SetMatch(AlwasMatch)
	i.item.ID = len(v.Action.items) + 1
	i.SetOrder(len(v.Items))
	v.Items = append(v.Items, i)
	v.Action.items = append(v.Action.items, i)
	return i
}

// AddItem add an Item to the view
func (v *View) AddItem(item *Item) *View {
	item.View = v
	if item.match == nil {
		item.SetMatch(AlwasMatch)
	}
	item.item.ID = len(v.Action.items) + 1
	v.Items = append(v.Items, item)
	v.Action.items = append(v.Action.items, item)
	return v
}

// Render executes each Item Render, Match functions and returns them.
func (v *View) Render() Items {
	if len(v.Items) == 0 {
		return Items(nil)
	}

	items := &Items{}
	for _, item := range v.Items {
		v.Action.context.Self = item
		if item.match != nil {
			vals, err := v.Action.Invoke(item.match)
			if err != nil {
				v.Action.Logger.Fatalln(err)
				panic(err)
			}
			if len(vals) > 0 {
				if !vals[0].Bool() {
					continue
				}
			}
		}
		if item.render != nil {
			_, err := v.Action.Invoke(item.render)
			if err != nil {
				v.Action.Logger.Fatalln(err)
				panic(err)
			}
		}
		item.item.Arg = v.Action.Input.String()
		items.Add(item)
	}
	sort.Sort(itemsByOrder(*items))

	return *items

}

// Compile renders and output the view.Items as a json string.
func (v *View) Compile() string {
	items := v.Render()

	if items == nil {
		return ""
	}

	b, err := json.Marshal(items.getItems())
	if err != nil {
		return fmt.Sprintf(`[{"title": "%v","subtitle":"error"}]`, err)
	}
	return string(b)
}

// Join returns a new view with the items of v, w
func (v *View) Join(w *View) *View {
	if w == nil {
		return v
	}
	return &View{v.Action, v.Name, append(v.Items, w.Items...)}
}
