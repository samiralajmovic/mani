package core

import (
	"fmt"
	"os"
)

type FailedToOpenFile struct {
	Name string
}

func (f *FailedToOpenFile) Error() string {
	return fmt.Sprintf("error: failed to open %q", f.Name)
}

type MissingFile struct {
	Name string
}

func (f *MissingFile) Error() string {
	return fmt.Sprintf("error: missing %q", f.Name)
}

type FailedToParseFile struct {
	Name string
	Msg  error
}

type FailedToParsePath struct {
	Name string
}

func (f *FailedToParsePath) Error() string {
	return fmt.Sprintf("error: failed to parse path %q", f.Name)
}

func (f *FailedToParseFile) Error() string {
	return fmt.Sprintf("error: failed to parse %q \n%s", f.Name, f.Msg)
}

type PathDoesNotExist struct {
	Path string
}

func (p *PathDoesNotExist) Error() string {
	return fmt.Sprintf("fatal: path %q does not exist", p.Path)
}

type CommandNotFound struct {
	Name string
}

func (c *CommandNotFound) Error() string {
	return fmt.Sprintf("fatal: could not find command %q", c.Name)
}

type ConfigNotFound struct {
	Names []string
}

func (f *ConfigNotFound) Error() string {
	return fmt.Sprintf("fatal: could not find any configuration file %v in current directory or any of the parent directories", f.Names)
}

type FileNotFound struct {
	Name string
}

func (f *FileNotFound) Error() string {
	return fmt.Sprintf("fatal: could not find %q (in current directory or any of the parent directories)", f.Name)
}

func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("%s\n", err)
	os.Exit(1)
}
