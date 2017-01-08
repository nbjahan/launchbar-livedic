package launchbar

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Input represents the object that LaunchBar passes to scripts
type Input struct {
	Item *Item

	args           []string
	isObject       bool
	isString       bool
	isPaths        bool
	isNumber       bool
	isFloat        bool
	isInt          bool
	isLiveFeedback bool
	hasFunc        bool
	hasData        bool
	paths          []string
	number         float64
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func NewInput(a *Action, args []string) *Input {
	item := item{}
	var in = &Input{
		args: args,
	}

	if len(args) == 0 {
		return in
	}

	in.isLiveFeedback = a.IsBackground()
	if len(args) > 1 {
		in.isPaths = true
		in.paths = args
		return in
	}

	if len(args) == 1 && exists(args[0]) {
		in.isPaths = true
		in.paths = args
		return in
	}

	if err := json.Unmarshal([]byte(args[0]), &item); err == nil {
		in.isObject = true
		if item.Data != nil && len(item.Data) > 0 {
			in.hasData = true
		}
		in.Item = a.GetItem(item.ID)
		if in.Item == nil {
			in.Item = newItem(&item)
		}
		in.Item.item.Arg = item.Arg

		in.Item.item.Order = item.Order
		in.Item.item.FuncName = item.FuncName
		in.Item.item.Data = item.Data
		if item.FuncName != "" {
			in.hasFunc = true
		}
		if in.Item.item.Path != "" && exists(in.Item.item.Path) {
			in.isPaths = true
			in.paths = []string{in.Item.item.Path}
		}
	} else {
		in.isString = true
		if f64, err := strconv.ParseFloat(in.String(), 64); err == nil {
			in.isNumber = true
			in.number = f64
			if fmt.Sprintf("%f", f64) == fmt.Sprintf("%f", float64(int64(f64))) {
				in.isInt = true
			} else {
				in.isFloat = true
			}
		}
	}
	return in

}

func (in *Input) Int() int         { return int(in.number) }
func (in *Input) Float64() float64 { return in.number }
func (in *Input) Int64() int64     { return int64(in.number) }
func (in *Input) Raw() string      { return strings.Join(in.args, "\n") }

func (in *Input) String() string {
	if in.IsObject() {
		return in.Item.item.Arg
	}
	return in.Raw()
}

func (in *Input) FuncArg() string {
	if out := in.FuncArgs(); len(out) > 0 {
		return out[0]
	}
	return ""
}

// TODO:
// Deprecated use FuncArgs
func (in *Input) FuncArgsString() []string {
	out := in.FuncArgs()
	l := len(out)
	if l == 0 {
		return nil
	}
	args := make([]string, l)
	for i := 0; i < l; i++ {
		args[i] = out[i]
	}
	return args
}

// TODO:
// Deprecated use FuncArgs
func (in *Input) FuncArgsMapString() map[int]string { return in.FuncArgs() }

func (in *Input) FuncArgs() map[int]string {
	out := make(map[int]string)
	if !in.isObject {
		return nil
	}
	var args []interface{}
	err := json.Unmarshal([]byte(in.Item.item.FuncArg), &args)
	if err != nil {
		out[0] = in.Item.item.FuncArg
		return out
	}
	for i, arg := range args {
		out[i] = fmt.Sprintf("%v", arg)
	}
	return out
}

func (in *Input) Title() string {
	if in.IsObject() {
		return in.Item.item.Title
	}
	return ""
}

func (in *Input) Data(key string) interface{} {
	if in.hasData {
		if i, ok := in.Item.item.Data[key]; ok {
			return i
		}
	}
	return nil
}

// DataString returns a customdata[key] as string
func (in *Input) DataString(key string) string {
	if in.Item == nil {
		return ""
	}
	if s, ok := in.Item.item.Data[key].(string); ok {
		return s
	}
	return ""
}

// DataInt returns a customdata[key] as string
func (in *Input) DataInt(key string) int {
	if s, ok := in.Item.item.Data[key].(int); ok {
		return s
	}
	return 0
}

func (in *Input) IsString() bool { return in.isString }
func (in *Input) IsObject() bool { return in.isObject }
func (in *Input) IsPaths() bool  { return in.isPaths }
func (in *Input) IsNumber() bool { return in.isNumber }
func (in *Input) IsInt() bool    { return in.isInt }
func (in *Input) IsFloat() bool  { return in.isFloat }
func (in *Input) IsEmpty() bool  { return in.String() == "" }

// FIXME: experimental
func (in *Input) IsLiveFeedback() bool { return in.isLiveFeedback }

func (in *Input) Paths() []string { return in.paths }
