package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type brewStep struct{ Package MacPackage }

func (s brewStep) ID() string       { return "package:" + s.Package.ID }
func (s brewStep) Describe() string { return "Install " + s.Package.Name }
func (s brewStep) Why() string {
	return "Application exists outside Homebrew; leaving it in place. Manage or remove it yourself before installing this cask."
}
func (s brewStep) Check(c *Context) (State, error) {
	brew := c.brewPath()
	if brew == "" {
		return StateUnknown, fmt.Errorf("Homebrew is missing; open magus run for guided setup")
	}
	out, err := c.macCommand(brew, "list", "--"+s.Package.Kind, "-1")
	if err != nil {
		return StateUnknown, err
	}
	for _, id := range strings.Fields(out) {
		if id == s.Package.ID {
			return StateOK, nil
		}
	}
	if s.Package.AppBundle != "" {
		roots := c.AppRoots
		if roots == nil {
			roots = []string{"/Applications", filepath.Join(c.Paths.Home, "Applications")}
		}
		for _, root := range roots {
			_, err := os.Stat(filepath.Join(root, s.Package.AppBundle))
			if err == nil {
				return StateNotApplicable, nil
			}
			if !os.IsNotExist(err) {
				return StateUnknown, err
			}
		}
	}
	return StateMissing, nil
}
func (s brewStep) Apply(c *Context) error {
	if c.DryRun {
		return nil
	}
	brew := c.brewPath()
	if brew == "" {
		return fmt.Errorf("Homebrew is missing")
	}
	_, err := c.macCommand(brew, "install", "--"+s.Package.Kind, s.Package.ID)
	return err
}

// Magus never assumes ownership of packages or dependencies on this shared Mac.
func (s brewStep) Remove(c *Context) error { return errNotReversible }
