package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	var err error
	var cmd *exec.Cmd
	err = sshKeygen()
	if err != nil {
		panic(err)
	}
	err = writeFiles()
	if err != nil {
		panic(err)
	}
	cmd, err = sshServer()
	var cmds = []*exec.Cmd{}
	cmds = append(cmds, cmd)
	err = sshClient()
	for _, cmd = range cmds {
		var err = cmd.Process.Signal(os.Interrupt)
		if err != nil {
			panic(err)
		}
	}
	if err != nil {
		panic(err)
	}
}

func sshKeygen() error {
	var buffer bytes.Buffer
	buffer.WriteString("y\n")
	var cmd = exec.Command("ssh-keygen", "-t", "rsa", "-N", "", "-f", "/root/.ssh/id_rsa")
	cmd.Stdout = os.Stdout
	cmd.Stdin = &buffer
	cmd.Stderr = os.Stderr
	var err = cmd.Run()
	if err != nil {
		return err
	}
	os.Remove("/root/.ssh/authorized_keys")
	return os.Link("/root/.ssh/id_rsa.pub", "/root/.ssh/authorized_keys")
}

func writeFiles() error {
	var env string
	var err error
	for _, env = range os.Environ() {
		var parts = strings.Split(env, "=")
		var key = parts[0]
		if strings.HasPrefix(strings.ToLower(key), "file_") {
			var value = strings.Join(parts[1:], "=")
			value = os.ExpandEnv(value)
			key = getFilename(key)
			err = ioutil.WriteFile(key, []byte(value), 0600)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getFilename(env string) string {
	return env[6:]
}

func sshServer() (*exec.Cmd, error) {
	var cmd = exec.Command("/usr/sbin/sshd", "-D")
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	var err = cmd.Start()
	if err != nil {
		return nil, err
	}
	time.Sleep(100 * time.Millisecond)
	var tries = 10
	for !testport("127.0.0.1", 22) {
		tries--
		time.Sleep(1 * time.Second)
		if tries == 0 {
			err = cmd.Process.Signal(os.Interrupt)
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("port 22 never went up")
		}
	}
	log.Println("server up")
	return cmd, nil
}

func sshClient() error {
	var args = os.Args[1:]
	var cmd = exec.Command("ssh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	log.Printf("%+v", cmd)
	return cmd.Run()
}

func testport(host string, port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 500*time.Millisecond)
	if err, ok := err.(*net.OpError); ok && err.Timeout() {
		return false
	}
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
