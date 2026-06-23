const video = document.getElementById("video");
const qualitySelect = document.getElementById("quality");
const currentQuality = document.getElementById("current-quality");

const source = window.STREAM_PLAYER.source;
const mode = window.STREAM_PLAYER.mode;

function setCurrentQualityLabel(text) {
  currentQuality.textContent = `Current: ${text}`;
}

function levelLabel(level, index) {
  if (level.height) return `${level.height}p`;
  if (level.bitrate) return `${Math.round(level.bitrate / 1000)} kbps`;
  return `Level ${index}`;
}

if (Hls.isSupported()) {
  const hls = new Hls({
    lowLatencyMode: false,
  });

  hls.loadSource(source);
  hls.attachMedia(video);

  hls.on(Hls.Events.MANIFEST_PARSED, () => {
    hls.levels.forEach((level, index) => {
      const option = document.createElement("option");
      option.value = String(index);
      option.text = levelLabel(level, index);
      qualitySelect.appendChild(option);
    });

    if (mode === "vod") {
      video.currentTime = 0;
    }

    video.play().catch(() => {});
  });

  qualitySelect.addEventListener("change", () => {
    const selectedLevel = Number(qualitySelect.value);

    hls.currentLevel = selectedLevel;

    if (selectedLevel === -1) {
      setCurrentQualityLabel("Auto");
      return;
    }

    const level = hls.levels[selectedLevel];
    setCurrentQualityLabel(levelLabel(level, selectedLevel));
  });

  hls.on(Hls.Events.LEVEL_SWITCHED, (_, data) => {
    const selectedLevel = Number(qualitySelect.value);

    if (selectedLevel === -1) {
      const level = hls.levels[data.level];
      setCurrentQualityLabel(`Auto (${levelLabel(level, data.level)})`);
    }
  });

  hls.on(Hls.Events.ERROR, (_, data) => {
    console.error("hls error", data);
  });
} else if (video.canPlayType("application/vnd.apple.mpegurl")) {
  video.src = source;
  qualitySelect.disabled = true;
  setCurrentQualityLabel("Native HLS");

  video.addEventListener("loadedmetadata", () => {
    if (mode === "vod") {
      video.currentTime = 0;
    }

    video.play().catch(() => {});
  });
} else {
  document.body.insertAdjacentHTML(
    "beforeend",
    "<p>HLS is not supported in this browser.</p>"
  );
}