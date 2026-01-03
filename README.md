# Tweakio

Prowlarr indexer for Torrentio

> [!NOTE]
> Torrentio only returns the size of a single episode, so file size estimates for full seasons will be inaccurate by default. Providing a TMDB API key allows Tweakio to fetch the actual episode count, improving accuracy. If left empty, Tweakio will assume 10 episodes per season.

> [!TIP]
> If Prowlarr and Tweakio are **NOT** in the same Docker Compose file, create a new network and connect it to the two containers.

### Docker Compose

<details>
<summary>Advanced Configuration</summary>
These environment variables add optional overrides:

- **`TMDB_API_KEY`**  
  Used to fetch accurate episode counts from TMDB.  
  If unset, Tweakio assumes 10 episodes per season for size estimates.  
  Can be found at https://www.themoviedb.org/settings/api  
  You can use either API Read Access Token (V4) or API Key (V3).  
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

- **`PROXY_URL`**  
  Proxies all requests through the specified URL (gletun, warp etc).  
  Default: _(empty)_

- **`DEBUG`**  
  Enable detailed debug logging when set to `true`.  
  Default: `false`
  <br>
  </details>

```yaml
services:
  tweakio:
    image: varthe/tweakio:latest
    container_name: tweakio
    restart: unless-stopped
    environment:
      - TMDB_API_KEY= # Optional but recommended for best results. See https://www.themoviedb.org/settings/api
      - PROXY_URL= # Set this if Torrentio requests return 403 Forbidden
    ports: 3185:3185
```

### Prowlarr Integration

In Prowlarr:

1. Click on **Add Indexer**
2. Search for **Generic Torznab** and click it
3. Change **Name** to `Tweakio`
4. Set **Url** to `http://tweakio:3185`
5. Click **Test** and **Save**
