package util

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func runGitClone(args []string) error {
	path, err := exec.LookPath("git")
	if err != nil {
		fmt.Printf("[-] git is not installed or not available in the current PATH variable")
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		fmt.Printf("[-] Failed to get path to current executable")
		return err
	}
	exePath := filepath.Dir(exe)
	// git -c http.sslVerify=false clone --recurse-submodules --single-branch --branch $2 $1 temp

	command := exec.Command(path, args...)
	command.Dir = exePath
	command.Env = getMythicEnvList()

	stdout, err := command.StdoutPipe()
	if err != nil {
		fmt.Printf("[-] Failed to get stdout pipe for running git")
		return err
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		fmt.Printf("[-] Failed to get stderr pipe for running git")
		return err
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
		fmt.Printf("[-] Error trying to start git: %v\n", err)
		return err
	}
	err = command.Wait()
	if err != nil {
		fmt.Printf("[-] Error trying to run git: %v\n", err)
		return err
	}
	return nil
}

// https://blog.depa.do/post/copy-files-and-directories-in-go
func copyDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = copyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = copyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

// https://blog.depa.do/post/copy-files-and-directories-in-go
func copyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

// https://gist.github.com/r0l1/3dcbb0c8f6cfe9c66ab8008f55f8f28b
func askConfirm(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s [y/n]: ", prompt)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Failed to read user input")
			return false
		}
		input = strings.ToLower(strings.TrimSpace(input))
		if input == "y" || input == "yes" {
			return true
		} else if input == "n" || input == "no" {
			return false
		}
	}
}

