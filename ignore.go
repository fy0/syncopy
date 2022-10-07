package main

import (
	"github.com/gobwas/glob"
	"os"
	"strings"
)

type SyncIgnore struct {
	patterns []glob.Glob
}

func (si *SyncIgnore) ReadFile(fn string) error {
	data, err := os.ReadFile(fn)
	if err != nil {
		return err
	}

	si.patterns = []glob.Glob{}
	s := strings.ReplaceAll(string(data), "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	patterns := strings.SplitN(s, "\n", -1)

	for _, p := range patterns {
		g, err := glob.Compile(strings.TrimSpace(p))
		if err != nil {
			continue
		}

		si.patterns = append(si.patterns, g)
	}
	return nil
}

func (si *SyncIgnore) AddPattern(fn string) {

}

func (si *SyncIgnore) Match(fn string) bool {
	for _, g := range si.patterns {
		if g.Match(fn) {
			return true
		}
	}
	return false
}
