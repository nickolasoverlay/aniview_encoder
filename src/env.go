package src

import (
	"log"
	"os"
	"os/user"
	"sync"
)

type encoderEnv struct {
	OutputPath string
}

var envSingleton *encoderEnv
var once sync.Once

func createFolderIfNotExists(path string) {
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func GetEncoderEnv() *encoderEnv {
	outputPath := os.Getenv("OUTPUT_PATH")

	if outputPath == "" {
		user, err := user.Current()
		if err != nil {
			log.Fatal("Could not start AniviewEncoder. OUTPUT_PATH was not provided and program was not able to get user info")
		}

		outputPath = user.HomeDir + "/" + "encoded_videos"
	}

	createFolderIfNotExists(outputPath)

	once.Do(func() {
		envSingleton = &encoderEnv{
			OutputPath: outputPath,
		}
	})

	return envSingleton
}
