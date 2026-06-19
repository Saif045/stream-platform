package httpapi

import (
	"fmt"
	"html"
	"net/http"
)

func (s *Server) watchLive(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	writeWatchPage(w, id, "live", "/streams/"+id+"/live/master.m3u8")
}

func (s *Server) watchVOD(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	writeWatchPage(w, id, "vod", "/streams/"+id+"/vod/master.m3u8")
}
func writeWatchPage(w http.ResponseWriter, id string, mode string, playlistURL string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	title := html.EscapeString(id + " - " + mode)
	playlist := html.EscapeString(playlistURL)

	page := fmt.Sprintf(`<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>%s</title>
</head>
<body style="background:#111;color:#eee;font-family:sans-serif;">
  <h2>%s</h2>

  <video id="video" controls autoplay muted style="width:80vw;max-width:1100px;background:#000;"></video>

  <script src="https://cdn.jsdelivr.net/npm/hls.js@latest"></script>
  <script>
    const video = document.getElementById("video");
    const source = "%s";
    const mode = "%s";

    if (Hls.isSupported()) {
      const hls = new Hls({
        lowLatencyMode: false
      });

      hls.loadSource(source);
      hls.attachMedia(video);

      hls.on(Hls.Events.MANIFEST_PARSED, () => {
        if (mode === "vod") {
          video.currentTime = 0;
        }

        video.play().catch(() => {});
      });
    } else if (video.canPlayType("application/vnd.apple.mpegurl")) {
      video.src = source;

      video.addEventListener("loadedmetadata", () => {
        if (mode === "vod") {
          video.currentTime = 0;
        }

        video.play().catch(() => {});
      });
    } else {
      document.body.insertAdjacentHTML("beforeend", "<p>HLS is not supported in this browser.</p>");
    }
  </script>
</body>
</html>`, title, title, playlist, mode)

	_, _ = w.Write([]byte(page))
}
