package tradingviewsocket

import (
	"TradingBot/src/utils"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

// SocketMessage ...
type SocketMessage struct {
	Message string      `json:"m"`
	Payload interface{} `json:"p"`
}

// TradingviewSocket ...
type TradingviewSocket struct {
	OnReceiveMarketDataCallback func(symbol string, data map[string]interface{})
	OnErrorCallback             func(error)

	conn      *websocket.Conn
	sessionID string
	symbols   []string
}

// AddSymbol ...
func (s *TradingviewSocket) AddSymbol(symbol string) (err error) {
	s.symbols = append(s.symbols, symbol)
	if s.conn != nil {
		err = s.addSymbolsToSocketSession()
	}
	return
}

// Init connects to the tradingview web socket
func (s *TradingviewSocket) Init() {
	var err error

	s.conn, _, err = (&websocket.Dialer{}).Dial("wss://data.tradingview.com/socket.io/websocket", getHeaders())
	if err != nil {
		s.onError(err)
		return
	}

	err = s.checkFirstReceivedMessage()
	if err != nil {
		s.onError(err)
		return
	}

	s.generateSessionID()

	err = s.sendFirstMessages()
	if err != nil {
		s.onError(err)
		return
	}

	go s.connectionLoop()
}

// Close ...
func (s *TradingviewSocket) Close() (err error) {
	return s.conn.Close()
}

func (s *TradingviewSocket) checkFirstReceivedMessage() (err error) {
	var msg []byte

	_, msg, err = s.conn.ReadMessage()
	if err != nil {
		s.onError(err)
		return
	}

	payload := msg[getPayloadStartingIndex(msg):]
	var p map[string]interface{}

	err = json.Unmarshal(payload, &p)
	if err != nil {
		return
	}

	if p["session_id"] == nil {
		err = errors.New("Cannot recognize the first received message after establishing the connection")
		return
	}

	return
}

func (s *TradingviewSocket) generateSessionID() {
	s.sessionID = "qs_" + utils.GetRandomString(12)
}

func (s *TradingviewSocket) sendFirstMessages() (err error) {
	messages := []*SocketMessage{
		&SocketMessage{
			Message: "set_auth_token",
			Payload: []string{"unauthorized_user_token"},
		},
		&SocketMessage{
			Message: "quote_create_session",
			Payload: []string{s.sessionID},
		},
		&SocketMessage{
			Message: "quote_set_fields",
			Payload: []string{s.sessionID, "lp", "volume", "bid", "ask"},
		},
	}

	for _, msg := range messages {
		err = s.sendPayload(msg)
		if err != nil {
			return
		}
	}

	err = s.addSymbolsToSocketSession()
	return
}

func (s *TradingviewSocket) addSymbolsToSocketSession() (err error) {
	for _, symbol := range s.symbols {
		flags := struct {
			Flags []string `json:"flags"`
		}{
			Flags: []string{"force_permission"},
		}
		msg := &SocketMessage{
			Message: "quote_add_symbols",
			Payload: []interface{}{s.sessionID, symbol, flags},
		}
		err = s.sendPayload(msg)
		if err != nil {
			return
		}
	}
	return
}

func (s *TradingviewSocket) sendPayload(p *SocketMessage) (err error) {
	payload, _ := json.Marshal(p)
	err = s.conn.WriteMessage(websocket.TextMessage, prependHeader(payload))
	if err != nil {
		return
	}
	return
}

func (s *TradingviewSocket) connectionLoop() {
	var readMsgError error

	for readMsgError == nil {
		var msgType int
		var msg []byte
		msgType, msg, readMsgError = s.conn.ReadMessage()

		if msgType != websocket.TextMessage {
			continue
		}

		if isKeepAliveMsg(msg) {
			err := s.conn.WriteMessage(msgType, msg)
			if err != nil {
				s.onError(err)
				return
			}
			continue
		}

		s.parsePacket(msg)
	}

	s.onError(readMsgError)
}

func isKeepAliveMsg(msg []byte) bool {
	return string(msg[getPayloadStartingIndex(msg)]) == "~"
}

func getPayloadStartingIndex(msg []byte) int {
	char := ""
	index := 3
	for char != "~" {
		char = string(msg[index])
		index++
	}
	index += 2
	return index
}

func (s *TradingviewSocket) onError(err error) {
	if s.conn != nil {
		s.conn.Close()
	}
	s.OnErrorCallback(err)
}

func prependHeader(payload []byte) []byte {
	lengthAsString := strconv.Itoa(len(payload))
	return []byte("~m~" + lengthAsString + "~m~" + string(payload))
}

func (s *TradingviewSocket) parsePacket(packet []byte) {
	index := 0
	for index < len(packet) {
		payloadLength, err := getPayloadLength(packet[index:])
		if err != nil {
			s.onError(err)
			return
		}

		headerLength := 6 + len(strconv.Itoa(payloadLength))
		payload := packet[index+headerLength : index+headerLength+payloadLength]
		index = index + headerLength + len(payload)

		s.parseJSON(payload)
	}
}

func getPayloadLength(msg []byte) (length int, err error) {
	char := ""
	index := 3
	lengthAsString := ""
	for char != "~" {
		char = string(msg[index])
		if char != "~" {
			lengthAsString += char
		}
		index++
	}
	length, err = strconv.Atoi(lengthAsString)
	return
}

func (s *TradingviewSocket) parseJSON(msg []byte) {
	var decodedJSON map[string]interface{}

	json.Unmarshal(msg, &decodedJSON)

	if decodedJSON["m"] != "qsd" {
		return
	}

	if decodedJSON["p"] == nil {
		s.onError(errors.New("Msg does not include 'p' -> " + string(msg)))
		return
	}

	p, isPOk := decodedJSON["p"].([]interface{})
	if !isPOk || len(p) != 2 {
		s.onError(errors.New("There is something wrong with the payload - can't be parsed -> " + string(msg)))
		return
	}

	messageThatMatters, isMessageThatMattersOk := p[1].(map[string]interface{})
	if !isMessageThatMattersOk || messageThatMatters["n"] == nil || messageThatMatters["s"] != "ok" || messageThatMatters["v"] == nil {
		s.onError(errors.New("There is something wrong with the payload - can't be parsed -> " + string(msg)))
		return
	}

	symbol, isSymbolOK := messageThatMatters["n"].(string)
	data, isDataOK := messageThatMatters["v"].(map[string]interface{})

	if !isSymbolOK || !isDataOK {
		s.onError(errors.New("Can't parse message -> " + string(msg)))
		return
	}

	s.OnReceiveMarketDataCallback(symbol, data)
}

func getHeaders() http.Header {
	headers := http.Header{}

	headers.Set("Accept-Encoding", "gzip, deflate, br")
	headers.Set("Accept-Language", "en-US,en;q=0.9,es;q=0.8")
	headers.Set("Cache-Control", "no-cache")
	headers.Set("Host", "data.tradingview.com")
	headers.Set("Origin", "https://www.tradingview.com")
	headers.Set("Pragma", "no-cache")
	headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.193 Safari/537.36")

	return headers
}