func InstallFolder(installPath string, force bool) error {
	workingPath := GetCwdFromExe()
	overWrite := force
	if FileExists(filepath.Join(installPath, "config.json")) {
		var config = viper.New()
		config.SetConfigName("config")
		config.SetConfigType("json")
		fmt.Printf("[*] Parsing config.json\n")
		config.AddConfigPath(installPath)
		if err := config.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				fmt.Printf("[-] Error while reading in config file: %s", err)
				return err
			} else {
				fmt.Printf("[-] Error while parsing config file: %s", err)
				return err
			}
		}
		if !config.GetBool("exclude_payload_type") {
			// handle the payload type copying here
			files, err := ioutil.ReadDir(filepath.Join(installPath, "Payload_Type"))
			if err != nil {
				fmt.Printf("[-] Failed to list contents of new Payload_Type folder: %v\n", err)
				return err
			}
			for _, f := range files {
				if f.IsDir() {
					fmt.Printf("[*] Processing Payload Type %s\n", f.Name())
					if dirExists(filepath.Join(workingPath, "Payload_Types", f.Name())) {
						if overWrite || askConfirm(f.Name()+" already exists. Replace current version? ") {
							fmt.Printf("[*] Stopping current container\n")
							StartPayload([]string{f.Name()})
							fmt.Printf("[*] Removing current version\n")
							err = os.RemoveAll(filepath.Join(workingPath, "Payload_Types", f.Name()))
							if err != nil {
								fmt.Printf("[-] Failed to remove current version: %v\n", err)
								fmt.Printf("[-] Continuing to the next payload\n")
								continue
							} else {
								fmt.Printf("[+] Successfully removed the current version\n")
							}
						} else {
							fmt.Printf("[!] Skipping Payload Type, %s\n", f.Name())
							continue
						}
					}
					fmt.Printf("[*] Copying new version into place\n")
					err = copyDir(filepath.Join(installPath, "Payload_Type", f.Name()), filepath.Join(workingPath, "Payload_Types", f.Name()))
					if err != nil {
						fmt.Printf("[-] Failed to copy directory over: %v\n", err)
						continue
					}
					// need to make sure the payload_service.sh file is executable
					if FileExists(filepath.Join(workingPath, "Payload_Types", f.Name(), "mythic", "payload_service.sh")) {
						err = os.Chmod(filepath.Join(workingPath, "Payload_Types", f.Name(), "mythic", "payload_service.sh"), 0777)
						if err != nil {
							fmt.Printf("[-] Failed to make payload_service.sh file executable\n")
							continue
						}
					} else {
						fmt.Printf("[-] failed to find payload_service file for %s\n", f.Name())
						continue
					}
					//find ./Payload_Types/ -name "payload_service.sh" -exec chmod +x {} \;
					// now add payload type to yaml config
					fmt.Printf("[*] Adding payload into docker-compose\n")
					AddRemoveDockerComposeEntries("add", "payload", []string{f.Name()})
				}
			}
			fmt.Printf("[+] Successfully installed agent\n")
		}
		if !config.GetBool("exclude_c2_profiles") {
			// handle the c2 profile copying here
			files, err := ioutil.ReadDir(filepath.Join(installPath, "C2_Profiles"))
			if err != nil {
				fmt.Printf("[-] Failed to list contents of C2_Profiles folder from clone\n")
				return err
			}
			for _, f := range files {
				if f.IsDir() {
					fmt.Printf("[*] Processing C2 Profile %s\n", f.Name())
					if dirExists(filepath.Join(workingPath, "C2_Profiles", f.Name())) {
						if overWrite || askConfirm(f.Name()+" already exists. Replace current version? ") {
							fmt.Printf("[*] Stopping current container\n")
							StopC2([]string{f.Name()})
							fmt.Printf("[*] Removing current version\n")
							err = os.RemoveAll(filepath.Join(workingPath, "C2_Profiles", f.Name()))
							if err != nil {
								fmt.Printf("[-] Failed to remove current version: %v\n", err)
								fmt.Printf("[-] Continuing to the next c2 profile\n")
								continue
							} else {
								fmt.Printf("[+] Successfully removed the current version\n")
							}
						} else {
							fmt.Printf("[!] Skipping C2 Profile, %s\n", f.Name())
							continue
						}
					}
					fmt.Printf("[*] Copying new version into place\n")
					err = copyDir(filepath.Join(installPath, "C2_Profiles", f.Name()), filepath.Join(workingPath, "C2_Profiles", f.Name()))
					if err != nil {
						fmt.Printf("[-] Failed to copy directory over\n")
						continue
					}
					// now add payload type to yaml config
					fmt.Printf("[*] Adding c2, %s, into docker-compose\n", f.Name())
					AddRemoveDockerComposeEntries("add", "c2", []string{f.Name()})
				}
			}
			fmt.Printf("[+] Successfully installed c2\n")
		}
		if !config.GetBool("exclude_documentation_payload") {
			// handle payload documentation copying here
			files, err := ioutil.ReadDir(filepath.Join(installPath, "documentation-payload"))
			if err != nil {
				fmt.Printf("[-] Failed to list contents of documentation_payload folder from clone\n")
				return err
			}
			for _, f := range files {
				if f.IsDir() {
					fmt.Printf("[*] Processing Documentation for %s\n", f.Name())
					if dirExists(filepath.Join(workingPath, "documentation-docker", "content", "Agents", f.Name())) {
						if overWrite || askConfirm(f.Name()+" documentation already exists. Replace current version? ") {
							fmt.Printf("[*] Removing current version\n")
							err = os.RemoveAll(filepath.Join(workingPath, "documentation-docker", "content", "Agents", f.Name()))
							if err != nil {
								fmt.Printf("[-] Failed to remove current version: %v\n", err)
								fmt.Printf("[-] Continuing to the next payload documentation\n")
								continue
							} else {
								fmt.Printf("[+] Successfully removed the current version\n")
							}
						} else {
							fmt.Printf("[!] Skipping documentation for , %s\n", f.Name())
							continue
						}
					}
					fmt.Printf("[*] Copying new version into place\n")
					err = copyDir(filepath.Join(installPath, "documentation-payload", f.Name()), filepath.Join(workingPath, "documentation-docker", "content", "Agents", f.Name()))
					if err != nil {
						fmt.Printf("[-] Failed to copy directory over\n")
						continue
					}
				}
			}
			fmt.Printf("[+] Successfully installed agent documentation\n")
		}
		if !config.GetBool("exclude_documentation_c2") {
			// handle the c2 documentation copying here
			files, err := ioutil.ReadDir(filepath.Join(installPath, "documentation-c2"))
			if err != nil {
				fmt.Printf("[-] Failed to list contents of documentation_payload folder from clone")
				return err
			}
			for _, f := range files {
				if f.IsDir() {
					fmt.Printf("[*] Processing Documentation for %s\n", f.Name())
					if dirExists(filepath.Join(workingPath, "documentation-docker", "content", "C2 Profiles", f.Name())) {
						if overWrite || askConfirm(f.Name()+" documentation already exists. Replace current version? ") {
							fmt.Printf("[*] Removing current version\n")
							err = os.RemoveAll(filepath.Join(workingPath, "documentation-docker", "content", "C2 Profiles", f.Name()))
							if err != nil {
								fmt.Printf("[-] Failed to remove current version: %v\n", err)
								fmt.Printf("[-] Continuing to the next c2 documentation\n")
								continue
							} else {
								fmt.Printf("[+] Successfully removed the current version\n")
							}
						} else {
							fmt.Printf("[!] Skipping documentation for %s\n", f.Name())
							continue
						}
					}
					fmt.Printf("[*] Copying new version into place\n")
					err = copyDir(filepath.Join(installPath, "documentation-c2", f.Name()), filepath.Join(workingPath, "documentation-docker", "content", "C2 Profiles", f.Name()))
					if err != nil {
						fmt.Printf("[-] Failed to copy directory over\n")
						continue
					}
				}
			}
			fmt.Printf("[+] Successfully installed c2 documentation\n")
		}
		if !config.GetBool("exclude_agent_icons") {
			// handle copying over the agent's svg icons
			files, err := ioutil.ReadDir(filepath.Join(installPath, "agent_icons"))
			if err != nil {
				fmt.Printf("[-] Failed to list contents of agent_icons folder from clone: %v\n", err)
				return err
			}
			for _, f := range files {
				if !f.IsDir() && f.Name() != ".gitkeep" && f.Name() != ".keep" {
					fmt.Printf("[*] Processing agent icon %s\n", f.Name())
					if FileExists(filepath.Join(workingPath, "mythic-docker", "app", "static", f.Name())) {
						if overWrite || askConfirm(f.Name()+" agent icon already exists. Replace current version? ") {
							fmt.Printf("[*] Removing current version\n")
							err = os.RemoveAll(filepath.Join(workingPath, "mythic-docker", "app", "static", f.Name()))
							if err != nil {
								fmt.Printf("[-] Failed to remove current version: %v\n", err)
								fmt.Printf("[-] Continuing to the next icon\n")
								continue
							} else {
								fmt.Printf("[+] Successfully removed the current version\n")
							}
						} else {
							fmt.Printf("[!] Skipping agent icon for %s\n", f.Name())
							continue
						}
					}
					fmt.Printf("[*] Copying new version into place\n")
					err = copyFile(filepath.Join(installPath, "agent_icons", f.Name()), filepath.Join(workingPath, "mythic-docker", "app", "static", f.Name()))
					if err != nil {
						fmt.Printf("[-] Failed to copy icon over: %v\n", err)
						continue
					}
					if FileExists(filepath.Join(workingPath, "mythic-react-docker", "mythic", "public", f.Name())) {
						if overWrite || askConfirm(f.Name()+" agent icon already exists for new UI. Replace current version? ") {
							fmt.Printf("[*] Removing current version\n")
							err = os.RemoveAll(filepath.Join(workingPath, "mythic-react-docker", "mythic", "public", f.Name()))
							if err != nil {
								fmt.Printf("[-] Failed to remove current version: %v\n", err)
								fmt.Printf("[-] Continuing to the next agent icon\n")
								continue
							} else {
								fmt.Printf("[+] Successfully removed the current version\n")
							}
						} else {
							fmt.Printf("[!] Skipping new UI agent icon for %s\n", f.Name())
							continue
						}
					}
					fmt.Printf("[*] Copying new version into place\n")
					err = copyFile(filepath.Join(installPath, "agent_icons", f.Name()), filepath.Join(workingPath, "mythic-react-docker", "mythic", "public", f.Name()))
					if err != nil {
						fmt.Printf("[-] Failed to copy icon over: %v\n", err)
						continue
					}
				}
			}
			fmt.Printf("[+] Successfully installed agent icons\n")
		}

	} else {
		fmt.Printf("[-] Failed to find config.json in cloned down repo\n")
		return nil
	}
	return nil
}

func InstallAgent(url string, force bool) error {
	// make our temp directory to clone into
	workingPath := GetCwdFromExe()
	fmt.Printf("[*] Creating temporary directory\n")
	err := os.Mkdir(filepath.Join(workingPath, "tmp"), 0755)
	defer os.RemoveAll(filepath.Join(workingPath, "tmp"))
	if err != nil {
		log.Fatalf("[-] Failed to make temp directory for cloning")
	}
	branch := ""
	fmt.Printf("[*] Cloning %s\n", url)
	if branch == "" {
		err = runGitClone([]string{"-c", "http.sslVerify=false", "clone", "--recurse-submodules", "  --single-branch", url, filepath.Join(workingPath, "tmp")})
	} else {
		err = runGitClone([]string{"-c", "http.sslVerify=false", "clone", "--recurse-submodules", "  --single-branch", "--branch", branch, url, filepath.Join(workingPath, "tmp")})
	}
	if err != nil {
		return err
	}
	return InstallFolder(filepath.Join(workingPath, "tmp"), force)
}
