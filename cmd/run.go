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

type Stuff struct {
	Ip         string
	Region     string
	InstanceId string
}

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
		var data Stuff

		server := os.Getenv("SERVER")
		if server == "" {
			return errors.New("missing server")
		}

		username := os.Getenv("USERNAME")
		if username == "" {
			return errors.New("missing username")
		}

		password := os.Getenv("PASSWORD")
		if password == "" {
			return errors.New("missing password")
		}

		opts := MQTT.NewClientOptions()
		opts.AddBroker(server)

		getAWSMetadata(&data)
		opts.SetClientID(data.InstanceId)

		opts.SetCleanSession(true)
		opts.SetUsername(username)
		opts.SetPassword(password)
		opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: false})

		c := MQTT.NewClient(opts)
		if token := c.Connect(); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}

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

// ip
func getInstanceIp(client *imds.Client) (string, error) {
	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instancedata-data-retrieval.html
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
	return string(ip), nil
}

// region
func getInstanceRegion(client *imds.Client) (string, error) {
	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instancedata-data-retrieval.html
	region, err := client.GetRegion(context.TODO(), &imds.GetRegionInput{})
	if err != nil {
		log.Printf("Unable to retrieve the region from the EC2 instance %v\n", err)
	}
	return string(region.Region), nil
}

// instance-id
func getInstanceId(client *imds.Client) (string, error) {
	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instancedata-data-retrieval.html

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
	return string(instanceId), nil
}

func getAWSMetadata(data *Stuff) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("unable to load config: %w", err)
	}

	client := imds.NewFromConfig(cfg)

	instanceId, _ := getInstanceId(client)
	data.InstanceId = string(instanceId)

	ip, _ := getInstanceIp(client)
	data.Ip = string(ip)

	region, _ := getInstanceRegion(client)
	data.Region = string(region)
}
