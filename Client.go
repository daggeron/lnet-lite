package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
)

const (
	maxPacketSize = 32768
)

type Pong struct{}

type lnetRequest struct {
	XMLName      xml.Name `xml:"request"`
	From         string   `xml:"from,attr,omitempty"`
	Type         string   `xml:"type,attr"`
	To           string   `xml:"to,attr,omitempty"`
}

type queryRequest struct {
	XMLName      xml.Name `xml:"request"`
	From         string   `xml:"from,attr"`
	Type         string   `xml:"type,attr"`
}


type XMLElement struct {
	XMLName  xml.Name
	InnerXML string `xml:",innerxml"`
}

type pingRequest struct {
	XMLName      xml.Name  `xml:"ping"`
}

type pongResponse struct {
	XMLName      xml.Name  `xml:"pong"`
}

type Login struct {
	XMLName  xml.Name `xml:"login"`
	Client   string   `xml:"client,attr,omitempty"`
	Game     string   `xml:"game,attr,omitempty"`
	Lich     string   `xml:"lich,attr,omitempty"`
	Name     string   `xml:"name,attr,omitempty"`
	Password string   `xml:"password,attr,omitempty"`
}

type Data struct {
	XMLName  xml.Name `xml:"data"`
	Type         string   `xml:"type,attr,omitempty"`
	From         string   `xml:"from,attr,omitempty"`
	To           string   `xml:"to,attr,omitempty"`
	Text         string   `xml:",innerxml"`
}


type Message struct {
	XMLName      xml.Name `xml:"message"`
	Type         string   `xml:"type,attr,omitempty"`
	From         string   `xml:"from,attr,omitempty"`
	To           string   `xml:"to,attr,omitempty"`
	Subscription string   `xml:",attr,omitempty"`
	Channel      string   `xml:"channel,attr,omitempty"`
	Text         string   `xml:",innerxml"`
}

type Client struct {
	conn       net.Conn
	nickname   string
	readWriter io.ReadWriter
	decoder    *xml.Decoder
	encoder    *xml.Encoder
}








func (c *Client) Recv() (stanza interface{}, err error) {
	for {
		_, val, err := c.next()
		if err != nil {
			return Pong{}, err
		}
		switch v := val.(type) {
		case *Data:
			v.From = c.nickname
			return v, nil
		case *lnetRequest:
			v.From = c.nickname
			return v, nil
		case *Message:
			v.From = c.nickname
			return v, nil
		case *pingRequest:
			fmt.Println("Ping!")
			c.Send(Pong{})
			return Pong{}, nil
		case *Login:

			c.nickname = v.Name
			var msg = &Message{
				To: c.nickname,
				From: "server",
				Text: "Hello",
				Type: "server",

			}
			c.Send(msg)
			return msg, nil
		}
	}
}
func (e *XMLElement) String() string {
	r := bytes.NewReader([]byte(e.InnerXML))
	d := xml.NewDecoder(r)
	var buf bytes.Buffer
	for {
		tok, err := d.Token()
		if err != nil {
			break
		}
		switch v := tok.(type) {
		case xml.StartElement:
			err = d.Skip()
		case xml.CharData:
			_, err = buf.Write(v)
		}
		if err != nil {
			break
		}
	}
	return buf.String()
}

func (c *Client) SendKeepAlive() (n int, err error) {
	return fmt.Fprintf(c.conn, " ")
}

// Scan XML token stream to find next StartElement.
func (c *Client) nextStart() (xml.StartElement, error) {
	for {
		t, err := c.decoder.Token()
		if err != nil || t == nil {
			return xml.StartElement{}, err
		}
		switch t := t.(type) {
		case xml.StartElement:
			return t, nil
		}
	}
}

// Scan XML token stream for next element and save into val.
// If val == nil, allocate new element based on proto map.
// Either way, return val.
func (c *Client) next() (xml.Name, interface{}, error) {
	// Read start element to find out what type we want.
	se, err := c.nextStart()
	if err != nil {
		return xml.Name{}, nil, err
	}

	// Put it in an interface and allocate one.
	var nv interface{}
	switch se.Name.Local {
	case "message":
		nv = &Message{}
	case "login":
		nv = &Login{}
	case "pong":
		nv = &pongResponse{}
	case "request":
		nv = &lnetRequest{}
	case "data":
		nv = &Data{}
	default:
		return xml.Name{}, nil, errors.New("unexpected LNET message " +
			se.Name.Space + " <" + se.Name.Local + "/>")
	}

	// Unmarshal into that storage.
	if err = c.decoder.DecodeElement(nv, &se); err != nil {
		return xml.Name{}, nil, err
	}

	return se.Name, nv, err
}


func (c *Client) Send(msg interface{}) error {

	err := c.encoder.Encode(msg)

	return err
}


func (c *Client) Write(msg string) error {
	_, err := fmt.Fprint(c.readWriter, msg)

	return err
}

func (c *Client) init() {
	c.readWriter = NewStreamLogger(c.conn, os.Stdout)
	c.decoder = xml.NewDecoder(bufio.NewReaderSize(c.readWriter, maxPacketSize))
	c.encoder = xml.NewEncoder(bufio.NewWriterSize(c.readWriter, maxPacketSize))
}

func (c *Client) Close() error {
	return c.conn.Close()
}
