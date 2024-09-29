// Copyright 2024 Matthew P. Dargan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Mooch downloads and organizes torrents from RSS feeds.
//
// Usage:
//
//	mooch [file]
//
// Mooch matches torrent titles with regular expressions and adds them
// to a [Rain] client session. Torrents are organized upon completion if
// destination directories are specified.
//
// The configuration file should either be passed as an argument or
// exist at $XDG_CONFIG_HOME/mooch/config.json. The configuration should
// look similar to:
//
//	{
//	  "data_dir": "~/Downloads",
//	  "feeds": [
//	    {
//	      "url": "https://example.com/rss?user=bob",
//	      "pattern": "Popular Series - (\\d+) \\[1080p\\]\\[HEVC\\]",
//	      "dst_dir": "~/Media/Popular Series/Season 01"
//	    },
//	    {
//	      "url": "https://another.org/feed?category=fantasy",
//	      "pattern": "Ongoing Show - S03E(\\d+) \\[720p\\]"
//	    }
//	  ]
//	}
//
// [Rain]: https://github.com/cenkalti/rain
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"syscall"

	"github.com/cenkalti/rain/torrent"
	"github.com/matthewdargan/epify/media"
	"github.com/mmcdole/gofeed"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: mooch [file]\n")
	os.Exit(2)
}

func main() {
	log.SetPrefix("mooch: ")
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()
	var name string
	switch flag.NArg() {
	case 0:
		dir, err := os.UserConfigDir()
		if err != nil {
			log.Fatal(err)
		}
		name = filepath.Join(dir, "mooch", "config.json")
	case 1:
		name = flag.Arg(0)
	default:
		usage()
	}
	f, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	cfg, err := readConfig(f)
	if err != nil {
		log.Fatal(err)
	}
	fp := gofeed.NewParser()
	for i, fd := range cfg.Feeds {
		var f *gofeed.Feed
		f, err = fp.ParseURL(fd.URL)
		if err != nil {
			log.Fatal(err)
		}
		for _, it := range f.Items {
			if fd.regexp.MatchString(it.Title) {
				log.Printf("matched %s", it.Title)
				cfg.Feeds[i].link = it.Link
				break
			}
		}
	}
	tcfg := torrent.DefaultConfig
	tcfg.Database = filepath.Join(cfg.DataDir, "session.db")
	tcfg.DataDir = filepath.Join(cfg.DataDir, "data")
	tcfg.DataDirIncludesTorrentID = false
	sess, err := torrent.NewSession(tcfg)
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()
	for i, f := range cfg.Feeds {
		cfg.Feeds[i].torr, err = sess.AddURI(f.link, nil)
		if err != nil {
			log.Print(err)
		}
	}
	for _, t := range sess.ListTorrents() {
		<-t.NotifyComplete()
	}
	for _, f := range cfg.Feeds {
		if f.DstDir == nil || f.torr == nil {
			continue
		}
		ps, err := f.torr.Files()
		if err != nil {
			log.Print(err)
			continue
		}
		eps := make([]string, 0, len(ps))
		for _, p := range ps {
			name := filepath.Join(tcfg.DataDir, p.Path())
			log.Printf("p.Path() = %s, name = %s\n", p.Path(), name)
			fi, sErr := os.Stat(name)
			if sErr != nil {
				continue
			}
			sys := fi.Sys()
			if sys == nil {
				eps = append(eps, name)
				continue
			}
			stat, ok := sys.(*syscall.Stat_t)
			if !ok {
				eps = append(eps, name)
				continue
			}
			if stat.Nlink < 2 {
				eps = append(eps, name)
			}
		}
		a := media.Addition{SeasonDir: *f.DstDir, Episodes: eps}
		if err = media.AddEpisodes(a); err != nil {
			log.Print(err)
		}
	}
}

type config struct {
	DataDir string `json:"data_dir"`
	Feeds   []feed `json:"feeds"`
}

type feed struct {
	URL     string  `json:"url"`
	Pattern string  `json:"pattern"`
	DstDir  *string `json:"dst_dir"`
	regexp  *regexp.Regexp
	link    string
	torr    *torrent.Torrent
}

func readConfig(r io.Reader) (config, error) {
	var c config
	if err := json.NewDecoder(r).Decode(&c); err != nil {
		return config{}, err
	}
	if err := c.validate(); err != nil {
		return config{}, err
	}
	return c, nil
}

var (
	errDataDir = errors.New("data directory not found")
	errFeeds   = errors.New("feeds not found")
	errURL     = errors.New("URL cannot be empty")
	errPattern = errors.New("pattern cannot be empty")
)

func (c *config) validate() error {
	if c.DataDir == "" {
		return errDataDir
	}
	if len(c.Feeds) == 0 {
		return errFeeds
	}
	for i, f := range c.Feeds {
		if f.URL == "" {
			return errURL
		}
		if f.Pattern == "" {
			return errPattern
		}
		var err error
		c.Feeds[i].regexp, err = regexp.Compile(f.Pattern)
		if err != nil {
			return err
		}
	}
	return nil
}
