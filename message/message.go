package message

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/crc64"
	"strconv"
	"strings"

	"github.com/cubevlmu/CZeroBot/utils/helper"
)

// Message impl the array form of message
// https://github.com/botuniverse/onebot-11/tree/master/message/array.md#%E6%95%B0%E7%BB%84%E6%A0%BC%E5%BC%8F
type Message []Segment

// Segment impl the single message
// Segment 消息数组
// https://github.com/botuniverse/onebot-11/tree/master/message/array.md#%E6%95%B0%E7%BB%84%E6%A0%BC%E5%BC%8F
type Segment struct {
	Type string            `json:"type"`
	Data map[string]string `json:"data"`
}

// CQCoder 用于 log 打印 CQ 码
type CQCoder interface {
	CQCode() string
}

// EscapeCQText escapes special characters in a non-media plain message.\
//
// CQ码字符转换
func EscapeCQText(str string) string {
	var buf [512]byte
	dst := buf[:0]
	for i := 0; i < len(str); i++ {
		switch str[i] {
		case '&':
			dst = append(dst, '&', 'a', 'm', 'p', ';')
		case '[':
			dst = append(dst, '&', '#', '9', '1', ';')
		case ']':
			dst = append(dst, '&', '#', '9', '3', ';')
		default:
			dst = append(dst, str[i])
		}
	}
	return string(dst)
}

// UnescapeCQText unescapes special characters in a non-media plain message.
//
// CQ码反解析
func UnescapeCQText(str string) string {
	str = strings.ReplaceAll(str, "&#93;", "]")
	str = strings.ReplaceAll(str, "&#91;", "[")
	str = strings.ReplaceAll(str, "&amp;", "&")
	return str
}

// EscapeCQCodeText escapes special characters in a cqcode value.
//
// https://github.com/botuniverse/onebot-11/tree/master/message/string.md#%E8%BD%AC%E4%B9%89
//
// cq码字符转换
func EscapeCQCodeText(str string) string {
	str = strings.ReplaceAll(str, "&", "&amp;")
	str = strings.ReplaceAll(str, "[", "&#91;")
	str = strings.ReplaceAll(str, "]", "&#93;")
	str = strings.ReplaceAll(str, ",", "&#44;")
	return str
}

// UnescapeCQCodeText unescapes special characters in a cqcode value.
// https://github.com/botuniverse/onebot-11/tree/master/message/string.md#%E8%BD%AC%E4%B9%89
//
// cq码反解析
func UnescapeCQCodeText(str string) string {
	str = strings.ReplaceAll(str, "&#44;", ",")
	str = strings.ReplaceAll(str, "&#93;", "]")
	str = strings.ReplaceAll(str, "&#91;", "[")
	str = strings.ReplaceAll(str, "&amp;", "&")
	return str
}

// CQCode 将数组消息转换为CQ码
// 与 String 不同之处在于，对于
// base64 的图片消息会将其哈希
// 方便 log 打印，不可用作发送
func (m Segment) CQCode() string {
	sb := strings.Builder{}
	sb.WriteString("[CQ:")
	sb.WriteString(m.Type)
	for k, v := range m.Data { // 消息参数
		// sb.WriteString("," + k + "=" + escape(v))
		sb.WriteByte(',')
		sb.WriteString(k)
		sb.WriteByte('=')
		switch m.Type {
		case "node":
			sb.WriteString(v)
		case "image":
			if strings.HasPrefix(v, "base64://") {
				v = v[9:]
				b, err := base64.StdEncoding.DecodeString(v)
				if err != nil {
					sb.WriteString(err.Error())
				} else {
					m := md5.Sum(b)
					_, _ = hex.NewEncoder(&sb).Write(m[:])
				}
				sb.WriteString(".image")
				break
			}
			fallthrough
		default:
			sb.WriteString(EscapeCQCodeText(v))
		}
	}
	sb.WriteByte(']')
	return sb.String()
}

// String impls the interface fmt.Stringer
func (m Segment) String() string {
	sb := strings.Builder{}
	sb.WriteString("[CQ:")
	sb.WriteString(m.Type)
	for k, v := range m.Data { // 消息参数
		// sb.WriteString("," + k + "=" + escape(v))
		sb.WriteByte(',')
		sb.WriteString(k)
		sb.WriteByte('=')
		if m.Type == "node" {
			sb.WriteString(v)
		} else {
			sb.WriteString(EscapeCQCodeText(v))
		}
	}
	sb.WriteByte(']')
	return sb.String()
}

// CQCode 将数组消息转换为CQ码
// 与 String 不同之处在于，对于
// base64 的图片消息会将其哈希
// 方便 log 打印，不可用作发送
func (m Message) CQCode() string {
	sb := strings.Builder{}
	for _, media := range m {
		if media.Type != "text" {
			sb.WriteString(media.CQCode())
		} else {
			sb.WriteString(EscapeCQText(media.Data["text"]))
		}
	}
	return sb.String()
}

