package launchbar

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"time"

	"github.com/DHowett/go-plist"
)

func update(c *Context) string {
	updateLink := c.Action.info["LBDescription"].(map[string]interface{})["LBUpdate"].(string)
	var updateStartTime time.Time
	if _, err := c.Cache.Get("updateStartTime", &updateStartTime); err == nil {
		return die("update in progress", fmt.Sprintf("update check in progress (started %v ago)", time.Now().Sub(updateStartTime)))
	}
	c.Cache.Set("updateStartTime", time.Now(), 3*time.Minute)
	defer func() {
		c.Cache.Delete("updateStartTime")
	}()

	var data, updatePlist []byte
	var etag, updateETag string
	c.Cache.Get("updateETag", &etag)
	if etag != "" {
		resp, err := http.Head(updateLink)
		if err != nil {
			return die("cannot get updateLink", fmt.Sprintf("%v", err))
		}
		updateETag = resp.Header.Get("etag")
		if etag == updateETag {
			c.Cache.Get("updatePlist", &updatePlist)
			if len(updatePlist) != 0 {
				data = updatePlist
			}
		}
	}

	if len(data) == 0 {
		resp, err := http.Get(updateLink)
		if err != nil {
			return die("cannot get updateLink", fmt.Sprintf("%v", err))
		}
		if resp.StatusCode >= 400 {
			return die("cannot get updateLink", fmt.Sprintf("%v", resp.Status))
		}
		defer resp.Body.Close()
		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return die("cannot get updateLink", fmt.Sprintf("%v", err))
		}
		c.Cache.Set("updateETag", resp.Header.Get("etag"), 0)
		c.Cache.Set("updatePlist", data, 0)
	}

	var v map[string]interface{}
	_, err := plist.Unmarshal(data, &v)
	if err != nil {
		return die("cannot parse updateLink", fmt.Sprintf("Error: %v\nData: %s", err, string(data)))
	}

	var updateVersion, updateDownload, updateChangelog string

	if v["CFBundleVersion"] != nil {
		if s, ok := v["CFBundleVersion"].(string); ok {
			updateVersion = s
		}
	}
	if updateVersion == "" {
		return die("no remote version", "cannot get the remote version!")
	}

	if v["LBDescription"] != nil && v["LBDescription"].(map[string]interface{}) != nil {
		if v["LBDescription"].(map[string]interface{})["LBDownload"] != nil {
			if s, ok := v["LBDescription"].(map[string]interface{})["LBDownload"].(string); ok {
				updateDownload = s
			}
		}
		if updateDownload == "" {
			return die("no remote download", "cannot get the remote download link!")
		}

		if v["LBDescription"].(map[string]interface{})["LBChangelog"] != nil {
			if s, ok := v["LBDescription"].(map[string]interface{})["LBChangelog"].(string); ok {
				updateChangelog = s
			}
		}
	}

	return write(map[string]interface{}{
		"error":     "",
		"version":   updateVersion,
		"download":  updateDownload,
		"changelog": updateChangelog,
	})
}

func write(m map[string]interface{}) string {
	data, err := json.Marshal(m)

	if err != nil {
		return "\"\""
	}
	return string(data)
}

func die(err, desc string) string {
	return write(map[string]interface{}{
		"error":       err,
		"description": desc,
	})
}
