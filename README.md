# Tweakio

Tweakio makes Torrentio usable as an indexer in Prowlarr, allowing it to integrate seamlessly with Radarr and Sonarr.

#### ⚠️ Note about file sizes and TMDB

Torrentio only returns the size of a single episode, so file size estimates for full seasons will be inaccurate by default. Providing a TMDB API key allows Tweakio to fetch the actual episode count, improving accuracy. If left empty, Tweakio will assume 10 episodes per season.

#### ⚠️ Oracle VPS users will need to route Tweakio through Warp or a VPN

### Docker Compose

> [!TIP]
> If Prowlarr and Tweakio are **NOT** in the same Docker Compose file, create a new network and connect it to the Prowlarr container. Then uncomment the `networks` section of the Docker Compose.
>
> ```bash
> docker network create tweakio_network
> docker network connect tweakio_network prowlarr_container
> ```

```yaml
services:
  tweakio:
  image: varthe/tweakio:latest
  container_name: tweakio
  hostname: tweakio
  ports:
    - "3185:3185"
  volumes:
    - /opt/tweakio/config.yaml:/app/config.yaml
  # networks:
  #     - tweakio_network
# networks:
#     tweakio_network:
#         external: true
```

### Config.yaml

```yaml
torrentio:
  base_url: https://torrentio.strem.fun/
  options: "providers=yts,eztv,rarbg,1337x,thepiratebay,kickasstorrents,torrentgalaxy,magnetdl,horriblesubs,nyaasi,tokyotosho,anidex|sort=qualitysize|qualityfilter=scr,cam"

tmdb:
  api_key: "" # If empty, defaults to 10 episodes for everything
  cache_size: 1000
```

### Prowlarr Integration

Then in Prowlarr:

2. Click on **Add Indexer**
3. Search for **Generic Torznab** and click it
4. Change **Name** to `Tweakio`
5. Set **Url** to `http://tweakio:3185`
6. Click **Test** and **Save**