// String impls the interface fmt.Stringer
func (m Message) String() string {
	sb := strings.Builder{}
	for _, media := range m {
		if media.Type != "text" {
			sb.WriteString(media.String())
		} else {
			sb.WriteString(EscapeCQText(media.Data["text"]))
		}
	}
	return sb.String()
}

// Text 纯文本
// https://github.com/botuniverse/onebot-11/tree/master/message/segment.md#%E7%BA%AF%E6%96%87%E6%9C%AC
func Text(text ...interface{}) Segment {
	return Segment{
		Type: "text",
		Data: map[string]string{
			"text": fmt.Sprint(text...),
		},
	}
}

// Face QQ表情
// https://github.com/botuniverse/onebot-11/tree/master/message/segment.md#qq-%E8%A1%A8%E6%83%85
func Face(id int) Segment {
	return Segment{
		Type: "face",
		Data: map[string]string{
			"id": strconv.Itoa(id),
		},
	}
}

// File 文件
// https://llonebot.github.io/zh-CN/develop/extends_api
func File(file, name string) Segment {
	return Segment{
		Type: "file",
		Data: map[string]string{
			"file": file,
			"name": name,
		},
	}
}

// Image 普通图片
//
// https://github.com/botuniverse/onebot-11/tree/master/message/segment.md#%E5%9B%BE%E7%89%87
//
// https://llonebot.github.io/zh-CN/develop/extends_api
//
// summary: LLOneBot的扩展字段：图片预览文字
func Image(file string, summary ...interface{}) Segment {
	m := Segment{
		Type: "image",
		Data: map[string]string{
			"file": file,
		},
	}
	if len(summary) > 0 {
		m.Data["summary"] = fmt.Sprint(summary...)
	}
	return m
}

// ImageBytes 普通图片
// https://github.com/botuniverse/onebot-11/tree/master/message/segment.md#%E5%9B%BE%E7%89%87
func ImageBytes(data []byte) Segment {
	return Segment{
		Type: "image",
		Data: map[string]string{
			"file": "base64://" + base64.StdEncoding.EncodeToString(data),
		},
	}
}

// Record 语音
// https://github.com/botuniverse/onebot-11/tree/master/message/segment.md#%E8%AF%AD%E9%9F%B3
func Record(file string) Segment {
	return Segment{
		Type: "record",
		Data: map[string]string{
			"file": file,
		},
	}
}

// Video 短视频
// https://github.com/botuniverse/onebot-11/blob/master/message/segment.md#%E7%9F%AD%E8%A7%86%E9%A2%91
func Video(file string) Segment {
	return Segment{
		Type: "video",
		Data: map[string]string{
			"file": file,
		},
	}
}

// At @某人
// https://github.com/botuniverse/onebot-11/tree/master/message/segment.md#%E6%9F%90%E4%BA%BA
func At(qq int64) Segment {
	if qq == 0 {
		return AtAll()
	}
	return Segment{
		Type: "at",
		Data: map[string]string{
			"qq": strconv.FormatInt(qq, 10),
		},
	}
}

// AtAll @全体成员
// https://github.com/botuniverse/onebot-11/tree/master/message/segment.md#%E6%9F%90%E4%BA%BA
func AtAll() Segment {
	return Segment{
		Type: "at",
		Data: map[string]string{
			"qq": "all",
		},
	}
}

// Music 音乐分享
// https://github.com/botuniverse/onebot-11/tree/master/message/segment.md#%E9%9F%B3%E4%B9%90%E5%88%86%E4%BA%AB-
func Music(mType string, id int64) Segment {
	return Segment{
		Type: "music",
		Data: map[string]string{
			"type": mType,
			"id":   strconv.FormatInt(id, 10),
		},
	}
}

// CustomMusic 音乐自定义分享
// https://github.com/botuniverse/onebot-11/tree/master/message/segment.md#%E9%9F%B3%E4%B9%90%E8%87%AA%E5%AE%9A%E4%B9%89%E5%88%86%E4%BA%AB-
func CustomMusic(url, audio, title string) Segment {
	return Segment{
		Type: "music",
		Data: map[string]string{
			"type":  "custom",
			"url":   url,
			"audio": audio,
			"title": title,
		},
	}
}

// ID 对于 qq 消息, i 与 s 相同
// 对于 guild 消息, i 为 s 的 ISO crc64
type ID struct {
	i int64
	s string
}

func NewMessageIDFromString(raw string) (m ID) {
	var err error
	m.i, err = strconv.ParseInt(raw, 10, 64)
	if err != nil {
		c := crc64.New(crc64.MakeTable(crc64.ISO))
		c.Write(helper.StringToBytes(raw))
		m.i = int64(c.Sum64())
	}
	m.s = raw
	return
}

func NewMessageIDFromInteger(raw int64) (m ID) {
	m.s = strconv.FormatInt(raw, 10)
	m.i = raw
	return
}

func (m ID) MarshalJSON() ([]byte, error) {
	sb := bytes.NewBuffer(make([]byte, 0, len(m.s)+2))
	_, err := strconv.ParseInt(m.s, 10, 64)
	if err != nil {
		sb.WriteByte('"')
		json.HTMLEscape(sb, []byte(m.s))
		sb.WriteByte('"')
	} else {
		sb.WriteString(m.s)
	}
	return sb.Bytes(), nil
}

