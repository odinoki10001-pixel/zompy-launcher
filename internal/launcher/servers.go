package servers

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type Server struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Emoji    string   `json:"emoji"`
	IP       string   `json:"ip"`
	Port     int      `json:"port"`
	Mode     string   `json:"mode"` // survival, creative, minigames, rpg, anarchy
	Version  string   `json:"version"`
	Auth     string   `json:"auth"` // offline, online, both
	Featured bool     `json:"featured"`
	Tags     []string `json:"tags"`
	Desc     string   `json:"desc"`
}

type Ping struct {
	Online    int    `json:"online"`
	Max       int    `json:"max"`
	LatencyMS int    `json:"latency"`
	MOTD      string `json:"motd"`
	Version   string `json:"version"`
	OK        bool   `json:"ok"`
	Error     string `json:"error,omitempty"`
}

func Default() []Server {
	return []Server{
		{ID: 1, Name: "Hypixel", Emoji: "🏆", IP: "mc.hypixel.net", Port: 25565, Mode: "minigames", Version: "1.21", Auth: "online", Featured: true, Tags: []string{"BedWars", "SkyWars", "PvP"}, Desc: "Крупнейший сервер с десятками мини-игр."},
		{ID: 2, Name: "MineBlaze", Emoji: "🔥", IP: "play.mineblaze.net", Port: 25565, Mode: "survival", Version: "1.21", Auth: "both", Featured: true, Tags: []string{"Survival", "Кланы"}, Desc: "Уютное выживание с кланами и экономикой."},
		{ID: 3, Name: "CraftLandia", Emoji: "🌋", IP: "craftlandia.ru", Port: 25565, Mode: "survival", Version: "1.16", Auth: "offline", Tags: []string{"Анархия", "PvP"}, Desc: "Классическая анархия без правил."},
		{ID: 4, Name: "RPGLife", Emoji: "⚔️", IP: "play.rpglife.su", Port: 25565, Mode: "rpg", Version: "1.20", Auth: "offline", Tags: []string{"RPG", "Квесты"}, Desc: "Настоящая MMORPG в Minecraft."},
		{ID: 5, Name: "CreativePlots", Emoji: "🎨", IP: "creative.zompy.net", Port: 25565, Mode: "creative", Version: "1.21", Auth: "both", Tags: []string{"Креатив", "Плоты"}, Desc: "Творческий режим с приватами."},
		{ID: 6, Name: "PixelRealms", Emoji: "🎮", IP: "play.pixelrealms.net", Port: 25565, Mode: "rpg", Version: "1.20", Auth: "both", Featured: true, Tags: []string{"Pixelmon"}, Desc: "Ловите покемонов и сражайтесь."},
	}
}

// Minecraft Server List Ping (SLP) — современный протокол (1.7+)
func PingServer(host string, port int, timeout time.Duration) Ping {
	if port == 0 {
		port = 25565
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return Ping{OK: false, Error: err.Error()}
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))

	// Handshake packet
	var buf bytes.Buffer
	writeVarInt(&buf, 0x00) // packet id
	writeVarInt(&buf, 765)  // protocol version (1.20.4)
	writeString(&buf, host)
	_ = binary.Write(&buf, binary.BigEndian, uint16(port))
	writeVarInt(&buf, 0x01) // next state: status

	// wrap in length prefix
	var out bytes.Buffer
	writeVarInt(&out, int32(buf.Len()))
	out.Write(buf.Bytes())

	if _, err := conn.Write(out.Bytes()); err != nil {
		return Ping{OK: false, Error: err.Error()}
	}
	// status request
	if _, err := conn.Write([]byte{0x01, 0x00}); err != nil {
		return Ping{OK: false, Error: err.Error()}
	}

	// read response
	_, err = readVarInt(conn)
	if err != nil {
		return Ping{OK: false, Error: err.Error()}
	}
	pktID, _ := readVarInt(conn)
	if pktID != 0x00 {
		return Ping{OK: false, Error: "unexpected packet"}
	}
	jsonLen, _ := readVarInt(conn)
	raw := make([]byte, jsonLen)
	if _, err := readFull(conn, raw); err != nil {
		return Ping{OK: false, Error: err.Error()}
	}

	var data struct {
		Version struct {
			Name string `json:"name"`
		} `json:"version"`
		Players struct {
			Online int `json:"online"`
			Max    int `json:"max"`
		} `json:"players"`
		Description any `json:"description"`
	}
	_ = json.Unmarshal(raw, &data)

	return Ping{
		OK:        true,
		Online:    data.Players.Online,
		Max:       data.Players.Max,
		Version:   data.Version.Name,
		LatencyMS: int(time.Since(start).Milliseconds()),
		MOTD:      motdString(data.Description),
	}
}

func motdString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case map[string]any:
		if t, ok := x["text"].(string); ok {
			return t
		}
	}
	return ""
}

// --- протокол helpers ---

func writeVarInt(w *bytes.Buffer, v int32) {
	u := uint32(v)
	for {
		if u&^0x7F == 0 {
			w.WriteByte(byte(u))
			return
		}
		w.WriteByte(byte(u&0x7F | 0x80))
		u >>= 7
	}
}

func writeString(w *bytes.Buffer, s string) {
	writeVarInt(w, int32(len(s)))
	w.WriteString(s)
}

func readVarInt(r net.Conn) (int32, error) {
	var numRead int
	var result int32
	for {
		b := make([]byte, 1)
		if _, err := r.Read(b); err != nil {
			return 0, err
		}
		result |= int32(b[0]&0x7F) << (7 * numRead)
		numRead++
		if numRead > 5 {
			return 0, fmt.Errorf("varint too big")
		}
		if b[0]&0x80 == 0 {
			break
		}
	}
	return result, nil
}

func readFull(r net.Conn, buf []byte) (int, error) {
	total := 0
	for total < len(buf) {
		n, err := r.Read(buf[total:])
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}
