package tradingviewsocket

import (
	"TradingBot/src/utils"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

// TradingviewSocket ...
type TradingviewSocket struct {
	OnReceiveMarketDataCallback func(data *MarketData)
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
	dialer := &websocket.Dialer{}

	headers := http.Header{}
	headers.Set("Host", "data.tradingview.com")
	headers.Set("Origin", "https://www.tradingview.com")

	var err error

	s.conn, _, err = dialer.Dial("wss://data.tradingview.com/socket.io/websocket", headers)
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

func (s *TradingviewSocket) checkFirstReceivedMessage() (err error) {
	var msg []byte

	_, msg, err = s.conn.ReadMessage()
	if err != nil {
		s.onError(err)
		return
	}

	payload := getPayload(msg)
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
			}
			continue
		}

		s.parseMessage(msg)
	}

	s.onError(readMsgError)
}

func isKeepAliveMsg(msg []byte) bool {
	return string(msg[getPayloadStartingIndex(msg)]) == "~"
}

func getPayload(msg []byte) []byte {
	return msg[getPayloadStartingIndex(msg):]
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
	s.conn.Close()
	s.OnErrorCallback(err)
}

func prependHeader(payload []byte) []byte {
	lengthAsString := strconv.Itoa(len(payload))
	return []byte("~m~" + lengthAsString + "~m~" + string(payload))
}

func (s *TradingviewSocket) parseMessage(msg []byte) {
	/**
		I can receive 2 or more messages in the same packet, like so

		"~m~97~m~{\"m\":\"qsd\",\"p\":[\"qs_scpSwsLPuGzA\",{\"n\":\"FX:EURUSD\",\"s\":\"ok\",\"v\":{\"volume\":242063,\"lp\":1.17821}}]}~m~96~m~{\"m\":\"qsd\",\"p\":[\"qs_scpSwsLPuGzA\",{\"n\":\"FX:EURUSD\",\"s\":\"ok\",\"v\":{\"bid\":1.17821,\"ask\":1.17821}}]}~m~59~m~{\"m\":\"quote_completed\",\"p\":[\"qs_scpSwsLPuGzA\",\"FX:EURUSD\"]}"

		So handle that carefully.
	**/
	fmt.Printf("%#v\n\n", string(msg))
}
