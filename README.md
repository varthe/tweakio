# Tweakio

Tweakio makes Torrentio usable as an indexer in Prowlarr, allowing it to integrate seamlessly with Radarr and Sonarr.

> [!NOTE]
> Torrentio only returns the size of a single episode, so file size estimates for full seasons will be inaccurate by default. Providing a TMDB API key allows Tweakio to fetch the actual episode count, improving accuracy. If left empty, Tweakio will assume 10 episodes per season.

> [!TIP]
> If Prowlarr and Tweakio are **NOT** in the same Docker Compose file, create a new network and connect it to the Prowlarr container.
>
> ```bash
> docker network create tweakio_network
> docker network connect tweakio_network prowlarr_container
> ```

### Docker Compose

<details>
<summary>Advanced Configuration</summary>
These environment variables add optional overrides:

- **`TMDB_API_KEY`**  
  Used to fetch accurate episode counts from TMDB.  
  If unset, Tweakio assumes 10 episodes per season for size estimates.  
  Default: _(empty)_

- **`TMDB_CACHE_SIZE`**  
  Max number of episode count results to cache from TMDB.  
  Default: `1000`

- **`TORRENTIO_BASE_URL`**  
  Overrides the base URL used for Torrentio requests.  
  Default: `https://torrentio.strem.fun/`

- **`TORRENTIO_OPTIONS`**  
   Overrides providers and filtering options used by Torrentio.
  Default:
  ```
  providers=yts,eztv,rarbg,1337x,thepiratebay,kickasstorrents,
  torrentgalaxy,magnetdl,horriblesubs,nyaasi,tokyotosho,anidex
  |sort=qualitysize|qualityfilter=scr,cam
  ```
  <br>
  </details>

```yaml
services:
  tweakio:
    image: varthe/tweakio:latest
    container_name: tweakio
    environment:
      - TMDB_API_KEY="" # Optional but recommended for best results
    ports:
      - "3185:3185"
    networks:
      - tweakio_network

networks:
  tweakio_network:
    external: true
```

### Prowlarr Integration

In Prowlarr:

1. Click on **Add Indexer**
2. Search for **Generic Torznab** and click it
3. Change **Name** to `Tweakio`
4. Set **Url** to `http://tweakio:3185`
5. Click **Test** and **Save**
