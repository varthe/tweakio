# tweakio

Prowlarr indexer for Torrentio

### Docker Compose

```yaml
tweakio:
  image: varthe/tweakio:latest
  container_name: tweakio
  hostname: tweakio
  ports:
    - "3185:3185"
  volumes:
    - /opt/tweakio/config.yaml:/app/config.yaml
```

### Config.yaml

```yaml
torrentio:
  base_url: https://torrentio.strem.fun/
  options: "providers=yts,eztv,rarbg,1337x,thepiratebay,kickasstorrents,torrentgalaxy,magnetdl,horriblesubs,nyaasi,tokyotosho,anidex|sort=qualitysize|qualityfilter=scr,cam"

tmdb:
  api_key: "" # If empty, defaults to 10 episodes for everything
  cache_size: 1000

regex:
  # Matches season numbers and season packs: "Season 3", "S02 COMPLETE", "Season 2 Pack", "S02.MULTi"
  season: "(?i)\\b(?:season\\s*|s)(\\d{1,2})(?:\\s+(?:complete|full|pack|multi)|\\.[a-z]+)?\\b"

  # Matches season ranges: "S01-S03", "Season 1-3", "S03-05", "S1-3"
  season_range: "(?i)\\b(?:s(?:eason)?\\s*\\d{1,2}(?:-s?(?:eason)?\\s*\\d{1,2})|s\\d{1,2}e\\d{1,2}-\\d{1,2})\\b"

  # Matches single episodes: "S02E05", "s01e01", "S1E1"
  single_episode: "(?i)\\bs\\d{1,2}e\\d{1,2}\\b"

  # Matches generic episode numbers: "E05", "e10"
  episode: "\\b[eE]\\d{2,3}\\b"

  # Matches episode ranges:"S02E01-09", "S1E01-10", "E01-09" (without season)
  episode_range: "\\bS?\\d{1,2}E(\\d{1,3})-(\\d{1,3})\\b"

  # Extracts peers, size, and source from text: 👤 15 💾 28.68 GB ⚙️ TorrentGalaxy.
  info: "👤\\s*(\\d+)\\s*💾\\s*([\\d.]+)\\s*(GB|MB)\\s*⚙️\\s*(.+)"
```

#### Prowlarr Integration

1. Click on **Add Indexer**
2. Search for **Generic Torznab** and click it
3. Change **Name** to `Tweakio`
4. Set **Url** to `http://tweakio:3185`
5. Click **Test** and **Save**
