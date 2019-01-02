package discovery

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/mitchellh/go-homedir"
)

// The default location where plugins are stored.
var pluginLocation = ".tflint.d"

// A collection of plugins returned by a search.
type PluginSearch struct {
	Plugins []PluginMeta
}

// Metadata for each plugin.
type PluginMeta struct {
	Name string
	Path string
}

/*
Searches through the .tflint.d directory for plugins and returns metadata for those plugins.
*/
func (p *PluginSearch) Find() PluginSearch {
	homeDir, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}

	homeDirExpand, err := homedir.Expand(homeDir)
	if err != nil {
		log.Fatal(err)
	}

	pluginDirectory := fmt.Sprintf("%s/%s", homeDirExpand, pluginLocation)

	if _, err := os.Stat(pluginDirectory); os.IsNotExist(err) {
		return PluginSearch{}
	}

	pluginDirectoryContents, err := ioutil.ReadDir(pluginDirectory)
	if err != nil {
		log.Fatal(err)
	}

	for _, pluginFile := range pluginDirectoryContents {
		pluginPath := fmt.Sprintf("%s/%s", pluginDirectory, pluginFile.Name())

		pluginMeta := PluginMeta{
			Name: pluginFile.Name(),
			Path: pluginPath,
		}

		p.Plugins = append(p.Plugins, pluginMeta)
	}

	return *p
}
