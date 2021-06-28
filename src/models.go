package src

import "time"

const (
	StatusNone      = 0 // Special status for curTask, means queue is empty now
	StatusScheduled = 1
	StatusInProcess = 2
	StatusSuccess   = 3
	StatusFailure   = 4
)

type ShortInputMetadata struct {
	Format      string `json:"format"`
	Size        int    `json:"size"`
	VideoCodec  string `json:"video_codec"`
	Height      int    `json:"height"`
	Width       int    `json:"width"`
	AspectRatio string `json:"aspect_ratio"`

	FPS int `json:"fps"`

	AudioCodec string `json:"audio_codec"`
	Duration   int    `json:"duration"`
}

type Task struct {
	ID         int    `json:"id"`
	ChannelID  int    `json:"channel_id"`
	PlaylistID int    `json:"playlist_id"`
	Input      string `json:"input"`

	Status int `json:"status"`

	ScheduledAt time.Time `json:"scheduled_at"`
	StartedAt   time.Time `json:"started_at"`
	FinishedAt  time.Time `json:"finished_at"`

	InputMeta *ShortInputMetadata `json:"input_meta,omitempty"`
}

type InProcess struct {
	IsRunning bool `json:"is_running"`
	Task      `json:"task"`
}

type Stats struct {
	Length int `json:"length"`

	InProcess `json:"in_process"`
	Waiting   []Task `json:"waiting"`
	Finished  []Task `json:"finished"`
}

type ProbeOutput struct {
	Streams []struct {
		Index              int    `json:"index"`
		CodecName          string `json:"codec_name"`
		CodecLongName      string `json:"codec_long_name"`
		Profile            string `json:"profile,omitempty"`
		CodecType          string `json:"codec_type"`
		CodecTimeBase      string `json:"codec_time_base"`
		CodecTagString     string `json:"codec_tag_string"`
		CodecTag           string `json:"codec_tag"`
		Width              int    `json:"width,omitempty"`
		Height             int    `json:"height,omitempty"`
		CodedWidth         int    `json:"coded_width,omitempty"`
		CodedHeight        int    `json:"coded_height,omitempty"`
		HasBFrames         int    `json:"has_b_frames,omitempty"`
		SampleAspectRatio  string `json:"sample_aspect_ratio,omitempty"`
		DisplayAspectRatio string `json:"display_aspect_ratio,omitempty"`
		PixFmt             string `json:"pix_fmt,omitempty"`
		Level              int    `json:"level,omitempty"`
		ColorRange         string `json:"color_range,omitempty"`
		ColorSpace         string `json:"color_space,omitempty"`
		ColorTransfer      string `json:"color_transfer,omitempty"`
		ColorPrimaries     string `json:"color_primaries,omitempty"`
		ChromaLocation     string `json:"chroma_location,omitempty"`
		FieldOrder         string `json:"field_order,omitempty"`
		Refs               int    `json:"refs,omitempty"`
		IsAvc              string `json:"is_avc,omitempty"`
		NalLengthSize      string `json:"nal_length_size,omitempty"`
		RFrameRate         string `json:"r_frame_rate"`
		AvgFrameRate       string `json:"avg_frame_rate"`
		TimeBase           string `json:"time_base"`
		StartPts           int    `json:"start_pts"`
		StartTime          string `json:"start_time"`
		BitsPerRawSample   string `json:"bits_per_raw_sample,omitempty"`
		Disposition        struct {
			Default         int `json:"default"`
			Dub             int `json:"dub"`
			Original        int `json:"original"`
			Comment         int `json:"comment"`
			Lyrics          int `json:"lyrics"`
			Karaoke         int `json:"karaoke"`
			Forced          int `json:"forced"`
			HearingImpaired int `json:"hearing_impaired"`
			VisualImpaired  int `json:"visual_impaired"`
			CleanEffects    int `json:"clean_effects"`
			AttachedPic     int `json:"attached_pic"`
			TimedThumbnails int `json:"timed_thumbnails"`
		} `json:"disposition"`
		SampleFmt     string `json:"sample_fmt,omitempty"`
		SampleRate    string `json:"sample_rate,omitempty"`
		Channels      int    `json:"channels,omitempty"`
		ChannelLayout string `json:"channel_layout,omitempty"`
		BitsPerSample int    `json:"bits_per_sample,omitempty"`
		Tags          struct {
			Language string `json:"language"`
			Title    string `json:"title"`
		} `json:"tags,omitempty"`
		DurationTs int    `json:"duration_ts,omitempty"`
		Duration   string `json:"duration,omitempty"`
	} `json:"streams"`
	Format struct {
		Filename       string `json:"filename"`
		NbStreams      int    `json:"nb_streams"`
		NbPrograms     int    `json:"nb_programs"`
		FormatName     string `json:"format_name"`
		FormatLongName string `json:"format_long_name"`
		StartTime      string `json:"start_time"`
		Duration       string `json:"duration"`
		Size           string `json:"size"`
		BitRate        string `json:"bit_rate"`
		ProbeScore     int    `json:"probe_score"`
		Tags           struct {
			Encoder      string    `json:"encoder"`
			CreationTime time.Time `json:"creation_time"`
		} `json:"tags"`
	} `json:"format"`
}

type Rendition struct {
	Height int
	Width  int
}
