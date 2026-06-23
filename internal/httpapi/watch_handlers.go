package httpapi

import (
	"fmt"
	"html"
	"net/http"
)

func (s *Server) watchLive(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing stream id")
		return
	}

	writeWatchPage(w, id, "live", "/streams/"+id+"/live/master.m3u8")
}

func (s *Server) watchVOD(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing stream id")
		return
	}

	writeWatchPage(w, id, "vod", "/streams/"+id+"/vod/master.m3u8")
}

func writeWatchPage(w http.ResponseWriter, id string, mode string, playlistURL string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	title := html.EscapeString(id + " - " + mode)
	playlist := html.EscapeString(playlistURL)
	escapedMode := html.EscapeString(mode)

	page := fmt.Sprintf(`<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>%s</title>
  <link rel="stylesheet" href="/static/player.css">
</head>
<body>
  <main class="page">
    <h2>%s</h2>

    <video id="video" controls autoplay muted></video>

    <div class="controls">
      <label for="quality">Quality</label>
      <select id="quality">
        <option value="-1">Auto</option>
      </select>
      <span id="current-quality">Current: Auto</span>
    </div>
  </main>

  <script>
    window.STREAM_PLAYER = {
      source: "%s",
      mode: "%s"
    };
  </script>
  <script src="https://cdn.jsdelivr.net/npm/hls.js@latest"></script>
  <script src="/static/player.js"></script>
</body>
</html>`, title, title, playlist, escapedMode)

	_, _ = w.Write([]byte(page))
}
