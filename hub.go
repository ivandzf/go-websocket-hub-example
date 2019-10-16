package main

import (
	"log"
	"time"
)

type hub struct {
	hubs         map[string]*member
	join         chan *member
	delete       chan string
	removeMember chan *member
}

func newHub() *hub {
	return &hub{
		hubs:         make(map[string]*member),
		join:         make(chan *member),
		delete:       make(chan string),
		removeMember: make(chan *member),
	}
}

func (h *hub) run() {
	for {
		select {
		case member := <-h.join:
			if _, ok := h.hubs[member.name]; !ok {
				// join room
				h.hubs[member.name] = member
				log.Printf("%s join hub", member.name)
				go h.process(member, 0)
			} else {
				h.hubs[member.name] = member
			}
		case name := <-h.delete:
			if _, ok := h.hubs[name]; ok {
				log.Printf("deleteting hub %s", name)
				delete(h.hubs, name)
			}
		case member := <-h.removeMember:
			if _, ok := h.hubs[member.name]; ok {
				log.Printf("removing member %s", member.name)
				h.hubs[member.name] = nil
			}
		}
	}
}

func (h *hub) process(member *member, attempt int) {
	// do process
	time.Sleep(5 * time.Second)

	if member := h.hubs[member.name]; member != nil {
		log.Printf("sending to member : %s", member.name)
		member.payload <- "SEND PAYLOAD"
		return
	} else {
		// check if member still nil we must wait member to reconnect
		if attempt < 5 {
			time.Sleep(1 * time.Second)
			attempt++
			h.process(member, attempt)
		} else {
			// when member still disconnect and retry attempt reach max, we must delete the hub
			h.delete <- member.name
		}
	}
}
