package main

// idea: instead of having python/ruby/curl/wget/fetch/git installed
// for a vim-plugin-manager to fetch the plugins i just want one binary
// which does it all.
//
// plugins: http://vimawesome.com/

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var allowed_actions = []string{
	"u",
	"up",
	"update",
	"c",
	"clean",
	"sample",
}

func usage() {

	fmt.Fprintln(os.Stderr, `vopher - acquire vim-plugins the gopher-way

usage: vopher [flags] <action>

actions
  update - acquire the given plugins from the -f list
  clean  - remove given plugins frmo the -f list
  sample - print sample vopher.list to stdout

flags`)
	flag.PrintDefaults()
}

func sample() {
	fmt.Println(`# sample vopher.list file
# a comment starts with a '#', the whole line gets ignored.
# empty lines are ignored as well.

# fetch tpope's 'vim-fugitive' plugin, the master branch (rox, btw)
# and places the content of the zip-file into -dir <folder>/vim-fugitive.
https://github.com/tpope/vim-fugitive

# fetch tpope's 'vim-fugitive' plugin, but grab the tagged release 'v2.1'
# instead of 'master'.
https://github.com/tpope/vim-fugitive#v2.1

# fetch tpope's 'vim-fugitive' plugin and place it under -dir <folder>/foo
foo https://github.com/tpope/vim-fugitive

# fetch tpope's 'vim-fugitive' plugin, but do not strip any directories
# from the filenames in the zip. the default is to strip the first directory
# name, but sometimes you need to have more control.
vim-fugitive https://github.com/tpope/vim-fugitive strip=0`)
}

func main() {

	log.SetPrefix("vopher.")
	cli := struct {
		action string
		force  bool
		file   string
		dir    string
		ui     string
	}{action: "update", dir: ".", ui: "progressline"}

	flag.BoolVar(&cli.force, "force", cli.force, "if already existant: refetch plugins")
	flag.StringVar(&cli.file, "f", cli.file, "path to list of plugins")
	flag.StringVar(&cli.dir, "dir", cli.dir, "directory to extract the plugins to")
	flag.StringVar(&cli.ui, "ui", cli.ui, "ui mode")

	flag.Usage = usage
	flag.Parse()

	if len(flag.Args()) > 0 {
		cli.action = flag.Args()[0]
	}

	if prefix_in_stringslice(allowed_actions, cli.action) == -1 {
		log.Fatal("error: unknown action")
	}

	if cli.action == "sample" {
		sample()
		return
	}

	var ui JobUi
	switch cli.ui {
	case "progressline":
		ui = &UiOneLine{
			ProgressTicker: NewProgressTicker(0),
			prefix:         "vopher",
			duration:       25 * time.Millisecond,
		}
	case "simple":
		ui = &UiSimple{jobs: make(map[string]_ri)}
	}

	switch cli.action {
	case "update", "u", "up":
		plugins := must_read_plugins(cli.file)
		update(plugins, cli.dir, cli.force, ui)
	case "clean", "c", "cl":
		plugins := must_read_plugins(cli.file)
		clean(plugins, cli.dir, cli.force)
	}
}

func update(plugins PluginList, dir string, force bool, ui JobUi) {

	ui.Start()

	for _, plugin := range plugins {

		plugin_folder := filepath.Join(dir, plugin.name)

		_, err := os.Stat(plugin_folder)
		if err == nil { // plugin_folder exists
			if !force {
				continue
			}
		}

		if !strings.HasSuffix(plugin.url.Path, ".zip") {
			switch plugin.url.Host {
			case "github.com":
				remote_zip := first_not_empty(plugin.url.Fragment, "master") + ".zip"
				plugin.url.Path = path.Join(plugin.url.Path, "archive", remote_zip)
			default:
				ext, err := httpdetect_ftype(plugin.url.String())
				if err != nil {
					log.Printf("error: %q: %s", plugin.url, err)
					continue
				}
				if ext != ".zip" {
					log.Printf("error: %q: not a zip", plugin.url)
					continue
				}
			}
		}

		ui.AddJob(plugin_folder)
		go acquire(plugin_folder, plugin.url.String(), plugin.strip_dir, ui)
	}
	ui.Wait()
	ui.Stop()
}

func clean(plugins PluginList, dir string, force bool) {

	if !force {
		log.Println("'clean' needs -force flag")
		return
	}

	var prefix, suffix string

	for _, plugin := range plugins {
		plugin_folder := filepath.Join(dir, plugin.name)
		prefix = ""
		suffix = "ok"
		_, err := os.Stat(plugin_folder)
		if err == nil { // plugin_folder exists
			err = os.RemoveAll(plugin_folder)
			if err != nil {
				prefix = "error:"
				suffix = err.Error()
			}
		} else {
			prefix = "info:"
			suffix = "does not exist"
		}
		log.Println("'clean'", prefix, plugin_folder, suffix)
	}
}

func must_read_plugins(path string) PluginList {
	plugins, err := ScanPluginFile(path)
	if err != nil {
		log.Fatal(err)
	}

	if len(plugins) == 0 {
		log.Fatalf("empty plugin-file %q", path)
	}
	return plugins
}
