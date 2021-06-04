package util

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func getAllGroupNames(group string) ([]string, error) {
	// given a group of {c2|payload}, get all of them that exist within the loaded config
	viper.SetConfigName("docker-compose")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(GetCwdFromExe())
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("[-] Error while reading in docker-compose file: %s", err)
			return []string{}, err
		} else {
			fmt.Printf("[-] Error while parsing docker-compose file: %s", err)
			return []string{}, err
		}
	}
	servicesSub := viper.Sub("services")
	services := servicesSub.AllSettings()
	var absPath string
	var err error
	if group == "c2" {
		absPath, err = filepath.Abs(filepath.Join(GetCwdFromExe(), "C2_Profiles"))
		if err != nil {
			fmt.Printf("[-] failed to get the absolute path to the C2_Profiles folder")
			return []string{}, err
		}
	} else if group == "payload" {
		absPath, err = filepath.Abs(filepath.Join(GetCwdFromExe(), "Payload_Types"))
		if err != nil {
			fmt.Printf("[-] failed to get the absolute path to the C2_Profiles folder")
			return []string{}, err
		}
	}
	var containerList []string
	for container, _ := range services {
		build := servicesSub.GetString(container + ".build")
		buildAbsPath, err := filepath.Abs(build)
		if err != nil {
			fmt.Printf("[-] failed to get the absolute path to the container's docker file")
			continue
		}
		if strings.HasPrefix(buildAbsPath, absPath) {
			// the service we're looking at has a build path that's a child of our folder, it should be a service
			containerList = append(containerList, container)
		}
	}
	return containerList, nil
}

func getMythicEnvList() []string {
	env := viper.AllSettings()
	var envList []string
	for key, _ := range env {
		val := viper.GetString(key)
		if val != "" {
			// prevent trying to append arrays or dictionaries to our environment list
			//fmt.Println(strings.ToUpper(key), val)
			envList = append(envList, strings.ToUpper(key)+"="+val)
		}
	}
	return envList
}

func runDockerCompose(args []string) error {
	path, err := exec.LookPath("docker-compose")
	if err != nil {
		log.Fatalf("[-] docker-compose is not installed or not available in the current PATH variable")
	}
	exe, err := os.Executable()
	if err != nil {
		log.Fatalf("[-] Failed to get path to current executable")
	}
	exePath := filepath.Dir(exe)
	command := exec.Command(path, args...)
	command.Dir = exePath
	command.Env = getMythicEnvList()

	stdout, err := command.StdoutPipe()
	if err != nil {
		log.Fatalf("[-] Failed to get stdout pipe for running docker-compose")
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		log.Fatalf("[-] Failed to get stderr pipe for running docker-compose")
	}

	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)
	go func() {
		for stdoutScanner.Scan() {
			fmt.Printf("%s\n", stdoutScanner.Text())
		}
	}()
	go func() {
		for stderrScanner.Scan() {
			fmt.Printf("%s\n", stderrScanner.Text())
		}
	}()
	err = command.Start()
	if err != nil {
		log.Fatalf("[-] Error trying to start docker-compose: %v\n", err)
	}
	err = command.Wait()
	if err != nil {
		fmt.Printf("[-] Error from docker-compose: %v\n", err)
		return err
	}
	return nil
}

func StartMythic(containerNames []string) error {
	viper.WriteConfig()
	fmt.Printf("[+] Successfully updated configuration in .env\n")

	if len(containerNames) > 0 {
		runDockerCompose(append([]string{"up", "--build", "-d"}, containerNames...))
	} else {
		runDockerCompose([]string{"down", "--volumes", "--remove-orphans"})
		RabbitmqReset()
		err := checkPorts()
		if err != nil {
			return err
		}
		c2ContainerList, err := getAllGroupNames("c2")
		if err != nil {
			fmt.Printf("[-] Failed to get all c2 services: %v\n", err)
			return err
		}
		c2ContainerList = removeExclusionsFromSlice("c2", c2ContainerList)
		payloadContainerList, err := getAllGroupNames("payload")
		if err != nil {
			fmt.Printf("[-] Failed to get all payload services: %v\n", err)
			return err
		}
		payloadContainerList = removeExclusionsFromSlice("payload", payloadContainerList)
		finalList := append(mythicServices, c2ContainerList...)
		finalList = append(finalList, payloadContainerList...)
		runDockerCompose(append([]string{"up", "--build", "-d"}, finalList...))
	}

	Status()
	return nil
}

