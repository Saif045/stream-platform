package vod

type VODStatus string

const (
	VODStatusGrowing VODStatus = "growing"
	VODStatusReady   VODStatus = "ready"
)

type PublicVOD struct {
	ID          string    `json:"id"`
	Status      VODStatus `json:"status"`
	PlaylistURL string    `json:"playlist_url"`
}

type VOD struct {
	PublicVOD
}

func (v *VOD) Public() any {
	if v == nil {
		return nil
	}

	return v.PublicVOD
}
