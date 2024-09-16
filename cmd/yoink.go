/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type resources struct {
	Images []string `yaml:"images"`
	Charts []string `yaml:"charts"`
}

// yoinkCmd represents the yoink command
var yoinkCmd = &cobra.Command{
	Use:   "yoink",
	Short: "Pull images and tar them into an images.tar",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fileName, _ := cmd.Flags().GetString("file")
		yoinkFromFile(fileName)
	},
}

func init() {
	rootCmd.AddCommand(yoinkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// yoinkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// yoinkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	yoinkCmd.Flags().String("file", "", "the file to read image names from")
	yoinkCmd.MarkFlagRequired("file")
}

func yoinkFromFile(fileName string) {
	var rcs resources

	yamlFile, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatal("Unable to read file.")
	}

	err = yaml.Unmarshal(yamlFile, &rcs)
	if err != nil {
		log.Fatal("Unable to parse file. Is it in the yaml format?")
	}

	handleImages(rcs.Images)

	for i := 0; i < len(rcs.Charts); i++ {
		fmt.Print(rcs.Charts[i] + "\n")
	}
}

func handleImages(Images []string) {
	cli, err := client.NewClientWithOpts()
	if err != nil {
		log.Fatal(err)
	}
	Images = pullImages(cli, Images)
	saveImages(cli, Images)
	cli.Close()
}

func pullImages(cli *client.Client, Images []string) []string {
	ctx := context.Background()

	for i := 0; i < len(Images); i++ {
		fmt.Printf("Pulling image: %s\n", Images[i])
		reader, err := cli.ImagePull(ctx, Images[i], image.PullOptions{})
		if err != nil {
			fmt.Printf("Unable to Pull image: %s\n", Images[i])
		}
		io.Copy(os.Stdout, reader)
		if err != nil {
			log.Fatal("Error reading from reader:", err)
		}
		reader.Close()
	}
	ctx.Done()
	return Images
}

func saveImages(cli *client.Client, Images []string) {
	ctx := context.Background()
	// Save the image as a tar file
	tarFile, err := os.Create("images.tar")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tarFile.Close()
	fmt.Print(Images)
	reader, err := cli.ImageSave(ctx, Images)
	if err != nil {
		log.Fatal(err)
	}

	_, err = io.Copy(tarFile, reader)
	if err != nil {
		log.Fatal(err)
	}
	reader.Close()
	ctx.Done()
}