func StopMythic(containerNames []string) {
	if len(containerNames) > 0 {
		runDockerCompose(append([]string{"rm", "-s", "-v", "-f"}, containerNames...))
	} else {
		runDockerCompose(append([]string{"down", "--volumes", "--remove-orphans"}, containerNames...))
	}
	Status()
}

func StartC2(containerNames []string) error {
	if len(containerNames) == 0 {
		containerList, err := getAllGroupNames("c2")
		if err != nil {
			fmt.Printf("[-] Failed to get all c2 services: %v\n", err)
			return err
		}
		if len(containerList) == 0 {
			fmt.Printf("[*] No C2 Profiles currently registered. Try installing an agent or using t  he add subcommand\n")
			return nil
		}
		if len(containerList) > 0 {
			listWithoutExclusions := removeExclusionsFromSlice("c2", containerList)
			if len(listWithoutExclusions) == 0 {
				fmt.Printf("[*] All selected c2 profiles are in the exclusion list.\n")
				fmt.Printf("[*]   clear the list with: c2 config set excluded_c2_profiles ''\n")
				return nil
			}
			runDockerCompose(append([]string{"up", "--build", "-d"}, listWithoutExclusions...))
		}
	} else {
		runDockerCompose(append([]string{"rm", "-s", "-v", "-f"}, containerNames...))
		if len(containerNames) > 0 {
			runDockerCompose(append([]string{"up", "--build", "-d"}, containerNames...))
		}
	}
	return nil
}

func StopC2(containerNames []string) error {
	if len(containerNames) == 0 {
		containerList, err := getAllGroupNames("c2")
		if err != nil {
			fmt.Printf("[-] Failed to get all c2 services: %v\n", err)
			return err
		}
		if len(containerList) == 0 {
			fmt.Printf("[*] No C2 Profiles currently registered. Try installing an agent or using t  he add subcommand\n")
			return nil
		}
		runDockerCompose(append([]string{"rm", "-s", "-v", "-f"}, containerList...))
	}
	return nil
}

func StartPayload(containerNames []string) error {
	// we're looking at the payload type services
	if len(containerNames) == 0 {
		containerList, err := getAllGroupNames("payload")
		if err != nil {
			fmt.Printf("[-] Failed to get all payload services: %v\n", err)
			return err
		}
		if len(containerList) == 0 {
			fmt.Printf("[*] No Payloads currently registered. Try installing an agent or using the   add subcommand\n")
			return nil
		}
		if len(containerList) > 0 {
			listWithoutExclusions := removeExclusionsFromSlice("payload", containerList)
			if len(listWithoutExclusions) == 0 {
				fmt.Printf("[*] All selected payloads are in the exclusion list.\n")
				fmt.Printf("[*]   clear the list with: payload config set excluded_c2_profiles ''\n")
				return nil
			}
			runDockerCompose(append([]string{"up", "--build", "-d"}, listWithoutExclusions...))
		}
	} else {
		runDockerCompose(append([]string{"rm", "-s", "-v", "-f"}, containerNames...))
		if len(containerNames) > 0 {
			runDockerCompose(append([]string{"up", "--build", "-d"}, containerNames...))
		}
	}
	return nil
}

func StopPayload(containerNames []string) error {
	// we're looking at the payload type services
	if len(containerNames) == 0 {
		containerList, err := getAllGroupNames("payload")
		if err != nil {
			fmt.Printf("[-] Failed to get all payload services: %v\n", err)
			return err
		}
		if len(containerList) == 0 {
			fmt.Printf("[*] No Payloads currently registered. Try installing an agent or using the   add subcommand\n")
			return nil
		}
		if len(containerList) > 0 {
			runDockerCompose(append([]string{"rm", "-s", "-v", "-f"}, containerList...))
		}
	}
	return nil
}

