package httpapi

import (
	"fmt"
	"net/http"
	"strings"
)

const liveWindowSegments = 6

func (s *Server) getLiveMasterPlaylist(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.writeRewrittenMasterPlaylist(w, id, "live")
}

func (s *Server) getVODMasterPlaylist(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.writeRewrittenMasterPlaylist(w, id, "vod")
}

func (s *Server) getLiveVariantPlaylist(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	quality := r.PathValue("quality")

	data, err := s.store.ReadHLSVariantPlaylist(id, quality)
	if err != nil {
		http.Error(w, "playlist not found", http.StatusNotFound)
		return
	}

	playlist := trimPlaylistToLatestSegments(string(data), liveWindowSegments)
	playlist = rewriteSegmentPaths(playlist, id, quality)

	writeM3U8(w, playlist)
}

func (s *Server) getVODVariantPlaylist(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	quality := r.PathValue("quality")

	data, err := s.store.ReadHLSVariantPlaylist(id, quality)
	if err != nil {
		http.Error(w, "playlist not found", http.StatusNotFound)
		return
	}

	playlist := rewriteSegmentPaths(string(data), id, quality)
	playlist = markPlaylistAsEvent(playlist)

	writeM3U8(w, playlist)
}

func (s *Server) writeRewrittenMasterPlaylist(w http.ResponseWriter, id string, mode string) {
	data, err := s.store.ReadHLSMasterPlaylist(id)
	if err != nil {
		http.Error(w, "playlist not found", http.StatusNotFound)
		return
	}

	body := string(data)

	body = strings.ReplaceAll(
		body,
		"1080p/playlist.m3u8",
		"/streams/"+id+"/"+mode+"/1080p/playlist.m3u8",
	)

	body = strings.ReplaceAll(
		body,
		"720p/playlist.m3u8",
		"/streams/"+id+"/"+mode+"/720p/playlist.m3u8",
	)

	body = strings.ReplaceAll(
		body,
		"480p/playlist.m3u8",
		"/streams/"+id+"/"+mode+"/480p/playlist.m3u8",
	)

	writeM3U8(w, body)
}

func rewriteSegmentPaths(playlist string, streamID string, quality string) string {
	lines := strings.Split(playlist, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		if strings.HasPrefix(trimmed, "http://") ||
			strings.HasPrefix(trimmed, "https://") ||
			strings.HasPrefix(trimmed, "/") {
			continue
		}

		lines[i] = "/streams/" + streamID + "/hls/" + quality + "/" + trimmed
	}

	return strings.Join(lines, "\n")
}

func writeM3U8(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}

func trimPlaylistToLatestSegments(playlist string, maxSegments int) string {
	lines := strings.Split(playlist, "\n")

	header := make([]string, 0)
	segments := make([][]string, 0)

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#EXTINF:") {
			if i+1 < len(lines) {
				segmentLine := strings.TrimSpace(lines[i+1])
				segments = append(segments, []string{line, segmentLine})
				i++
			}
			continue
		}

		if strings.HasPrefix(line, "#EXT-X-ENDLIST") {
			continue
		}

		if strings.HasPrefix(line, "#") {
			header = append(header, line)
		}
	}

	originalSegmentCount := len(segments)

	if len(segments) > maxSegments {
		segments = segments[len(segments)-maxSegments:]
	}

	newMediaSequence := findMediaSequence(playlist)
	if originalSegmentCount > maxSegments {
		newMediaSequence += originalSegmentCount - maxSegments
	}

	out := make([]string, 0)

	for _, line := range header {
		if strings.HasPrefix(line, "#EXT-X-MEDIA-SEQUENCE:") {
			out = append(out, fmt.Sprintf("#EXT-X-MEDIA-SEQUENCE:%d", newMediaSequence))
			continue
		}

		out = append(out, line)
	}

	for _, pair := range segments {
		out = append(out, pair...)
	}

	return strings.Join(out, "\n") + "\n"
}

func findMediaSequence(playlist string) int {
	lines := strings.Split(playlist, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "#EXT-X-MEDIA-SEQUENCE:") {
			var seq int
			_, _ = fmt.Sscanf(line, "#EXT-X-MEDIA-SEQUENCE:%d", &seq)
			return seq
		}
	}

	return 0
}
func markPlaylistAsEvent(playlist string) string {
	lines := strings.Split(playlist, "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "#EXT-X-PLAYLIST-TYPE:EVENT" {
			return playlist
		}
	}

	out := make([]string, 0, len(lines)+1)

	for _, line := range lines {
		out = append(out, line)

		if strings.HasPrefix(line, "#EXT-X-VERSION:") {
			out = append(out, "#EXT-X-PLAYLIST-TYPE:EVENT")
		}
	}

	return strings.Join(out, "\n")
}
