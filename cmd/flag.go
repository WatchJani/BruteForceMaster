package cmd

import "flag"

func Load() string {
	path := flag.String("path", "", "path to the config file")

	flag.Parse()

	return *path
}
