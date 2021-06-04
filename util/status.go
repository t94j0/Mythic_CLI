package util

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func Status() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("[-] Failed to get client in status check: %v", err)
	}
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		log.Fatalf("[-] Failed to get container list: %v", err)
	}
	if len(containers) > 0 {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 2, '\t', 0)
		mythic_services := []string{}
		c2_services := []string{}
		payload_services := []string{}
		for _, container := range containers {
			if container.Labels["name"] == "" {
				continue
			}
			info := fmt.Sprintf("%s\t%s\t%s\t", container.Labels["name"], container.State, container.Status)
			if len(container.Ports) > 0 {
				for _, port := range container.Ports {
					if port.PublicPort > 0 {
						info = info + fmt.Sprintf("%d/%s -> %s:%d; ", port.PrivatePort, port.Type, port.IP, port.PublicPort)
					}
				}
			}
			if stringInSlice(container.Image, mythicServices) {
				mythic_services = append(mythic_services, info)
			} else {
				payloadAbsPath, err := filepath.Abs(filepath.Join(GetCwdFromExe(), "Payload_Types"))
				if err != nil {
					fmt.Printf("[-] failed to get the absolute path to the Payload_Types folder")
					continue
				}
				c2AbsPath, err := filepath.Abs(filepath.Join(GetCwdFromExe(), "C2_Profiles"))
				if err != nil {
					fmt.Printf("[-] failed to get the absolute path to the Payload_Types folder")
					continue
				}
				for _, mnt := range container.Mounts {
					if strings.HasPrefix(mnt.Source, payloadAbsPath) {
						payload_services = append(payload_services, info)
					} else if strings.HasPrefix(mnt.Source, c2AbsPath) {
						c2_services = append(c2_services, info)
					}
				}
			}
		}
		fmt.Printf("Mythic Main Services:\n")
		fmt.Fprintln(w, "NAME\tSTATE\tSTATUS\tPORTS")
		for _, line := range mythic_services {
			fmt.Fprintln(w, line)
		}
		w.Flush()
		fmt.Printf("\nPayload Type Services:\n")
		fmt.Fprintln(w, "NAME\tSTATE\tSTATUS\tPORTS")
		for _, line := range payload_services {
			fmt.Fprintln(w, line)
		}
		w.Flush()
		fmt.Printf("\nC2 Profile Services:\n")
		fmt.Fprintln(w, "NAME\tSTATE\tSTATUS\tPORTS")
		for _, line := range c2_services {
			fmt.Fprintln(w, line)
		}
		w.Flush()
	} else {
		fmt.Println("There are no containers running")
	}
}