func ListGroupEntries(group string) {
	// list out which group entities exist in the docker-compose file
	dockerComposeEntries, err := getAllGroupNames(group)
	if err != nil {
		log.Fatalf("Failed to get group names from docker-compose: %v\n", err)
	}
	fmt.Printf("Docker-compose entries:\n")
	for _, entry := range dockerComposeEntries {
		fmt.Printf("[+] %s\n", entry)
	}
	var exclusion_list []string
	if group == "c2" {
		exclusion_list = strings.Split(viper.GetString("EXCLUDED_C2_PROFILES"), ",")
	} else if group == "payload" {
		exclusion_list = strings.Split(viper.GetString("EXCLUDED_PAYLOAD_TYPES"), ",")
	}
	if len(exclusion_list) > 0 && exclusion_list[0] != "" {
		fmt.Printf("Excluded entries from .env and environment variables:\n")
		for _, entry := range exclusion_list {
			fmt.Printf("[-] %s\n", entry)
		}

	}
	// list out which group entities exist on disk, which could be different than what's in the d  ocker-compose file
	var targetFolder string
	var groupName string
	if group == "c2" {
		targetFolder = "C2_Profiles"
		groupName = "C2 Profiles"
	} else {
		targetFolder = "Payload_Types"
		groupName = "Payload Types"
	}
	files, err := ioutil.ReadDir(filepath.Join(GetCwdFromExe(), targetFolder))
	if err != nil {
		fmt.Printf("[-] Failed to list contents of %s folder\n", targetFolder)
		return
	}
	fmt.Printf("\n%s on disk:\n", groupName)
	for _, f := range files {
		if f.IsDir() {
			fmt.Printf("[+] %s\n", f.Name())
		}
	}
	// list out which group entities are running
}

func AddRemoveDockerComposeEntries(action string, group string, names []string) error {
	// add c2/payload [name] as type [group] to the main yaml file
	type dockerComposeStruct struct {
		Build          string
		Network_mode   string
		Hostname       string
		Labels         map[string]string
		Container_name string
		Logging        map[string]interface{}
		Restart        string
		Volumes        []string
		Environment    []string
	}
	viper.SetConfigName("docker-compose")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(GetCwdFromExe())
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalf("[-] Error while reading in docker-compose file: %s", err)
		} else {
			log.Fatalf("[-] Error while parsing docker-compose file: %s", err)
		}
	}
	for _, payload := range names {
		if action == "add" {
			var absPath string
			var err error
			if group == "payload" {
				absPath, err = filepath.Abs(filepath.Join(GetCwdFromExe(), "Payload_Types", payload))
				if err != nil {
					fmt.Printf("[-] failed to get the absolute path to the Payload_Types folder")
					continue
				}
			} else if group == "c2" {
				absPath, err = filepath.Abs(filepath.Join(GetCwdFromExe(), "C2_Profiles", payload))
				if err != nil {
					fmt.Printf("[-] failed to get the absolute path to the Payload_Types folder")
					continue
				}
			}
			if !dirExists(absPath) {
				fmt.Printf("[-] %s does not exist, not adding to Mythic\n", absPath)
				continue
			}
			pStruct := dockerComposeStruct{
				Build:        absPath,
				Network_mode: "host",
				Labels: map[string]string{
					"name": payload,
				},
				Hostname:       payload,
				Container_name: payload,
				Logging: map[string]interface{}{
					"driver": "json-file",
					"options": map[string]string{
						"max-file": "1",
						"max-size": "10m",
					},
				},
				Restart: "always",
				Volumes: []string{
					absPath + ":/Mythic/",
				},
				Environment: []string{
					"MYTHIC_ADDRESS=http://${MYTHIC_SERVER_HOST}:${MYTHIC_SERVER_PORT}/api/v1.4/agent_mes  sage",
					"MYTHIC_WEBSOCKET=ws://${MYTHIC_SERVER_HOST}:${MYTHIC_SERVER_PORT}/ws/agent_message",
					"MYTHIC_USERNAME=${RABBITMQ_USER}",
					"MYTHIC_PASSWORD=${RABBITMQ_PASSWORD}",
					"MYTHIC_VIRTUAL_HOST=${RABBITMQ_VHOST}",
					"MYTHIC_HOST=${RABBITMQ_HOST}",
				},
			}
			viper.Set("services."+strings.ToLower(payload), pStruct)
			viper.WriteConfig()
			fmt.Println("[+] Successfully updated docker-compose.yml")
		} else if action == "remove" {
			// remove all entries from yaml file that are in `names`
			for _, payload := range names {
				if !stringInSlice(payload, mythicServices) {
					containerName := []string{payload}
					// Disregard errors
					StopC2(containerName)
					StopPayload(containerName)
					delete(viper.Get("services").(map[string]interface{}), strings.ToLower(payload))
				}
			}
			viper.WriteConfig()
			fmt.Println("[+] Successfully updated docker-compose.yml")
		}
	}
	return nil
}
