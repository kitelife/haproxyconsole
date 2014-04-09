package sshoperation

import (
	"os/exec"
	"config"
	"fmt"
	"io/ioutil"
	"errors"
	"code.google.com/p/go.crypto/ssh"
)

type clientPassword string

func (p clientPassword) Password(user string) (string, error) {
	return string(p), nil
}

func ScpHaproxyConf(appConf config.ConfigInfo) (errinfo error) {

	server := fmt.Sprintf("%s:%d", appConf.SlaveServerIp, appConf.SlaveServerSSHPort)
	username := appConf.SlaveRemoteUser
	password := clientPassword(appConf.SlaveRemotePasswd)

	// An SSH client is represented with a slete). Currently only
	// the "password" authentication method is supported.
	//
	// To authenticate with the remote server you must pass at least one
	// implementation of ClientAuth via the Auth field in ClientConfig.

	conf := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.ClientAuth{
			// ClientAuthPassword wraps a ClientPassword implementation
			// in a type that implements ClientAuth.
			ssh.ClientAuthPassword(password),
		},
	}
	client, err := ssh.Dial("tcp", server, conf)
	if err != nil {
		errinfo = errors.New(fmt.Sprintf("Failed to dial: %s", err.Error()))
		return
	}

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	defer client.Close()
	// Create a session
	session, err := client.NewSession()
	if err != nil {
		errinfo = errors.New(fmt.Sprintf("unable to create session: %s", err.Error()))
		return
	}
	defer session.Close()

	confBytes, err := ioutil.ReadFile(appConf.NewHAProxyConfPath)
	if err != nil {
		errinfo = errors.New(fmt.Sprintf("Failed to run: %s", err.Error()))
		return
	}
	content := string(confBytes)
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintln(w, "C0644", len(content), "new_conf")
		fmt.Fprint(w, content)
		fmt.Fprint(w, "\x00")
	}()
	cmd := fmt.Sprintf("%s -tq %s && %s", appConf.ScpCommandPath, appConf.SlaveConf, appConf.SlaveRestartScript)
	if err := session.Run(cmd); err != nil {
		errinfo = errors.New(fmt.Sprintf("Failed to run: %s", err.Error()))
		return
	}
	return
}

// 主到从已打通信任关系，故无需密码
func ConfidentialScpHAProxyConf(appConf config.ConfigInfo) (errinfo error) {
	scpTarget := fmt.Sprintf("%s@%s:%s", appConf.SlaveRemoteUser, appConf.SlaveServerIp, appConf.SlaveConf)
	scpTargetPortOption := fmt.Sprintf("-P %d", appConf.SlaveServerSSHPort)
	_, err := exec.Command(appConf.ScpCommandPath, scpTargetPortOption, appConf.NewHAProxyConfPath, scpTarget).Output()
	if err != nil {
		errinfo = errors.New(fmt.Sprintf("Failed to remote copy config file : %s", err.Error()))
		return
	}
	sshTarget := fmt.Sprintf("%s@%s", appConf.SlaveRemoteUser, appConf.SlaveServerIp)
	sshTargetPortOption := fmt.Sprintf("-p %d", appConf.SlaveServerSSHPort)
	_, err = exec.Command(appConf.SSHCommandPath, sshTargetPortOption, sshTarget, appConf.SlaveRestartScript).Output()
	if err != nil {
		errinfo = errors.New(fmt.Sprintf("Falied to restart haproxy : %s", err.Error()))
		return
	}
	return
}
