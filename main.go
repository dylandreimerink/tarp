// Copyright 2021 Dylan Reimerink. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.txt file.

package main

import (
	"fmt"
	"html/template"
	"os"
	"strings"

	_ "embed"

	"github.com/dylandreimerink/gocovmerge"
	"github.com/spf13/cobra"
	"golang.org/x/tools/cover"
)

func main() {
	makeCmd().Execute()
}

var flagOutput string

func makeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "tarp {coverage file 1} [coverage file N...]",
		Long: "Tarp generates an interactive HTML coverage report for Go coverage files",
		RunE: run,
		Args: cobra.MinimumNArgs(1),
	}

	f := cmd.Flags()
	f.StringVarP(&flagOutput, "output", "o", "./coverage.html", "The generated coverage report")

	return cmd
}

//go:embed cover.tpl.html
var reportTpl string

func run(cmd *cobra.Command, args []string) error {
	profiles, err := openAndMergeReports(args)
	if err != nil {
		return fmt.Errorf("Open and merge: %w", err)
	}

	pkgs, err := findPkgs(profiles)
	if err != nil {
		return fmt.Errorf("Find packages: %w", err)
	}

	ctx := renderContext{}

	for pkgName, pkg := range pkgs {
		if pkg.Error != nil {
			return fmt.Errorf("resolving package import path: %w", err)
		}

		node := ctx.PackageRadix.Make(pkgName)
		node.Pkg = true
	}

	for _, profile := range profiles {
		fn := profile.FileName

		file, err := findFile(pkgs, fn)
		if err != nil {
			return err
		}

		funcs, err := findFuncs(file)
		if err != nil {
			return err
		}

		node := ctx.PackageRadix.Make(fn)
		node.File = true

		// Now match up functions and profile blocks.
		for _, f := range funcs {
			c, t := f.coverage(profile)

			n := node
			for n != nil {
				n.Total += t
				n.Covered += c
				n = n.Parent
			}
		}

		if profile.Mode == "set" {
			node.SetMode = true
		}
		src, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("can't read %q: %v", fn, err)
		}
		var buf strings.Builder
		err = htmlGen(&buf, src, profile.Boundaries(src))
		if err != nil {
			return err
		}
		node.Body = template.HTML(buf.String())
	}

	ctx.PackageRadix.Simplify()

	output, err := os.Create(flagOutput)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	tpl := template.New("report")
	tpl = tpl.Funcs(template.FuncMap{
		"colors": colors,
	})
	tpl, err = tpl.Parse(reportTpl)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	err = tpl.Execute(output, ctx)
	if err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	return nil
}

func openAndMergeReports(args []string) ([]*cover.Profile, error) {
	var merged []*cover.Profile
	for _, coverFilePath := range args {
		profiles, err := cover.ParseProfiles(coverFilePath)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse coverage file: %w", err)
		}
		for _, p := range profiles {
			merged = gocovmerge.AddProfile(merged, p)
		}
	}

	return merged, nil
}

type renderContext struct {
	PackageRadix radixNode
}
