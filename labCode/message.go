package d7024e

import(
    "encoding/json"
    "fmt"
    "sync"
    "time"
)

type Message struct {
    Type string `json:"type"`
    Name string `json:"name"`
    RequestID string `json:"requestId"`
    Source string `json:"source"`
    Destination string `json:"destination"`
    Params []string `json:"params"`
}

type MessageRecord struct {
    record map[KademliaID]*func (Message, Contact)
    queue chan messageRecordQueueElem
    ttl time.Duration
    mux sync.Mutex
}

type messageRecordQueueElem struct {
    Deadline time.Time
    ID KademliaID
    dead func()
}

func ParseMessage(buf []byte) (Message, error) {
    var msg Message
    err := json.Unmarshal(buf, &msg)
    if err != nil {
        return msg, err
    }
    if msg.Type != "RPC" && msg.Type != "RETURN" {
        return msg, fmt.Errorf("Unknown message type: %q", msg.Type)
    }
    if  msg.Name != "PING"       && msg.Name != "FIND-NODE" &&
        msg.Name != "FIND-VALUE" && msg.Name != "STORE" {
        return msg, fmt.Errorf("Unknown message name: %q", msg.Name)
    }
    id_ok := hash_rx.MatchString(msg.RequestID)
    src_ok := hash_rx.MatchString(msg.Source)
    dst_ok := hash_rx.MatchString(msg.Destination)
    msg_ok := hash_rx.MatchString(msg.RequestID)
    if !id_ok || !src_ok || !dst_ok || !msg_ok {
        return msg, fmt.Errorf("Invalid hash in message:\n%+v", msg)
    }
    return msg, err
}


func NewMessageRecord() MessageRecord {
    record := MessageRecord{
        record: make(map[KademliaID]*func (Message, Contact)),
        queue: make(chan messageRecordQueueElem, 1024),
        ttl: TTL,
        mux: sync.Mutex{},
    }

    go func() {
        for {
            queueElem := <-record.queue
            ttl := time.Until(queueElem.Deadline)
            if ttl.Nanoseconds() > 0 {
                time.Sleep(ttl)
            }
            record.mux.Lock()
            if _, ok := record.record[queueElem.ID]; ok {
                delete(record.record, queueElem.ID)
                record.mux.Unlock()
                queueElem.dead()
            } else {
                record.mux.Unlock()
            }
        }
    }()

    return record
}

func (record *MessageRecord) RecordMessage(
    kademliaID KademliaID,
    success func (Message, Contact),
    dead func (),
) {
    record.mux.Lock()
    record.record[kademliaID] = &success
    record.mux.Unlock()

    record.queue <- messageRecordQueueElem{
        Deadline: time.Now().Add(record.ttl),
        ID: kademliaID,
        dead: dead,
    }
}

func (record *MessageRecord) ActOnMessage(msg Message, contact Contact) {
    record.mux.Lock()

    ID := NewKademliaID(msg.RequestID)
    success, ok := record.record[*ID]
    if ok {
        delete(record.record, *ID)
        record.mux.Unlock()
        (*success)(msg, contact)
    } else {
        record.mux.Unlock()
    }
}

