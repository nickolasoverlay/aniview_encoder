package src

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const randomPosterQuantity = 3
const fpsBitrateFactor = 1.5
const highBitrateFrom = 48

var input = make(chan<- Task)
var output = make(<-chan Task)
var inQueue []Task
var curTask = Task{Status: StatusNone}

var ratioMap = map[string][]Rendition{
	"16:9": {
		{Height: 1440, Width: 2560},
		{Height: 1080, Width: 1920},
		{Height: 720, Width: 1280},
		{Height: 480, Width: 854},
		{Height: 360, Width: 640},
		{Height: 240, Width: 426},
	},
}

var bitrateMap = map[Rendition]int{
	{Height: 1440, Width: 2560}: 16 * 1024,
	{Height: 1080, Width: 1920}: 8 * 1024,
	{Height: 720, Width: 1280}:  5 * 1024,
	{Height: 480, Width: 854}:   2.5 * 1024,
	{Height: 360, Width: 640}:   1 * 1024,
	{Height: 240, Width: 426}:   0.5 * 1024,
}

func getRenditionBitrate(r Rendition, fps int) int {
	bitrate := bitrateMap[r]

	if fps <= highBitrateFrom {
		return bitrate
	}

	return int(math.Round(float64(bitrate) * fpsBitrateFactor))
}

func Init() {
	input, output = MakeUnboundedQueue()

	go taskProcessor()
}

func Schedule(w http.ResponseWriter, r *http.Request) {
	t := Task{
		Status:      StatusScheduled,
		ScheduledAt: time.Now(),
	}
	json.NewDecoder(r.Body).Decode(&t)

	input <- t
}

func taskProcessor() {
	for {
		curTask = <-output

		curTask.Status = StatusInProcess
		curTask.StartedAt = time.Now()
		time.Sleep(time.Second * 1)

		meta := readMetadata(curTask.Input)
		curTask.InputMeta = &meta

		renditions := getRenditions(curTask.InputMeta.Width, curTask.InputMeta.Height, curTask.InputMeta.AspectRatio)
		// fmt.Println(renditions)

		posterCommands := getPosterCommands(curTask)
		// fmt.Println(posterCommands)
		executeFFCommands(curTask.Input, posterCommands)

		thumbnailCommands := getThumbnailCommands(curTask)
		// fmt.Println(thumbnailCommands)
		executeThumbnailCommands(curTask, thumbnailCommands)

		hlsCommands := getHLSCommands(curTask, renditions)
		executeFFCommands(curTask.Input, hlsCommands)

		saveMasterPlaylist(curTask, renditions)

		curTask.FinishedAt = time.Now()
		curTask = Task{Status: StatusNone}
	}
}

func readMetadata(input string) ShortInputMetadata {
	var stdOut bytes.Buffer
	var stdErr bytes.Buffer

	c := "-v quiet -print_format json -show_format -show_streams"
	p := strings.TrimSpace(input)

	args := strings.Split(c, " ")
	args = append(args, p)

	cmd := exec.Command("/usr/bin/ffprobe", args...)
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	err := cmd.Run()
	if err != nil {
		fmt.Println(stdErr.String())
		fmt.Println(stdOut.String())
		log.Fatal(err)
		return ShortInputMetadata{}
	}

	probeOutput := ProbeOutput{}
	m := ShortInputMetadata{}

	json.Unmarshal(stdOut.Bytes(), &probeOutput)

	size, _ := strconv.Atoi(probeOutput.Format.Size)
	videCodec := probeOutput.Streams[0].CodecName
	height := probeOutput.Streams[0].Height
	width := probeOutput.Streams[0].Width
	aspectRatio := probeOutput.Streams[0].DisplayAspectRatio

	fpsString := probeOutput.Streams[0].RFrameRate
	fpsSplit := strings.Split(fpsString, "/")
	fpsFirstNum, _ := strconv.ParseFloat(fpsSplit[0], 32)
	fpsSecondNum, _ := strconv.ParseFloat(fpsSplit[1], 32)
	fps := int(fpsFirstNum / fpsSecondNum)

	audioCodec := probeOutput.Streams[1].CodecName
	duration, _ := strconv.ParseFloat(probeOutput.Format.Duration, 32)

	m.Format = probeOutput.Format.FormatLongName
	m.Size = size
	m.VideoCodec = videCodec
	m.Height = height
	m.Width = width
	m.AspectRatio = aspectRatio
	m.FPS = fps
	m.AudioCodec = audioCodec
	m.Duration = int(math.Round(duration))

	return m
}

func getRenditions(width int, height int, aspectRatio string) []Rendition {
	resultRenditions := []Rendition{}

	for _, r := range ratioMap[aspectRatio] {
		if r.Height <= height {
			resultRenditions = append(resultRenditions, r)
		}
	}

	return resultRenditions
}

