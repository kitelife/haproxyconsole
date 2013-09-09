// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Modify by linuz.ly
package sshoperation
 
import (
    "fmt"
    "io/ioutil"
    "code.google.com/p/go.crypto/ssh"
    "log"
)
var (
    server = "192.168.2.193:36000"
    username = "root"
    password = clientPassword("xxx")
)
type clientPassword string
func (p clientPassword) Password(user string) (string, error) {
    return string(p), nil
}
 
func ScpHaproxyConf() {
    // An SSH client is represented with a slete). Currently only
    // the "password" authentication method is supported.
    //
    // To authenticate with the remote server you must pass at least one
    // implementation of ClientAuth via the Auth field in ClientConfig.
 
    config := &ssh.ClientConfig{
        User: username,
        Auth: []ssh.ClientAuth{
            // ClientAuthPassword wraps a ClientPassword implementation
            // in a type that implements ClientAuth.
            ssh.ClientAuthPassword(password),
        },
    }
    client, err := ssh.Dial("tcp", server, config)
    if err != nil {
        panic("Failed to dial: " + err.Error())
    }
 
    // Each ClientConn can support multiple interactive sessions,
    // represented by a Session.
    defer client.Close()
    // Create a session
    session, err := client.NewSession()
    if err != nil {
        log.Fatalf("unable to create session: %s", err)
    }
    defer session.Close()

    confBytes, err := ioutil.ReadFile("/usr/local/haproxy/conf/haproxy.conf")
    if err != nil {
        panic("Failed to run: " + err.Error())
    }
    content := string(confBytes)
    go func() {
        w, _ := session.StdinPipe()
        defer w.Close()
        fmt.Fprintln(w, "C0644", len(content), "testfile")
        fmt.Fprint(w, content)
        fmt.Fprint(w, "\x00")
    }()
    if err := session.Run("/usr/bin/scp -tq /usr/local/haproxy/conf/haproxy.conf && /usr/local/haproxy/restart_haproxy.sh"); err != nil {
        panic("Failed to run: " + err.Error())
    }
    return
}
