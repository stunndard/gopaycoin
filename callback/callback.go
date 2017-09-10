package callback

import (
	"log"
	"sync"
	"time"

	"github.com/parnurzeal/gorequest"
	"github.com/stunndard/gopaycoin/model"
)

var (
	m               sync.Mutex
	activeCallbacks []uint
)

func worker(p model.Payment) {
    log.Println("CLB: Worker started, ID:", p.ID)

    event := ""
    if p.Pending {
        event = "pending"
    } else {
        event = "completed"
    }

    defer removeCallbackWorker(p.ID)
	defer log.Println("CLB: Worker ended, ID:", p.ID)
	c, err := model.CreateCallback(&model.Callback{
		Payment: p.ID,
		Url:     p.Callback,
        Event:   event,
	}, p.ID)
	if err != nil {
		log.Println("CLB: DB Error creating callback", err)
		return
	}

	// try to deliver callback up to 5 times
	for i := 1; i < 6; i++ {

		// Sleep interval, Sleep(i * time)
		var pause time.Duration
		switch i {
		case 1:
			pause = 0
		case 2:
			pause = time.Duration(time.Minute) * time.Duration(5)
		case 3:
			pause = time.Duration(time.Minute) * time.Duration(20)
		case 4:
			pause = time.Duration(time.Hour) * time.Duration(1)
		case 5:
			pause = time.Duration(time.Hour) * time.Duration(12)
		}
        log.Println("CLB: Sleeping for", pause.String(), "before next attempt")
		time.Sleep(pause)

		c.Attempts = uint(i)

		log.Println("CLB: Delivering "+event+" callback, ID:", p.ID, "URL:", c.Url, "Attempt:", i)

		request := gorequest.New()
		resp, _, errs := request.Post(c.Url).
			Param("event", event).
			Param("reference", p.Reference).
			Param("custom", p.Custom).
			//TLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
			Timeout(60 * time.Second).
			End()

		if errs != nil {
			// err request
			log.Println("CLB: Error making HTTP request, ID:", c.ID, errs)
			if err := c.Save(); err != nil {
				log.Println("CLB: DB error saving Callback, ID:", c.ID, err)
			}
			continue
		}

		c.LastResponseCode = uint(resp.StatusCode)
		if resp.StatusCode != 200 {
			// error
			log.Println("CLB: Error HTTP response, ID:", p.ID, "http return code:", resp.StatusCode)
			if err := c.Save(); err != nil {
				log.Println("CLB: DB error saving Callback, ID:", p.ID, err)
			}
			continue
		}

		log.Println("CLB: Success Callback ID:", p.ID, "event:", event, "URL:", c.Url)

		c.Success = true
		if err := c.Save(); err != nil {
			log.Println("CLB: DB ERROR updating Callback", err)
		}
		return
	}
    c.Failed = true
    if err := c.Save(); err != nil {
        log.Println("CLB: DB ERROR updating Callback", err)
    }
}

func AddCallbackWorker(p model.Payment) {
    m.Lock()
    defer m.Unlock()
	// if worker is already active
    for i := range activeCallbacks {
        if activeCallbacks[i] == p.ID {
            return
        }
    }

	if len(activeCallbacks) > 100 {
		log.Println("CLB: Cannot add calback worker, already", len(activeCallbacks), "active workers")
		return
	}
	activeCallbacks = append(activeCallbacks, p.ID)
	log.Println("CLB: Added callback worker, active workers:", len(activeCallbacks))

	go worker(p)
}

func removeCallbackWorker(id uint) {
	m.Lock()
	for i := range activeCallbacks {
		if id == activeCallbacks[i] {
			activeCallbacks = append(activeCallbacks[:i], activeCallbacks[i+1:]...)
			break
		}
	}
	m.Unlock()
}