func (m ID) String() string {
	return m.s
}

func (m ID) ID() int64 {
	return m.i
}

// Reply 回复
// https://github.com/botuniverse/onebot-11/tree/master/message/segment.md#%E5%9B%9E%E5%A4%8D
func Reply(id interface{}) Segment {
	s := ""
	switch i := id.(type) {
	case int64:
		s = strconv.FormatInt(i, 10)
	case int:
		s = strconv.Itoa(i)
	case string:
		s = i
	case float64:
		s = strconv.Itoa(int(i)) // json 序列化 interface{} 默认为 float64
	case fmt.Stringer:
		s = i.String()
	}
	return Segment{
		Type: "reply",
		Data: map[string]string{
			"id": s,
		},
	}
}

// Forward 合并转发
// https://github.com/botuniverse/onebot-11/tree/master/message/segment.md#%E5%90%88%E5%B9%B6%E8%BD%AC%E5%8F%91-
func Forward(id string) Segment {
	return Segment{
		Type: "forward",
		Data: map[string]string{
			"id": id,
		},
	}
}

// Node 合并转发节点
// https://github.com/botuniverse/onebot-11/tree/master/message/segment.md#%E5%90%88%E5%B9%B6%E8%BD%AC%E5%8F%91%E8%8A%82%E7%82%B9-
func Node(id int64) Segment {
	return Segment{
		Type: "node",
		Data: map[string]string{
			"id": strconv.FormatInt(id, 10),
		},
	}
}

// CustomNode 自定义合并转发节点
// https://github.com/botuniverse/onebot-11/tree/master/message/segment.md#%E5%90%88%E5%B9%B6%E8%BD%AC%E5%8F%91%E8%87%AA%E5%AE%9A%E4%B9%89%E8%8A%82%E7%82%B9
func CustomNode(nickname string, userID int64, content interface{}) Segment {
	var str string
	switch c := content.(type) {
	case string:
		str = c
	case Message:
		str = c.String()
	case []Segment:
		str = (Message)(c).String()
	default:
		b, _ := json.Marshal(content)
		str = helper.BytesToString(b)
	}
	return Segment{
		Type: "node",
		Data: map[string]string{
			"uin":     strconv.FormatInt(userID, 10),
			"name":    nickname,
			"content": str,
		},
	}
}

// XML 消息
// https://github.com/botuniverse/onebot-11/tree/master/message/segment.md#xml-%E6%B6%88%E6%81%AF
func XML(data string) Segment {
	return Segment{
		Type: "xml",
		Data: map[string]string{
			"data": data,
		},
	}
}

// JSON 消息
// https://github.com/botuniverse/onebot-11/tree/master/message/segment.md#xml-%E6%B6%88%E6%81%AF
func JSON(data string) Segment {
	return Segment{
		Type: "json",
		Data: map[string]string{
			"data": data,
		},
	}
}

// Expand CQCode

// Gift 群礼物
// https://github.com/Mrs4s/go-cqhttp/blob/master/docs/cqhttp.md#%E7%A4%BC%E7%89%A9
//
// Deprecated: 群礼物改版
func Gift(userID string, giftID string) Segment {
	return Segment{
		Type: "gift",
		Data: map[string]string{
			"qq": userID,
			"id": giftID,
		},
	}
}

// Poke 戳一戳
// https://github.com/Mrs4s/go-cqhttp/blob/master/docs/cqhttp.md#%E6%88%B3%E4%B8%80%E6%88%B3
func Poke(userID int64) Segment {
	return Segment{
		Type: "poke",
		Data: map[string]string{
			"qq": strconv.FormatInt(userID, 10),
		},
	}
}

// TTS 文本转语音
// https://github.com/Mrs4s/go-cqhttp/blob/master/docs/cqhttp.md#%E6%96%87%E6%9C%AC%E8%BD%AC%E8%AF%AD%E9%9F%B3
func TTS(text string) Segment {
	return Segment{
		Type: "tts",
		Data: map[string]string{
			"text": text,
		},
	}
}

// Add 为 MessageSegment 的 Data 增加一个字段
func (m Segment) Add(key string, val interface{}) Segment {
	switch val := val.(type) {
	case string:
		m.Data[key] = val
	case bool:
		m.Data[key] = strconv.FormatBool(val)
	case int:
		m.Data[key] = strconv.FormatInt(int64(val), 10)
	case fmt.Stringer:
		m.Data[key] = val.String()
	default:
		m.Data[key] = fmt.Sprint(val)
	}
	return m
}

// Chain 将两个 Data 合并
func (m Segment) Chain(data map[string]string) Segment {
	for k, v := range data {
		m.Data[k] = v
	}
	return m
}

// ReplyWithMessage returns a reply message
func ReplyWithMessage(messageID interface{}, m ...Segment) Message {
	return append(Message{Reply(messageID)}, m...)
}
