package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func (r *Runner) StartLiveHLS(rtmpURL string, outputDir string) (*exec.Cmd, error) {
	hlsDir := filepath.Join(outputDir, "hls")

	for _, quality := range []string{"1080p", "720p", "480p"} {
		if err := os.MkdirAll(filepath.Join(hlsDir, quality), 0755); err != nil {
			return nil, fmt.Errorf("create quality dir: %w", err)
		}
	}

	segmentPattern := filepath.Join(hlsDir, "%v", "segment_%06d.ts")
	variantPlaylist := filepath.Join(hlsDir, "%v", "playlist.m3u8")

	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-i", rtmpURL,

		"-filter_complex",
		"[0:v]split=3[v1080src][v720src][v480src];"+
			"[v1080src]scale=w=1920:h=1080:force_original_aspect_ratio=decrease[v1080];"+
			"[v720src]scale=w=1280:h=720:force_original_aspect_ratio=decrease[v720];"+
			"[v480src]scale=w=854:h=480:force_original_aspect_ratio=decrease[v480]",

		"-map", "[v1080]", "-map", "0:a?",
		"-map", "[v720]", "-map", "0:a?",
		"-map", "[v480]", "-map", "0:a?",

		"-c:v:0", "libx264",
		"-b:v:0", "6000k",
		"-maxrate:v:0", "6500k",
		"-bufsize:v:0", "12000k",

		"-c:v:1", "libx264",
		"-b:v:1", "3000k",
		"-maxrate:v:1", "3500k",
		"-bufsize:v:1", "6000k",

		"-c:v:2", "libx264",
		"-b:v:2", "1200k",
		"-maxrate:v:2", "1500k",
		"-bufsize:v:2", "2400k",

		"-preset", "veryfast",
		"-g", "60",
		"-keyint_min", "60",
		"-sc_threshold", "0",

		"-c:a", "aac",
		"-b:a", "128k",
		"-ac", "2",

		"-f", "hls",
		"-hls_time", "2",
		"-hls_list_size", "0",
		"-hls_flags", "independent_segments",
		"-hls_segment_filename", segmentPattern,
		"-master_pl_name", "master.m3u8",
		"-var_stream_map", "v:0,a:0,name:1080p v:1,a:1,name:720p v:2,a:2,name:480p",

		variantPlaylist,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start ffmpeg: %w", err)
	}

	return cmd, nil
}
