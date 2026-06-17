1. Capture Layer
   └── OBS / FFmpeg / Camera
       ├── Captures screen/camera/audio
       ├── Encodes video/audio
       └── Sends stream to ingest

2. Contribution / Transport Layer
   └── RTMP first
       ├── Streamer → server protocol
       ├── Carries H.264/AAC packets
       ├── Uses stream key
       └── Later: SRT / WebRTC

3. Ingest Service - Go
   └── Own backend service
       ├── Accept RTMP connection
       ├── Validate stream key
       ├── Create stream session
       ├── Track active streams
       ├── Start/stop FFmpeg worker
       ├── Handle disconnects
       └── Emit logs/metrics

4. Stream Control / API Service - Go
   └── HTTP API
       ├── Create stream key
       ├── Start stream record
       ├── Stop stream record
       ├── Get stream status
       ├── List active streams
       └── Later: users/auth/database

5. Media Processing Layer
   └── FFmpeg
       ├── Read RTMP input
       ├── Remux if input is acceptable
       ├── Transcode if qualities needed
       ├── Generate HLS
       ├── Write playlist.m3u8
       └── Write .ts or .m4s segments

6. Packaging Layer
   └── HLS
       ├── Master playlist
       ├── Variant playlists
       ├── 1080p / 720p / 480p qualities
       ├── Segment duration
       ├── Keyframe alignment
       └── Live playlist updates

7. Storage Layer
   └── Local disk first
       ├── data/streams/{streamID}
       ├── playlist.m3u8
       ├── segments
       └── Later: S3/MinIO/object storage

8. Delivery Service - Go / HTTP
   └── Viewer-facing server
       ├── Serve playlists
       ├── Serve segments
       ├── Correct MIME types
       ├── CORS if needed
       ├── Cache headers
       └── Later: CDN integration

9. Playback Layer
   └── Browser/player
       ├── HTML video
       ├── hls.js for browsers
       ├── Reads playlist
       ├── Downloads segments
       └── Switches quality later

10. Observability
   └── Logs + metrics
       ├── Structured logs
       ├── Active streams
       ├── FFmpeg process status
       ├── Bitrate
       ├── FPS
       ├── CPU/memory
       ├── Dropped frames
       └── Later: Prometheus/Grafana

11. Reliability / Production Hardening
   └── Backend hygiene
       ├── Config files/env vars
       ├── Graceful shutdown
       ├── Worker cleanup
       ├── Timeouts
       ├── Process restart policy
       ├── Disk cleanup
       ├── Rate limits
       └── Error handling

12. Testing / Performance
   └── Quality layer
       ├── Unit tests
       ├── Integration tests
       ├── FFmpeg worker tests
       ├── RTMP ingest tests
       ├── Load tests
       ├── Benchmarks
       └── Profiling



Main build roadmap:

Milestone 1:
Go launches FFmpeg and produces HLS from test.mp4

Milestone 2:
Go viewer server serves HLS correctly

Milestone 3:
RTMP input works from OBS/FFmpeg

Milestone 4:
Ingest service manages stream sessions + FFmpeg workers

Milestone 5:
Stream API + stream keys

Milestone 6:
Multi-quality HLS

Milestone 7:
Metrics, cleanup, tests, Docker

Core stack:

Go
  backend services, APIs, worker management

FFmpeg
  media processing, remuxing, transcoding, HLS packaging

RTMP
  broadcaster ingest

HLS
  viewer delivery

HTTP
  playlist/segment serving

Local disk first
  simple storage

Docker later
  deployment/runtime isolation