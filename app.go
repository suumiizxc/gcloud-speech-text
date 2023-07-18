package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	speech "cloud.google.com/go/speech/apiv1"
	"google.golang.org/api/option"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

// Get the sample rate from the WAV file header.
func getSampleRate(file *os.File) (uint32, error) {
	// Set the byte offset of the sample rate in the WAV file header.
	sampleRateOffset := int64(24) // Sample rate is at byte 24 in the WAV file header.

	// Seek to the byte offset of the sample rate.
	_, err := file.Seek(sampleRateOffset, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to seek to sample rate: %v", err)
	}

	// Read 4 bytes from the sample rate offset.
	sampleRateBytes := make([]byte, 4)
	_, err = file.Read(sampleRateBytes)
	if err != nil {
		return 0, fmt.Errorf("failed to read sample rate bytes: %v", err)
	}

	// Convert the little-endian byte slice to a uint32 value.
	sampleRate := binary.LittleEndian.Uint32(sampleRateBytes)

	return sampleRate, nil
}

func main() {
	start := time.Now()
	ctx := context.Background()

	if len(os.Args) < 2 {
		fmt.Println("No argument provided.")
		return
	}

	// audioFilePath := "mars1.wav"

	audioFilePath := os.Args[1]
	fmt.Println("Audio file: ", audioFilePath)

	// Open the WAV file.
	file, err := os.Open(audioFilePath)
	if err != nil {
		log.Fatalf("Failed to open WAV file: %v", err)
	}
	defer file.Close()

	// Read and parse the sample rate from the WAV file header.
	sampleRate, err := getSampleRate(file)
	if err != nil {
		log.Fatalf("Failed to get sample rate: %v", err)
	}

	// Print the sample rate.
	fmt.Println("Sample Rate (Hz):", sampleRate)

	// Read the audio file content.
	audioFileContent, err := ioutil.ReadFile(audioFilePath)
	if err != nil {
		log.Fatalf("Failed to read audio file: %v", err)
	}

	// Creates a Speech-to-Text client.
	client, err := speech.NewClient(ctx, option.WithCredentialsFile("chatkey.json"))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Start a long-running recognition operation using the audio content.
	resp, err := client.LongRunningRecognize(ctx, &speechpb.LongRunningRecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:        speechpb.RecognitionConfig_LINEAR16,
			SampleRateHertz: int32(sampleRate), // Update the sample rate to match the WAV file.
			LanguageCode:    "en-US",
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Content{Content: audioFileContent},
		},
	})
	if err != nil {
		log.Fatalf("Failed to start long-running recognition: %v", err)
	}

	// Get the long-running operation result.
	result, err := resp.Wait(ctx)
	if err != nil {
		log.Fatalf("Failed to get long-running result: %v", err)
	}

	fmt.Println()
	// Prints the results.
	for _, result := range result.Results {
		for _, alt := range result.Alternatives {
			fmt.Printf("\"%v\" (confidence=%.2f)\n", alt.Transcript, alt.Confidence)
		}
	}

	fmt.Println()
	end := time.Now()
	duration := end.Sub(start)
	fmt.Println("Time duration : ", duration)
}
