package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEurekaBaseURL(environment string, baseDomain string, appName string) string {
	return "http://eureka." +
		environment + "." +
		baseDomain +
		"/eureka/v2/apps/" +
		appName
}

func BuildRequest(url string) (http.Client, *http.Request) {
	eurekaClient := http.Client{
		Timeout: time.Second * 15, // Maximum of 15 secs
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr,"Unable to build Eureka response")
		panic(errors.New("Unable to build Eureka response"))
	}
	req.Header.Set("Accept", "application/json")
	return eurekaClient, req
}

func GetEurekaResponse(eurekaClient http.Client, req *http.Request) EurekaResponse {
	eurekaResponse := EurekaResponse{}
	res, getErr := eurekaClient.Do(req)
	if getErr != nil {
		fmt.Fprintln(os.Stderr,"Unable to get Eureka response")
		panic(getErr)
	}
	defer res.Body.Close()

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		fmt.Fprintln(os.Stderr,"Unable to read Eureka response")
		panic(readErr)
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()

	err := dec.Decode(&eurekaResponse)
	if err == nil {
		return eurekaResponse
	}

	//normalize those degenerate single instance weirdball responses
	eurekaSingleInstance := EurekaSingleInstanceResponse{}
	dec = json.NewDecoder(bytes.NewReader(body))
	err = dec.Decode(&eurekaSingleInstance)
	if err == nil {

		eurekaResponse.Application.Instances = make([]Instance, 0)
		mi := Instance{
			HostName:       eurekaSingleInstance.Application.Instance.HostName,
			IPAddr:         eurekaSingleInstance.Application.Instance.IPAddr,
			Status:         eurekaSingleInstance.Application.Instance.Status,
			Port:           Port{Value: eurekaSingleInstance.Application.Instance.Port.Value},
			HomePageURL:    eurekaSingleInstance.Application.Instance.HomePageURL,
			HealthCheckURL: eurekaSingleInstance.Application.Instance.HealthCheckURL,
		}
		eurekaResponse.Application.Instances = append(eurekaResponse.Application.Instances, mi)
		return eurekaResponse
	} else {
			fmt.Fprintf(os.Stderr,"Unable to decode single instance Eureka response %+v\n", string(body))
	}

	return eurekaResponse
}

type Port struct {
	Value string `json:"$"`
}
type Instance struct {
	HostName       string `json:"hostName"`
	IPAddr         string `json:"ipAddr"`
	Status         string `json:"status"`
	Port           Port   `json:"port"`
	HomePageURL    string `json:"homePageUrl"`
	HealthCheckURL string `json:"healthCheckUrl"`
}

type EurekaResponse struct {
	Application struct {
		Instances []Instance `json:"instance"`
	} `json:"application"`
}

type EurekaSingleInstanceResponse struct {
	Application struct {
		Name     string `json:"name"`
		Instance struct {
			HostName string `json:"hostName"`
			IPAddr   string `json:"ipAddr"`
			Status   string `json:"status"`
			Port     struct {
				Enabled string `json:"@enabled"`
				Value   string `json:"$"`
			} `json:"port"`
			SecurePort struct {
				Enabled string `json:"@enabled"`
				Value   string `json:"$"`
			} `json:"securePort"`
			HomePageURL    string `json:"homePageUrl"`
			HealthCheckURL string `json:"healthCheckUrl"`
		} `json:"instance"`
	} `json:"application"`
}

func getUpInstances(serviceName, environment string) []string {
	baseDomain := getEnv("BASE_DOMAIN", "cirrostratus.org")
	eurekaURL := getEurekaBaseURL(environment, baseDomain, serviceName)

	eurekaClient, req := BuildRequest(eurekaURL)

	eurekaResponse := GetEurekaResponse(eurekaClient, req)

	var ips []string

	for _, instance := range eurekaResponse.Application.Instances {
		if instance.Status == "UP" {
			if instance.HomePageURL != "null" && instance.HomePageURL != "" {
				ips = append(ips, instance.HomePageURL)
			} else {
				ips = append(ips, "http://"+instance.HostName+":"+instance.Port.Value+"/")
			}
		}
	}
	return ips
}

func getSingleUpInstance(serviceName, environment string) string {
	instances := getUpInstances(serviceName, environment)
	if len(instances) == 0 {
		return ""
	}
	rand.Seed(time.Now().Unix())
	return instances[rand.Intn(len(instances))]
}


// no args please
func getFlags() []string {
	var args []string
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-"){
			args = append(args,arg)
		}
	}
	return args
}

// no flags please
func getArgs() []string {
	var args []string
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-"){
			continue
		}
		args = append(args,arg)
	}
	return args
}

func main() {

	listFlag := flag.Bool("all", false, "list all instead of single item")

	flag.CommandLine.Parse(getFlags())
	environment := "test" //default
	argsWithoutProg := getArgs()
	if len(argsWithoutProg) == 0 {
		fmt.Fprintln(os.Stderr,"Usage: lookup <appname> <environment> --list=true")
		panic(errors.New("Usage: lookup <appname> <environment> --list=true"))
	} else if  len(argsWithoutProg) ==2 {
		environment =  argsWithoutProg[1]
	}
	serviceName := argsWithoutProg[0]

	if *listFlag {
		fmt.Println(strings.Join(getUpInstances(serviceName, environment), ","))
	} else {
		fmt.Println(getSingleUpInstance(serviceName, environment))
	}

}
