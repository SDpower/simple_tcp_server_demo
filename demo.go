package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"github.com/SDpower/simple_tcp_server_demo/client"
	sts "github.com/SDpower/tcpserver"
	humanize "github.com/vys/go-humanize"
)

type echoAgent struct {
	conn      net.Conn
	reader    *bufio.Reader
	writer    *bufio.Writer
	quit      chan struct{}
	apiClient client.APIClient
}

func (x *echoAgent) OnConnect() {
	log.Printf(`info: %v client connected.`, x.conn.RemoteAddr())
}

func (x *echoAgent) CloseConnect() {
	log.Printf(`info: %v client close connecte.`, x.conn.RemoteAddr())
}

func (x *echoAgent) Proceed() error {
	//timed out second.
	x.conn.SetDeadline(time.Now().Add(time.Second * 20))
	select {
	case <-x.quit:
		return sts.Error(`quit`)
	default:
	}
	line, err := x.reader.ReadBytes('\n')
	if err != nil {
		if err != io.EOF {
			log.Println(`ReadBytes error:`, err)
		}
		return err
	}
	if string(line) == "quit\n" {
		x.conn.Close()
		close(x.quit)
		return nil
	}
	if len(string(line)) > 1 {
		result := []byte(x.apiClient.DoSend(strings.Trim(string(line), "\n")))
		_, err = x.writer.Write(result)
		if err != nil {
			if err != io.EOF {
				log.Println(`Write error:`, err)
			} else {
				log.Printf(`info: %v client closed.`, x.conn.RemoteAddr())
			}
			return err
		}
		err = x.writer.Flush()
		if err != nil {
			if err != io.EOF {
				log.Println(`Flush error:`, err)
			}
			return err
		}
	} else {

	}
	return nil
}

var localAddr = flag.String("server", "localhost:30000", "tcp server address")
var demoType = flag.String("demoType", "server", "demo type server or client")
var totalConnect *int

func main() {
	flag.Parse()
	switch *demoType {
	case "server":
		ServerDemo()
	case "client":
		ClientTest()
	case "apiserver":
		APIServerMock()
	}
}

// APIBasicHelloHandler is
func APIBasicHelloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
		}
		log.Printf(`info: Got message "%v".`, string(body))
		io.WriteString(w, fmt.Sprintf("you send: %v", string(body)))
	} else {
		io.WriteString(w, time.Now().UTC().Format(time.UnixDate))
	}
}

// APIServerMock is simple external API mock
func APIServerMock() {
	log.Println(`info: API Mock Server Listen at :30002.`)
	mux := http.NewServeMux()
	mux.HandleFunc("/", APIBasicHelloHandler)
	http.ListenAndServe(":30002", mux)
}

//ClientTest a tcp client for testing
func ClientTest() {
	c1 := client.New(*localAddr)
	conn, err := net.Dial("tcp", c1.GetAddress())
	if err != nil {
		log.Println(`error:`, err)
		return
	}

	defer conn.Close()

	for {
		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Text to send: ")
		text, _ := reader.ReadString('\n')
		// send to socket
		fmt.Fprintf(conn, text+"\n")
		// listen for reply
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Println(`error:`, err)
			return
		}
		fmt.Print("Message from server: " + message)
	}
}

//ServerDemo simple server demo
func ServerDemo() {
	srv, err := sts.New(*localAddr, newEchoAgent, 100)
	if err != nil {
		log.Println(`error:`, err)
		return
	}
	err = srv.Start()
	if err != nil {
		log.Println(`error:`, err)
		return
	}
	log.Printf(`info: %v server start.`, *localAddr)
	totalConnect = srv.GetTotalConnect()
	log.Printf(`info: %d server connect.`, *totalConnect)

	go func() {
		srv.Wait()
		// close(waitSrv)
		log.Println(`info: server Ready.`)
		//close(waitSrv)
	}()

	go GoRuntimeStats()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	log.Println("CTRL-C; to exiting")

	select {
	// case <-waitSrv:
	case <-c:
		// case <-time.After(time.Second * 300):
		srv.Stop()
		log.Println(`info:`, "server stop!")
		// os.Exit(1)
	}
}

func newEchoAgent(conn net.Conn, reader *bufio.Reader, writer *bufio.Writer, quit chan struct{}) sts.Agent {
	return &echoAgent{conn, reader, writer, quit, *client.NewAPI("http://127.0.0.1:30002/")}
}

//GoRuntimeStats  an http service show stats.
func GoRuntimeStats() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var buffer bytes.Buffer
		m := &runtime.MemStats{}
		buffer.WriteString(fmt.Sprintln("# goroutines: ", runtime.NumGoroutine()))
		runtime.ReadMemStats(m)
		buffer.WriteString(fmt.Sprintln("Memory Acquired: ", humanize.Bytes(m.Sys)))
		buffer.WriteString(fmt.Sprintln("Memory Used    : ", humanize.Bytes(m.Alloc)))
		buffer.WriteString(fmt.Sprintln("# malloc       : ", m.Mallocs))
		buffer.WriteString(fmt.Sprintln("# free         : ", m.Frees))
		buffer.WriteString(fmt.Sprintln("GC enabled     : ", m.EnableGC))
		buffer.WriteString(fmt.Sprintln("# GC           : ", m.NumGC))
		buffer.WriteString(fmt.Sprintln("Last GC time   : ", m.LastGC))
		buffer.WriteString(fmt.Sprintln("Next GC        : ", humanize.Bytes(m.NextGC)))
		buffer.WriteString(fmt.Sprintln("Total connect  : ", *totalConnect))
		io.WriteString(w, buffer.String())
	})
	http.ListenAndServe(":30001", mux)
}
