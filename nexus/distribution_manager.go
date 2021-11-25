package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

var localhost string
var AZ1 bool = true
var AZ2 bool = true

func sender(ip string, puerto string, data string, ch chan string, remoteCon net.Conn) {
	con, err := net.Dial("tcp", ip+":"+puerto)
	if err != nil {
		fmt.Println("Error al conectar", err)
	}
	defer con.Close()
	fmt.Fprintln(con, data)
	r := bufio.NewReader(con)
	resp, _ := r.ReadString('\n')
	fmt.Println("Estamos por enviar la respuesta ", resp)
	fmt.Fprintln(remoteCon, resp)
	ch <- string(resp)
}

func distributionManager(port string, con net.Conn, data string, ch1 chan string, ch2 chan string) {
	// Leemos lo que llega de la conexión con los nodos
	// Si la comunicación es por el puerto 9090, entonces se envia a nodo 1 o nodo 2 dependiendo si está ocupado o no
	if port == "9090" {
		// Enviar a nodo disponible
		// usar este para devolver la data de los nodos
		if AZ1 {
			AZ1 = false
			fmt.Println("Se distribuye")
			go sender(localhost, "9095", data, ch1, con) // una vez terminado se debe enviar esta respuesta al backend
			time.Sleep(time.Second * 2)
			AZ1 = true

		} else if AZ2 {
			AZ2 = false
			ch2 := make(chan string)
			go sender(localhost, "9096", data, ch2, con) // una vez terminado se debe enviar esta respuesta al backend
			fmt.Fprintln(con, <-ch2)
			AZ2 = true
		}
	}

}

func receiver(ip string, puerto string) {
	// Función de escucha del centro de distribución
	// Se mantiene escuchando request del backend
	ln, err := net.Listen("tcp", ip+":"+puerto)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	for {
		con, err := ln.Accept()
		fmt.Println("Connection accepted", con.RemoteAddr())
		if err != nil {
			log.Fatal(err)
		}
		go senderConnectionHandler(con)
	}

}

// Envia las peticiones del backend al nodo disponible
func senderConnectionHandler(con net.Conn) {
	defer con.Close()
	ch1 := make(chan string)
	ch2 := make(chan string)
	// Leemos lo que llega de la conexión con los nodos
	bufferI := bufio.NewReader(con)
	data, _ := bufferI.ReadString('\n')
	// Extraer puerto del local address y distribuir las cargas dependiendo de eso
	_, port, err := net.SplitHostPort(con.LocalAddr().String())
	if err != nil {
		log.Fatal(err)
	}
	// Se distribuyen las conexiones que llegan a los nodos
	go distributionManager(port, con, data, ch1, ch2)

	if <-ch1 != "" {
		fmt.Fprintln(con, <-ch1)
	} else if <-ch2 != "" {
		fmt.Fprintln(con, <-ch2)
	}

}

func myIp() string {
	ifaces, err := net.Interfaces()
	// Manejador err
	if err != nil {
		log.Print(fmt.Errorf("localAddres: %v \n", err.Error()))
		return "127.0.0.1"
	}

	for _, iface := range ifaces {
		if strings.HasPrefix(iface.Name, "Ethernet") {
			addrs, err := iface.Addrs()
			// Manejador err
			if err != nil {
				log.Print(fmt.Errorf("localAddres: %v \n", err.Error()))
				return "127.0.0.1"
			}

			for _, addr := range addrs {
				switch d := addr.(type) {
				case *net.IPNet:
					if strings.HasPrefix(d.IP.String(), "192") {
						return d.IP.String()
					}
				}
			}
		}
	}
	return "127.0.0.1"
}

func main() {
	//configuracion
	localhost = myIp()
	// Escucha en el backend
	go receiver(localhost, "9090")
	fmt.Scanf("Enter")
}
