package util

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var mythicServices = []string{"mythic_postgres", "mythic_react", "mythic_server", "mythic_redis", "mythic_nginx", "mythic_rabbitmq", "mythic_graphql", "mythic_documentation"}

func GetCwdFromExe() string {
	exe, err := os.Executable()
	if err != nil {
		log.Fatalf("[-] Failed to get path to current executable")
	}
	return filepath.Dir(exe)
}

func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return !info.IsDir()
}

func stringInSlice(value string, list []string) bool {
	for _, e := range list {
		if e == value {
			return true
		}
	}
	return false
}

func GenerateRandomPassword(pw_length int) string {
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	var b strings.Builder
	for i := 0; i < pw_length; i++ {
		nBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			log.Fatalf("[-] Failed to generate random number for password generation\n")
		}
		b.WriteRune(chars[nBig.Int64()])
	}
	return b.String()
}

func removeExclusionsFromSlice(group string, suppliedList []string) []string {
	// use the EXCLUDED_C2_PROFILES and EXCLUDED_PAYLOAD_TYPES variables to limit what we start
	var exclusion_list []string
	if group == "c2" {
		exclusion_list = strings.Split(viper.GetString("EXCLUDED_C2_PROFILES"), ",")
	} else if group == "payload" {
		exclusion_list = strings.Split(viper.GetString("EXCLUDED_PAYLOAD_TYPES"), ",")
	}
	var final_list []string
	for _, element := range suppliedList {
		if !stringInSlice(element, exclusion_list) {
			final_list = append(final_list, element)
		} else {
			fmt.Printf("[*] Skipping %s because it's in an exclusion list\n", element)
		}
	}
	return final_list
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return info.IsDir()
}

func CheckCerts(certPath string, keyPath string) error {
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return err
	} else if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return err
	}
	return nil
}

func GenerateCerts() error {
	if !dirExists(filepath.Join(GetCwdFromExe(), "nginx-docker", "ssl")) {
		err := os.MkdirAll(filepath.Join(GetCwdFromExe(), "nginx-docker", "ssl"), os.ModePerm)
		if err != nil {
			fmt.Printf("[-] Failed to make ssl folder in nginx-docker folder\n")
			return err
		}
		fmt.Printf("[+] Successfully made ssl folder in nginx-docker folder\n")
	}
	certPath := filepath.Join(GetCwdFromExe(), "nginx-docker", "ssl", "mythic-cert.crt")
	keyPath := filepath.Join(GetCwdFromExe(), "nginx-docker", "ssl", "mythic-ssl.key")
	if CheckCerts(certPath, keyPath) == nil {
		fmt.Printf("[+] Mythic certificates already exist\n")
		return nil
	}
	fmt.Printf("[*] Failed to find SSL certs for Nginx container, generating now...\n")
	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		fmt.Printf("[-] failed to generate private key: %s\n", err)
		return err
	}
	notBefore := time.Now()
	oneYear := 365 * 24 * time.Hour
	notAfter := notBefore.Add(oneYear)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		fmt.Printf("[-] failed to generate serial number: %s\n", err)
		return err
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Mythic"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		fmt.Printf("[-] Failed to create certificate: %s\n", err)
		return err
	}
	certOut, err := os.Create(certPath)
	if err != nil {
		fmt.Printf("[-] failed to open "+certPath+" for writing: %s\n", err)
		return err
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Print("failed to open "+keyPath+" for writing:", err)
		return err
	}
	marshalKey, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		fmt.Printf("[-] Unable to marshal ECDSA private key: %v\n", err)
		return err
	}
	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: marshalKey})
	keyOut.Close()
	fmt.Printf("[+] Successfully generated new SSL certs\n")
	return nil
}

func checkPorts() error {
	// go through the different services in viper and check to make sure their ports aren't already used by trying to   open them
	//MYTHIC_SERVER_HOST:MYTHIC_SERVER_PORT
	//POSTGRES_HOST:POSTGRES_PORT
	//HASURA_HOST:HASURA_PORT
	//RABBITMQ_HOST:RABBITMQ_PORT
	//DOCUMENTATION_HOST:DOCUMENTATION_PORT
	//0.0.0.0:REDIS_PORT
	portChecks := map[string]string{
		"MYTHIC_SERVER_HOST": "MYTHIC_SERVER_PORT",
		"POSTGRES_HOST":      "POSTGRES_PORT",
		"HASURA_HOST":        "HASURA_PORT",
		"RABBITMQ_HOST":      "RABBITMQ_PORT",
		"DOCUMENTATION_HOST": "DOCUMENTATION_PORT",
	}
	for key, val := range portChecks {
		if viper.GetString(key) == "127.0.0.1" {
			p, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(viper.GetInt(val)))
			if err != nil {
				fmt.Printf("[-] Port %d appears to already be in use: %v\n", viper.GetInt(val), err)
				return err
			}
			err = p.Close()
			if err != nil {
				fmt.Printf("[-] Failed to close connection: %v\n", err)
				return err
			}
		}
	}
	p, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(viper.GetInt("REDIS_PORT")))
	if err != nil {
		fmt.Printf("[-] Port %d appears to already be in use: %v\n", viper.GetInt("REDIS_PORT"), err)
		return err
	}
	err = p.Close()
	if err != nil {
		fmt.Printf("[-] Failed to close connection: %v\n", err)
		return err
	}
	return nil
}
