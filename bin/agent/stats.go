package main

func statsRoutine() {
	for {
		select {
		case stats := <-process.StatsQueue():
			log.Println("Process", stats.Messages, stats.Instances, stats.Deliveries)

		case stats := <-gtransport.StatsQueue():
			if stats.BQueue.Total() > 0 {
				log.Println("BcastQ:", stats.BQueue)
			}
			if stats.DQueue.Total() > 0 {
				log.Println("DelivQ:", stats.DQueue)
			}
			if stats.Cache.Added > 0 {
				log.Println("CacheF:", stats.Cache)
			}
			if stats.RQueue.Total() > 0 {
				log.Println("RecevQ:", stats.RQueue)
			}
			if stats.SQueues.Total() > 0 {
				log.Println("SendsQ:", stats.SQueues)
			}
			if stats.Validator.Filtered > 0 {
				log.Println("ValidF:", stats.Validator)
			}
			if stats.MessageLoss.Lost > 0 {
				log.Println("MessageLoss:", stats.MessageLoss)
			}
		}
	}
}
