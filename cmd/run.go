package cmd

import (
	"github.com/spf13/cobra"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		var server string
		var username string
		var password string

		server = os.Getenv("server")
		username = os.Getenv("username")
		password = os.Getenv("password")

		opts := MQTT.NewClientOptions()
		opts.AddBroker(server)
		opts.SetClientID("myid")
		opts.SetCleanSession(true)
		opts.SetUsername(username)
		opts.SetPassword(password)
		opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: false})

		c := MQTT.NewClient(opts)
		if token := c.Connect(); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}

		var data Stuff
		getAWSMetadataAsJson(&data)

		b, err := json.Marshal(data)
		if err != nil {
			log.Println("error:", err)
		}
		js := string(b)

		topic := "aws/ec2/server/dns/" + data.InstanceId

		token := c.Publish(topic, 2, false, js)
		token.Wait()
		c.Disconnect(250)
		os.Stdout.Write([]byte(js))

	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type Stuff struct {
	Ip         string
	Region     string
	InstanceId string
}

func getAWSMetadataAsJson(data *Stuff) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("unable to load config: %w", err)
	}

	client := imds.NewFromConfig(cfg)

	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instancedata-data-retrieval.html

	// ip
	ipRes, err := client.GetMetadata(context.TODO(), &imds.GetMetadataInput{
		Path: "public-ipv4",
	})
	if err != nil {
		log.Fatal("unable to retrieve the ip from the EC2 instance: %w", err)
	}

	defer ipRes.Content.Close()
	ip, err := io.ReadAll(ipRes.Content)
	if err != nil {
		log.Fatal("cannot read ip from the EC2 instance: %w", err)
	}
	data.Ip = string(ip)

	// instance-id
	instanceIdRes, err := client.GetMetadata(context.TODO(), &imds.GetMetadataInput{
		Path: "instance-id",
	})
	if err != nil {
		log.Fatal("unable to retrieve the instanceId from the EC2 instance: %w", err)
	}

	defer instanceIdRes.Content.Close()
	instanceId, err := io.ReadAll(instanceIdRes.Content)
	if err != nil {
		log.Fatal("cannot read instanceId from the EC2 instance: %w", err)
	}
	data.InstanceId = string(instanceId)

	// region
	region, err := client.GetRegion(context.TODO(), &imds.GetRegionInput{})
	if err != nil {
		log.Printf("Unable to retrieve the region from the EC2 instance %v\n", err)
	}
	data.Region = region.Region
}
