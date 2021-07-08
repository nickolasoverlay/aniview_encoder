package src

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var input = make(chan<- Task)
var output = make(<-chan Task)
var inQueue []Task
var curTask = Task{Status: StatusNone}

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
		fmt.Println(renditions)

		hlsCommands := getHLSCommands(curTask, renditions)
		executeHLSCommands(curTask.Input, hlsCommands)

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
	ratioMap := map[string][]Rendition{
		"16:9": {
			{Height: 1440, Width: 2560},
			{Height: 1080, Width: 1920},
			{Height: 720, Width: 1280},
			{Height: 480, Width: 854},
			{Height: 360, Width: 640},
			{Height: 240, Width: 426},
		},
	}

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
	segmentDuration := "4"

	for _, r := range renditions {
		s := ""

		s += "-hide_banner -y" + " "
		if task.InputMeta.AudioCodec != "aac" {
			s += "-c:a aac" + " "
		}
		if task.InputMeta.VideoCodec != "h264" {
			s += "-c:v h264" + " "
		}

		s += "-profile:v main" + " "
		s += "-crf 30" + " "
		s += "-sc_threshold 0" + " "
		s += "-vf scale=" + fmt.Sprintf("%d", r.Height) + ":-2" + " "
		s += "-hls_time " + segmentDuration + " "
		s += "-hls_playlist_type vod" + " "
		s += "-hls_segment_filename "
		s += homePath + "/" + folder + "/" + fmt.Sprintf("%d", r.Height) + "_%03d.ts" + " " + homePath + "/" + folder + "/" + fmt.Sprintf("%d", r.Height) + "playlist.m3u8 "
		s += "-i"
		commands = append(commands, s)
	}

	return commands
}

func executeHLSCommands(input string, commands []string) {
	for _, c := range commands {
		p := strings.TrimSpace(input)

		args := strings.Split(c, " ")
		args = append(args, p)

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
