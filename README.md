# Mooch

Mooch downloads and organizes torrents from RSS feeds.

Usage:

    mooch [file]

Mooch matches torrent titles with regular expressions and adds them to a
[Rain](https://github.com/cenkalti/rain) client session. Torrents are
organized upon completion if destination directories are specified.

The configuration file should either be passed as an argument or exist
at $XDG_CONFIG_HOME/mooch/config.json. The configuration should look
similar to:

```json
{
  "data_dir": "~/Downloads",
  "feeds": [
    {
      "url": "https://example.com/rss?user=bob",
      "pattern": "Popular Series - (\\d+) \\[1080p\\]\\[HEVC\\]",
      "dst_dir": "~/Media/Popular Series/Season 01"
    },
    {
      "url": "https://another.org/feed?category=fantasy",
      "pattern": "Ongoing Show - S03E(\\d+) \\[720p\\]",
      "dst_dir": "~/Media/Ongoing Show/Season 03"
    }
  ]
}
```
