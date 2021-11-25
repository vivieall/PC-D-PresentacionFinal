package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

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

type DataSet struct {
	Data   [][]interface{}
	Labels []string
}

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

func loadData() DataSet {

	// Cargar el DataSet desde su CSV
	data := readDataSet()
	ds := DataSet{}

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
		}
	}
	return ds
}

func extractFeatures(u Usuaria) [][]interface{} {

	features := [][]interface{}{}

	featureData := []interface{}{u.Edad, u.Tipo, u.Actividad, u.Insumo}
	features = append(features, featureData)

	return features
}

func trainML() *Forest {
	// ENTRENAMIENTO DATASET
	ds := loadData()
	fmt.Println(len(ds.Data))

	forest := TrainForest(ds.Data, ds.Labels, len(ds.Data)/10, len(ds.Data[0]), 50)

	return forest

}

func predictMethod(usuariaJSON Usuaria, forest *Forest) Usuaria {

	features := extractFeatures(usuariaJSON)

	// OUTPUT
	var output string
	for i := 0; i < len(features); i++ {
		output = forest.Predicate(features[i])
	}
	usuariaJSON.Metodo = output

	return usuariaJSON

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

func send(user Usuaria) {
	//Nodo a cual mandar la usuaria
	conn, _ := net.Dial("tcp", myIp()+":9090")
	defer conn.Close()
	// Codificar JSON
	bytesMsg, err := json.Marshal(user)
	if err != nil {
		log.Fatal(err)
	}
	// Enviar mensaje serializado en string
	fmt.Fprintln(conn, string(bytesMsg))

}

func usuariaReceiver(forest *Forest) {
	ln, _ := net.Listen("tcp", myIp()+":9096")
	defer ln.Close()
	for {
		con, _ := ln.Accept()
		go usuariaHandler(con, forest) // podemos atender miles de clienes concurrentemente!
	}

}

func usuariaHandler(con net.Conn, forest *Forest) {
	defer con.Close()
	fmt.Println("Conectado")
	r := bufio.NewReader(con)
	msg, _ := r.ReadString('\n')
	// Decodificamos a la usuaria
	var user Usuaria
	json.Unmarshal([]byte(msg), &user)

	fmt.Println("PRE-ML")
	fmt.Println(user)

	//Predecimos su metodo
	user = predictMethod(user, forest)

	fmt.Println("POST-ML")
	fmt.Println(user)
	fmt.Fprintln(con, user)
	//send(user)

}

func main() {
	forest := trainML()
	go usuariaReceiver(forest)
	fmt.Scanf("Enter")
}
