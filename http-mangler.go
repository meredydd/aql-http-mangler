package main

import ("bufio"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strings"
)

func handleConnection(s net.Conn) {
	local, err := net.Dial("tcp", "localhost:80")

	if err != nil {
		s.Close()
		fmt.Println("Could not connect to localhost:80; ",err);
		return
	}

	close := func() { s.Close(); local.Close(); }

	sr := bufio.NewReader(s)
	lr := bufio.NewReader(local)

	go func() {
		done := false
		for ; !done ; {
			line, err := sr.ReadString('\n')
			if err != nil {
				close()
				return
			}
			if strings.HasPrefix(line, "Connection: ") {
				continue;
			}
			if line == "\n" || line == "\r\n" {
				line = "Connection: close\n\n";
				done = true;
			}
			io.Copy(local, strings.NewReader(line))
		}

		io.Copy(local, sr)
		close()
	}()

	go func() {
		line, err := lr.ReadString('\n')
		matched, _ := regexp.MatchString("^HTTP/... 200 .*", line);
		if err != nil || !matched {
			close()
			return
		}

		io.Copy(s, strings.NewReader(line));

		io.Copy(s, lr)
		close()
	}()
}

func main() {

	ss, err := net.Listen("tcp", ":8080");

	if err != nil {
		fmt.Println("Could not listen on port 8080; ",err);
		os.Exit(1);
	}

	for {
		s, err := ss.Accept();
		if err != nil {
			fmt.Println("Failed to accept connection; ", err);
			continue;
		}
		go handleConnection(s);
	}
}
