package main

import (
	"fmt"

	"github.com/Reidond/garrison/internal/server"
)

type ListCmd struct{}

func (c *ListCmd) Run(root *Root) error {
	items, err := server.ListMetadata()
	if err != nil {
		return err
	}
	for _, md := range items {
		fmt.Printf("%s\t%d\t%s\n", md.Name, md.AppID, md.InstallDir)
	}
	return nil
}
