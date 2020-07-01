package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"time"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("connect error: ", err)
	}
	defer conn.Close()

	// a channel session
	channel, err := conn.Channel()
	if err != nil {
		log.Fatal("create a channel error: ", err)
	}
	defer channel.Close()
	// Message durability
	//queue, err := channel.QueueDeclare("task_queue", true, false, false, false, nil)
	//if err != nil {
	//	log.Fatal("Failed to declare a queue")
	//}
	doneChan := make(chan struct{})
	go func() {
		//UseQueue(channel, queue)
		//UseExchange(channel)
		//UseDirectExchange(channel)
		UseTopic(channel)
		doneChan <- struct{}{}
	}()
	<- doneChan
}

func UseQueue(channel *amqp.Channel, queue amqp.Queue) {
	count := 1
	for ; count <= 5; count ++ {
		message := "hello rabbitmq"
		log.Println("send message: ", count)
		// exchange default exchange
		if err := channel.Publish("", queue.Name, false, false, amqp.Publishing{
			ContentType: "text/plain",
			DeliveryMode: amqp.Persistent,
			Body:        []byte(fmt.Sprintf("%d-%s", count, message)),
		}); err != nil {
			log.Fatal("Failed to send message to queue")
		}
		<- time.NewTimer(time.Second * 5).C
	}
}

func UseExchange(channel *amqp.Channel) {
	// kind direct, topic, headers, fanout.
	// fanout broadcast. to all queue it knows.
	err := channel.ExchangeDeclare("logs", "fanout", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("declare a exchange error")
	}
	// TODO test: sudo rabbitmqctl list_exchanges
	body := "hello rabbitmq exchange"
	for cnt := 1; cnt < 3; cnt ++ {
		if err = channel.Publish("logs", "", false, false,amqp.Publishing{
			ContentType:     "text/plain",
			Body:            []byte(body),
		}); err != nil {
			log.Fatalf("Failed to send message to exchange")
		}
		log.Println("[*] send ", body)
		<- time.NewTimer(time.Second*5).C
	}

}

func UseDirectExchange(channel *amqp.Channel) {
	err := channel.ExchangeDeclare("logs_direct", "direct", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("declare a direct exchange error ", err)
	}
	body := "test direct exchange"
	routingKeys := []string{"info", "error"}
	for cnt := 1; cnt < 4; cnt ++ {
		for _, route := range routingKeys {
			msg := fmt.Sprintf("%d --> %s", cnt, route) + body
			if err = channel.Publish("logs_direct", route, false, false, amqp.Publishing{
				ContentType: "text/plain",
				Body: []byte(msg),
			}); err != nil {
				log.Fatalf("Failed to send a message to exchange")
			}
			log.Println("send success: ", msg)
			<- time.NewTimer(time.Second*5).C
		}
	}
}

func UseTopic(channel *amqp.Channel) {
	err := channel.ExchangeDeclare("logs_topic", "topic", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("decleare a topic exchange error: %s", err)
	}
	msgs := []string{"kern.critical", "A critical kernel error"}
	for _, msg := range msgs {
		if err = channel.Publish("logs_topic", msg, false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body: []byte(msg),
		}); err != nil {
			log.Fatalf("send %s error: %s", msg, err)
		}
		log.Printf("send %s success\n", msg)
		<- time.NewTimer(time.Second*2).C
	}
}