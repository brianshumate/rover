// Package command for upload
// Upload puts the archive file into an S3 bucket
package command

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/brianshumate/rover/internal"
	"github.com/mitchellh/cli"
	"github.com/ryanuber/columnize"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	archiveFileDefault = "rover.zip"
	archiveFileDescr   = "Archive filename"
)

// UploadCommand describes info upload related fields
type UploadCommand struct {
	AccessKey   string
	ArchiveFile string
	Bucket      string
	HostName    string
	Prefix      string
	Region      string
	SecretKey   string
	Token       string
	UI          cli.Ui
}

// Help output
func (c *UploadCommand) Help() string {
	helpText := `
Usage: rover upload [options]
	Upload an archive to S3 bucket
`

	return strings.TrimSpace(helpText)
}

// Run command
func (c *UploadCommand) Run(args []string) int {

	cmdFlags := flag.NewFlagSet("upload", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.UI.Output(c.Help()) }
	cmdFlags.StringVar(&c.ArchiveFile, "file", archiveFileDefault, archiveFileDescr)
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	// Internal logging
	internal.LogSetup()

	c.AccessKey = os.Getenv("AWS_ACCESS_KEY_ID")
	c.SecretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	c.Bucket = os.Getenv("AWS_BUCKET")
	c.Prefix = os.Getenv("AWS_PREFIX")
	c.Region = os.Getenv("AWS_REGION")
	c.Token = ""

	log.Printf("[i] Hello from the rover upload module on %s!", c.HostName)

	if len(c.AccessKey) == 0 || len(c.SecretKey) == 0 || len(c.Bucket) == 0 || len(c.Region) == 0 {
		log.Println("[e] Missing one of the required AWS environment variables")
		columns := []string{}
		kvs := map[string]string{"AWS_ACCESS_KEY_ID": "Access key ID for AWS", "AWS_SECRET_ACCESS_KEY": "Secret access key ID for AWS", "AWS_BUCKET": " Name of the S3 bucket", "AWS_REGION": "AWS region for the bucket"}
		for k, v := range kvs {
			columns = append(columns, fmt.Sprintf("%s: | %s ", k, v))
		}
		envVars := columnize.SimpleFormat(columns)
		out := fmt.Sprintf("One or more upload related environment variables not set; please ensure that the following environment variables are set:\n\n%s", envVars)
		c.UI.Error(out)

		os.Exit(1)
	}

	creds := credentials.NewStaticCredentials(c.AccessKey, c.SecretKey, c.Token)
	cfg := aws.NewConfig().WithRegion(c.Region).WithCredentials(creds)
	svc := s3.New(session.New(), cfg)

	file, err := os.Open(c.ArchiveFile)
	if err != nil {
		out := fmt.Sprintf("Error opening archive file! Error: %v", err)
		c.UI.Error(out)
		log.Println(out)
		os.Exit(-1)
	}

	defer func() {
		// Close after zip file is successfully uploaded
		err = file.Close()
		if err != nil {
			out := fmt.Sprintf("Could not close file %s! Error: %v", c.ArchiveFile, err)
			c.UI.Error(out)
			os.Exit(-1)
		}
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		out := fmt.Sprintf("Could not stat file %s! Error: %v", c.ArchiveFile, err)
		c.UI.Error(out)
		os.Exit(-1)
	}
	var fileSize int64 = fileInfo.Size()
	buffer := make([]byte, fileSize)

	defer func() {
		// Read from the buffer
		_, err = file.Read(buffer)
		if err != nil {
			out := fmt.Sprintf("Could not read buffer! Error: %s", err)
			log.Println(out)
			c.UI.Error(out)
			os.Exit(-1)
		}
	}()

	path := fmt.Sprintf("%s/%s", c.Prefix, file.Name())
	fileBytes := bytes.NewReader(buffer)
	// For more than application/zip later
	fileType := http.DetectContentType(buffer)
	params := &s3.PutObjectInput{
		Bucket:        aws.String(c.Bucket),
		Key:           aws.String(path),
		Body:          fileBytes,
		ContentLength: aws.Int64(fileSize),
		ContentType:   aws.String(fileType),
	}

	resp, err := svc.PutObject(params)
	if err != nil {
		out := fmt.Sprintf("Bad response from AWS! Response: %s - %s", err, resp)
		c.UI.Error(out)
	}
	out := fmt.Sprintf("Success! Uploaded %s", file.Name())

	c.UI.Output(out)

	return 0
}

// Synopsis output
func (c *UploadCommand) Synopsis() string {
	return "Uploads rover archive file to S3 bucket"
}
