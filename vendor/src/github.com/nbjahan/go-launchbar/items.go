package launchbar

import (
	"encoding/json"
	"fmt"
)

// Items represents the collection of items
type Items []*Item

// NewItems creates an empty Items collection
func NewItems() *Items {
	return &Items{}
}

// Add adds passes Items to the collection and returns the collection.
func (items *Items) Add(i ...*Item) *Items {
	for _, item := range i {
		*items = append(*items, item)
	}
	return items
}

func (i *Items) setItems(items []*item) {
	for _, item := range items {
		i.Add(newItem(item))
	}
}

func (items *Items) getItems() []*item {
	a := make([]*item, len(*items))
	for i, item := range *items {
		a[i] = item.item
	}
	return a
}

type itemsByOrder Items

func (o itemsByOrder) Len() int           { return len(o) }
func (o itemsByOrder) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }
func (o itemsByOrder) Less(i, j int) bool { return o[i].item.Order < o[j].item.Order }

// Compile returns items collection as a json string.
func (items *Items) Compile() string {
	if items == nil {
		return ""
	}
	if len(*items) == 0 {
		return ""
	}

	b, err := json.Marshal(items.getItems())
	if err != nil {
		return fmt.Sprintf(`[{"title": "%v","subtitle":"error"}]`, err)
	}
	return string(b)
}
