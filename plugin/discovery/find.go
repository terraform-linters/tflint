package discovery

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/mitchellh/go-homedir"
)

var pluginLocation = ".tflint.d"

type PluginSearch struct {
	Plugins []PluginMeta
}

type PluginMeta struct {
	Name string
	Path string
}

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
