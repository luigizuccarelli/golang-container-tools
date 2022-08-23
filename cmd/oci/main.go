package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/luigizuccarelli/golang-container-tools/pkg/schema"
	"github.com/luigizuccarelli/golang-container-tools/pkg/service"
)

const (
	apiVersion string = "/v2/"
)

var (
	image     string
	version   string
	path      string
	action    string
	tls       string
	basicAuth string
)

func init() {
	flag.StringVar(&image, "i", "", "image url : quay.io/user/component")
	flag.StringVar(&version, "v", "", "version : v0.0.1")
	flag.StringVar(&path, "p", "", "path to copy to: oci")
	flag.StringVar(&action, "a", "", "copy or push")
	flag.StringVar(&tls, "t", "true", "tls verify true (default) or false")
	flag.StringVar(&basicAuth, "b", "false", "basic auth true or false (default)")
}

func main() {

	var reg = schema.ServiceSchema{}

	flag.Parse()

	if image == "" || version == "" || path == "" || action == "" {
		flag.Usage()
		os.Exit(1)
	}

	// set up the struct for the service to use
	reg.Image = image
	reg.Version = version
	tmp := strings.Split(image, "/")
	if len(tmp) == 2 {
		reg.Name = tmp[0]
		reg.User = ""
		reg.Component = tmp[1]
	} else {
		reg.Name = tmp[0]
		reg.User = tmp[1] + "/"
		reg.Component = tmp[2]
	}
	reg.Path = path
	val, err := strconv.ParseBool(tls)
	if err != nil {
		fmt.Println(fmt.Sprintf("ERROR: %v", err))
		os.Exit(1)
	}
	reg.TLS = val
	if !reg.TLS {
		reg.URL = "http://" + reg.Name + apiVersion + reg.User + reg.Component
	} else {
		reg.URL = "https://" + reg.Name + apiVersion + reg.User + reg.Component
	}

	val, err = strconv.ParseBool(basicAuth)
	if err != nil {
		fmt.Println(fmt.Sprintf("ERROR: %v", err))
		os.Exit(1)
	}
	reg.Auth = val

	fmt.Println("INFO: Executing OCI")
	fmt.Println("      Action     : ", action)
	fmt.Println("      Registry   : ", reg.Name)
	fmt.Println("      User       : ", reg.User)
	fmt.Println("      Version    : ", reg.Version)
	fmt.Println("      Path       : ", reg.Path)
	fmt.Println("      URL        : ", reg.URL)
	fmt.Println("      TLS        : ", reg.TLS)
	fmt.Println("      Basic Auth : ", reg.Auth)
	fmt.Println("")

	switch action {
	case "copy":
		err := service.OCICopyToDisk(reg)
		if err != nil {
			fmt.Println(fmt.Sprintf("ERROR: %v", err))
			os.Exit(1)
		}
		fmt.Println("INFO: OCI copy completed successfully")
	case "push":
		err := service.OCIPushToRegistry(reg)
		if err != nil {
			fmt.Println(fmt.Sprintf("ERROR: %v", err))
			os.Exit(1)
		}
		fmt.Println("INFO: OCI push completed successfully")
	default:
		fmt.Println("ERROR: action argument not recognized")
		os.Exit(1)
	}
	os.Exit(0)
}
