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
	"strings"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// yeetCmd represents the yeet command
var yeetCmd = &cobra.Command{
	Use:   "yeet",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		registry, _ := cmd.Flags().GetString("registry")
		creds, _ := cmd.Flags().GetString("creds")
		yeetToReg(registry, creds)
	},
}

func init() {
	rootCmd.AddCommand(yeetCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// yeetCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// yeetCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	yeetCmd.Flags().String("registry", "", "the registry to push everything to")
	yeetCmd.MarkFlagRequired("registry")
	yeetCmd.Flags().String("creds", "", "the credentials to authenticate with the registry base64 encoded")
}

func yeetToReg(registry string, creds string) {
	var rcs resources
	yamlFile, err := os.ReadFile("objects.yml")
	if err != nil {
		log.Fatal("Unable to read file.")
	}

	err = yaml.Unmarshal(yamlFile, &rcs)
	if err != nil {
		log.Fatal("Unable to parse file. Is it in the yaml format?")
	}

	cli, err := client.NewClientWithOpts()
	if err != nil {
		log.Fatal(err)
	}
	body := loadImages(cli)
	fmt.Print(body)
	Images := tagImages(rcs.Images, registry, cli)
	var pushOptions image.PushOptions
	if creds != "" {
		pushOptions = image.PushOptions{RegistryAuth: creds}
	} else {
		pushOptions = image.PushOptions{}
	}
	pushImages(Images, cli, body, pushOptions)

	cli.Close()
}

func loadImages(cli *client.Client) []byte {
	ctx := context.Background()
	file, err := os.OpenFile("images.tar", os.O_RDONLY, 0666)
	if err != nil {
		log.Printf("Error loading image %s, %s", "images.tar", err)
	}
	response, err := cli.ImageLoad(ctx, file, true)
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal("Unable to read response")
	}
	ctx.Done()
	return body
}

func tagImages(Images []string, registry string, cli *client.Client) []string {
	ctx := context.Background()
	var newImages []string
	for i := 0; i < len(Images); i++ {
		fmt.Printf("Retagging image: %s\n", Images[i])
		targetTag := changeTag(Images[i], registry)
		err := cli.ImageTag(ctx, Images[i], targetTag)
		if err != nil {
			log.Fatal(err)
		} else {
			newImages = append(newImages, targetTag)
		}
	}
	fmt.Print("Done tagging images\n")
	ctx.Done()
	return newImages
}

func changeTag(tag string, registry string) string {
	index := strings.Index(tag, "/")
	if index != -1 {
		// Slice the string from the character after the separator to the end
		result := tag[index+1:]
		tag = registry + "/" + result
	} else {
		tag = registry + "/" + tag
	}
	return tag
}

func pushImages(Images []string, cli *client.Client, body []byte, pushOptions image.PushOptions) {
	ctx := context.Background()
	fmt.Print(body)
	for i := 0; i < len(Images); i++ {
		fmt.Printf("Pushing image: %s\n", Images[i])
		reader, err := cli.ImagePush(ctx, Images[i], pushOptions)
		if err != nil {
			log.Fatal(err)
		}
		io.Copy(os.Stdout, reader)
		reader.Close()
	}
	ctx.Done()
}