func getHLSCommands(task Task, renditions []Rendition) []string {
	commands := []string{}

	homePath := GetEncoderEnv().OutputPath
	folder := fmt.Sprintf("%d", task.ID)
	hlsTime := 4

	for _, r := range renditions {
		s := ""

		s += "-hide_banner -y" + " "
		if task.InputMeta.AudioCodec != "aac" {
			s += "-c:a aac" + " "
		}
		if task.InputMeta.VideoCodec != "h264" {
			s += "-c:v h264" + " "
		}

		s += "-sn" + " "
		s += "-profile:v main" + " "
		s += "-crf 25" + " "
		s += "-r " + fmt.Sprintf("%d", task.InputMeta.FPS) + " "
		s += fmt.Sprintf("-force_key_frames expr:if(isnan(prev_forced_n),1,eq(n,prev_forced_n+%d))", hlsTime) + " "
		s += "-pix_fmt yuv420p" + " "
		s += "-movflags +faststart" + " "
		s += "-sc_threshold 0" + " "
		s += "-vf scale=" + fmt.Sprintf("%d", r.Width) + ":-2" + " "
		s += "-b:v " + fmt.Sprintf("%d", getRenditionBitrate(r, task.InputMeta.FPS)) + " "
		s += "-f hls" + " "
		s += fmt.Sprintf("-hls_time %d", hlsTime) + " "
		s += "-hls_playlist_type vod" + " "
		s += "-hls_segment_filename "
		s += homePath + "/" + folder + "/" + fmt.Sprintf("%d", r.Height) + "_%03d.ts" + " " + homePath + "/" + folder + "/" + fmt.Sprintf("%d", r.Height) + "_playlist.m3u8"

		commands = append(commands, s)
	}

	return commands
}

func executeFFCommands(input string, commands []string) {
	for _, c := range commands {
		p := strings.TrimSpace(input)

		args := strings.Split(c, " ")
		args = append(args, "-i")
		args = append(args, p)

		fmt.Println("ffmpeg", strings.Join(args, " "))
		fmt.Println("")

		cmd := exec.Command("/usr/bin/ffmpeg", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			log.Fatal("")
			return
		}
	}
}

func getRandomTimeAsString(max int) string {
	lower := int(float32(max) * 0.2)
	upper := int(float32(max) * 0.8)

	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(upper-lower) + lower

	modTime := time.Now().Round(0).Add(-(time.Duration(r) * time.Second))
	since := time.Since(modTime)

	h := since / time.Hour
	since -= h * time.Hour

	m := since / time.Minute
	since -= m * time.Minute

	s := since / time.Second

	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func getPosterCommands(task Task) []string {
	commands := []string{}
	posterBasePath := fmt.Sprintf("%s/%d", GetEncoderEnv().OutputPath, task.ID)

	for i := 0; i < randomPosterQuantity; i++ {
		posterTime := getRandomTimeAsString(task.InputMeta.Duration)

		s := ""
		s += "-hide_banner -y" + " "
		s += "-vframes 1" + " "

		jpgPoster := s + fmt.Sprintf("%s/poster_%d.jpg", posterBasePath, i+1)
		jpgPoster += " -ss " + posterTime

		webpPoster := s + fmt.Sprintf("%s/poster_%d.webp", posterBasePath, i+1)
		webpPoster += " -ss " + posterTime

		commands = append(commands, jpgPoster)
		commands = append(commands, webpPoster)
	}

	return commands
}

// Run after executePosterCommands
func getThumbnailCommands(task Task) []string {
	commands := []string{}
	posterBasePath := fmt.Sprintf("%s/%d", GetEncoderEnv().OutputPath, task.ID)

	for i := 0; i < randomPosterQuantity; i++ {
		s := ""
		s += "-hide_banner -y" + " "
		s += "-s 240x135 -frames:v 1" + " "

		jpgThumb := s + fmt.Sprintf("%s/thumbnail_%d.jpg", posterBasePath, i+1)
		webpThumb := s + fmt.Sprintf("%s/thumbnail_%d.webp", posterBasePath, i+1)

		commands = append(commands, jpgThumb)
		commands = append(commands, webpThumb)
	}

	return commands
}

func executeThumbnailCommands(task Task, commands []string) {
	posterExts := []string{"jpg", "webp"}
	posterBasePath := fmt.Sprintf("%s/%d", GetEncoderEnv().OutputPath, task.ID)
	posterIndexes := []int{1, 1, 2, 2, 3, 3}

	for i, c := range commands {
		posterPath := fmt.Sprintf("%s/poster_%d.%s", posterBasePath, posterIndexes[i], posterExts[i%2])
		executeFFCommands(posterPath, []string{c})
	}
}

func saveMasterPlaylist(task Task, renditions []Rendition) string {
	p := "#EXTM3U\n"
	p += "#EXT-X-VERSION:5\n\n"

	for _, r := range renditions {
		p += fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%dx%d\n%d_playlist.m3u8\n", getRenditionBitrate(r, task.InputMeta.FPS)*1000, r.Width, r.Height, r.Height)
	}

	ioutil.WriteFile(fmt.Sprintf("%s/%d/master_playlist.m3u8", GetEncoderEnv().OutputPath, task.ID), []byte(p), 0777)

	return p
}
