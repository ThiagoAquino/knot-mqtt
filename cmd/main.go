package main

import (
	"fmt"
	"github.com/CESARBR/knot-mqtt/internal/gateways/knot"
	"knot-mqtt/pkg/application"
)

func main() {
	//Create and Configure client
	client := application.ConfigureClient()
	defer client.Disconnect(250)

	// Configura o QoS para 2 (entrega exatamente uma vez)
	qos := byte(2)

	// Inscreve-se no tópico "/topico/subtopico" com o QoS configurado
	application.SubscribeTopic(client, qos)

	fmt.Println()

	// Aguarda o sinal para sair (CTRL+C)
	application.EndListen()

}
