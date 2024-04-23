package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
)

var (
	s3Client *s3.S3
	s3Bucket string
	wg       sync.WaitGroup
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	awsAccessKey := os.Getenv("AWS_ACCESS_KEY")
	awsSecretKey := os.Getenv("AWS_SECRET_KEY")
	session, err := session.NewSession(
		&aws.Config{
			Region: aws.String("sa-east-1"),
			Credentials: credentials.NewStaticCredentials(
				awsAccessKey,
				awsSecretKey,
				"",
			),
		},
	)
	if err != nil {
		panic(err)
	}
	s3Client = s3.New(session)
	s3Bucket = "uploadgo"

}

func main() {
	dir, err := os.Open("./tmp")
	if err != nil {
		panic(err)
	}

	defer dir.Close()
	uploadControl := make(chan struct{}, 200)
	errorFileUpload := make(chan string, 20)

	go func() {
		for fileName := range errorFileUpload {
			uploadControl <- struct{}{}
			wg.Add(1)
			go uploadFile(fileName, uploadControl, errorFileUpload)
		}
	}()

	for {
		files, err := dir.Readdir(1)
		if err != nil || len(files) == 0 || err == io.EOF {
			fmt.Printf("Error reading directory: %s\n", err)
			break
		}
		wg.Add(1)
		uploadControl <- struct{}{}
		go uploadFile(files[0].Name(), uploadControl, errorFileUpload)

	}
	wg.Wait()
}

func uploadFile(fileName string, uploadControl <-chan struct{}, errorFileUpload chan<- string) {
	defer wg.Done()
	completeFileName := fmt.Sprintf("./tmp/%s", fileName)
	fmt.Printf("Uploading file %s to bucket %s ", completeFileName, s3Bucket)
	file, err := os.Open(completeFileName)
	if err != nil {
		fmt.Printf("Error opening file %s\n", completeFileName)
		<-uploadControl
		errorFileUpload <- completeFileName
		return
	}
	defer file.Close()
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(fileName),
		Body:   file,
	})
	if err != nil {
		fmt.Printf("Error opening file %s\n", completeFileName)
		<-uploadControl
		return
	}
	fmt.Printf("File %s uploaded successfully\n", completeFileName)
	<-uploadControl
}
