package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

var localhost string
var remotehost string
var usuariaData DataSet

type Usuaria struct {
	ID        int     `json:"id"`
	Nombre    string  `json:"nombre"`
	DNI       int     `json:"dni"`
	Edad      float64 `json:"edad"`
	Tipo      float64 `json:"tipo"`
	Actividad float64 `json:"actividad"`
	Insumo    float64 `json:"insumo"`
	Metodo    string  `json:"metodo"`
}

type respuestaUsuaria struct {
	ID            string
	Recomendacion string
	Nombre        string
	Edad          string
	TiempoT       string
}
type DataSet struct {
	Usuarias []Usuaria
	Data     [][]interface{}
	Labels   []string
}

//////
func readDataSet() [][]string {
	// Obtener el dataset desde github
	metodoMatrix := [][]string{}
	url := "https://github.com/IPorteniu/TF-Concurrente-202102/raw/main/data/data.csv"
	dataset, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer dataset.Body.Close()

	// Maneja la codificación del archivo si es que hubiera
	br := bufio.NewReader(dataset.Body)
	r, _, err := br.ReadRune()
	if err != nil {
		panic(err)
	}
	if r != '\uFEFF' {
		br.UnreadRune()
	}

	// Leer el dataset
	reader := csv.NewReader(br)
	reader.Comma = ','
	reader.LazyQuotes = true
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		metodoMatrix = append(metodoMatrix, record)
	}

	return metodoMatrix
}

func (ds *DataSet) loadData() {

	// Cargar el DataSet desde su CSV
	data := readDataSet()

	// Inicializar la usuaria Struct para llenarlo con datos
	usuaria := Usuaria{}

	// Almacenar los datos en las estructuras
	for i, metodos := range data {
		// Drop de la primera fila (titles)
		if i == 0 {
			continue
		}

		temp := make([]interface{}, 0)
		// Convertimos los datos necesarios a floats para poder añadirlos
		for j, value := range metodos[:] {

			if j == 6 {
				switch value {
				case "12 a - 17 a":
					usuaria.Edad = 14.5
				case "18 a - 29 a":
					usuaria.Edad = 23.5
				case "30 a - 59 a":
					usuaria.Edad = 44.5
				case "> 60 a":
					usuaria.Edad = 65.0
				}
				temp = append(temp, usuaria.Edad)
			} else if j == 8 {
				// METODO
				usuaria.Metodo = value
			} else if j == 9 {
				// Si son Nuevas = 0 y si son Continuadoras = 1
				switch value {
				case "NUEVAS":
					usuaria.Tipo = 0.0
				case "CONTINUADORAS":
					usuaria.Tipo = 1.0
				}
				// TIPO DE USUARIA
				temp = append(temp, usuaria.Tipo)
			} else if j == 10 {
				parsedValue, err := strconv.ParseFloat(value, 64)
				if err != nil {
					panic(err)
				}
				// ACTIVIDAD
				usuaria.Actividad = parsedValue
				temp = append(temp, usuaria.Actividad)
			} else if j == 11 {
				parsedValue, err := strconv.ParseFloat(value, 64)
				if err != nil {
					panic(err)
				}
				// INSUMO
				usuaria.Insumo = parsedValue
				temp = append(temp, usuaria.Insumo)
			}

		}
		// Filtramos todas las filas que contengan MELA ya que no es un Metodo anticonceptivo que se pueda recomendar normalmente
		if metodos[7] != "MELA" {

			// Añadir los datos al DataSet struct ahora convertidos
			ds.Data = append(ds.Data, temp)
			ds.Labels = append(ds.Labels, metodos[8])
			ds.Usuarias = append(ds.Usuarias, usuaria)
		}
	}
}

//////

var listaUsuaria []Usuaria
var listaRespuestas []respuestaUsuaria

func Routes() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/dataset", MuestraDataSet)
	mux.HandleFunc("/api/agregar", agregarUsuaria)
	log.Fatal(http.ListenAndServe(":9080", mux))
}

func MuestraDataSet(res http.ResponseWriter, req *http.Request) {

	log.Println("llamada al endpoint /dataset")
	jsonBytes, _ := json.Marshal(usuariaData.Usuarias)
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(jsonBytes)

}

func agregarUsuaria(res http.ResponseWriter, req *http.Request) {
	var newUsuaria Usuaria
	if req.Method == "POST" {
		log.Println("Ingreso al metodo agregar")
		cuerpoMsg, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(res, "Error interno al leer el body", http.StatusInternalServerError)
		}
		//fmt.Print("imprimir usuariasJSON")
		json.Unmarshal(cuerpoMsg, &newUsuaria)
		newUsuaria.ID = len(listaUsuaria) + 1
		listaUsuaria = append(listaUsuaria, newUsuaria)
		fmt.Print(newUsuaria)
		go handle(newUsuaria)
		fmt.Printf("%T", newUsuaria)
		json.NewEncoder(res).Encode(newUsuaria)
		res.Header().Set("Content-Type", "application/json")
	}
}

func handle(newUsuaria Usuaria) {
	con, _ := net.Dial("tcp", myIp()+":9090")
	defer con.Close()
	// Codificar JSON
	bytesMsg, err := json.Marshal(newUsuaria)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(con, string(bytesMsg))
	r := bufio.NewReader(con)
	resp, _ := r.ReadString('\n')
	fmt.Printf(resp)
}

func receiver(ip string, puerto string) {
	// receive
	ln, err := net.Listen("tcp", ip+":"+puerto)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	con, err := ln.Accept()
	fmt.Println("Connection accepted", con.LocalAddr())
	if err != nil {
		log.Fatal(err)
	}
	bufferIn := bufio.NewReader(con)
	mensaje, _ := bufferIn.ReadString('\n')
	fmt.Println(mensaje)
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

func connectionHandler(con net.Conn) {
	defer con.Close()
	// Leemos lo que llega de la conexión con los nodos
	bufferI := bufio.NewReader(con)
	data, _ := bufferI.ReadString('\n')
	// Extraer puerto del local address y distribuir las cargas dependiendo de eso
	_, port, err := net.SplitHostPort(con.LocalAddr().String())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(port)
	fmt.Printf(data)
}
func main() {
	usuariaData.loadData()
	localhost = myIp()
	remotehost = "localhost"
	go receiver(localhost, "9001")
	Routes()
}
