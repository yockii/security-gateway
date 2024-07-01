package proxy

import (
	"bytes"
	"net/http"
	"security-gateway/pkg/server"
	"strings"
	"sync"
)

const (
	HEADER_NO_MASKING = "No-Masking"
)

type MaskingResponseWriter struct {
	http.ResponseWriter
	mutex            sync.Mutex
	maskLevel        int
	maskingFields    map[string]*server.DesensitizeField
	inQuotes         bool          // 当前是否在引号中
	readingKey       bool          // 当前是否在读取key
	readyToReadValue bool          // 当前是否准备读取value
	isEscaped        bool          // 当前是否在转义
	readingValue     bool          // 当前是否在读取value
	currentKey       string        // 当前key
	valueBuffer      *bytes.Buffer // 缓冲当前读取的值
	inArray          bool          // 当前是否在数组中

	cachedNonValueBuffer *bytes.Buffer // 缓冲非值的数据，用于在读取值后，未遇到逗号或右括号时，将缓冲的数据写入ResponseWriter

	needMasking bool // 是否需要脱敏

	cachedBody *bytes.Buffer
}

func NewMaskingResponseWriterWithFieldMap(w http.ResponseWriter, maskingFields map[string]*server.DesensitizeField, maskLevel int) *MaskingResponseWriter {
	return &MaskingResponseWriter{
		ResponseWriter:       w,
		maskLevel:            maskLevel,
		maskingFields:        maskingFields,
		valueBuffer:          bytes.NewBuffer(nil),
		cachedNonValueBuffer: bytes.NewBuffer(nil),

		cachedBody: bytes.NewBuffer(nil),
	}

}

func NewMaskingResponseWriter(w http.ResponseWriter, maskingFields []*server.DesensitizeField, maskLevel int) *MaskingResponseWriter {
	fm := make(map[string]*server.DesensitizeField)
	for _, f := range maskingFields {
		fm[f.Name] = f
	}
	return NewMaskingResponseWriterWithFieldMap(w, fm, maskLevel)
}

func (m *MaskingResponseWriter) writeToResponse(b []byte) (int, error) {
	_, _ = m.cachedBody.Write(b)
	return m.ResponseWriter.Write(b)
}

func (m *MaskingResponseWriter) Write(b []byte) (int, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 如果不需要脱敏，则直接写入ResponseWriter
	// 从ResponseWriter中获取header
	if strings.ToLower(m.ResponseWriter.Header().Get(HEADER_NO_MASKING)) == "true" {
		return m.writeToResponse(b)
	}

	// 获取contentType
	contentType := m.Header().Get("Content-Type")

	// 逐字节处理脱敏
	if len(m.maskingFields) > 0 && strings.Contains(contentType, "application/json") {
		m.processChunk(b) // 直接处理数据并写入ResponseWriter
	} else {
		// 直接写入ResponseWriter
		_, _ = m.writeToResponse(b)
	}
	return len(b), nil
}

func (m *MaskingResponseWriter) processChunk(b []byte) {
	for _, c := range b {
		m.processByte(c)
	}
}

func (m *MaskingResponseWriter) processByte(c byte) {
	switch c {
	case '\\':
		m.isEscaped = true
		if m.inQuotes {
			m.valueBuffer.WriteByte('\\')
		}
	case '"':
		if m.isEscaped {
			m.isEscaped = false
			if m.readingValue || m.readingKey {
				// 同时写入转义字符和引号
				m.valueBuffer.WriteByte(c)
			}
		} else {
			m.inQuotes = !m.inQuotes
			if m.inQuotes {
				// 在引号中
				if !m.readingKey && !m.readingValue {
					// 进入引号，开始读取key
					m.readingKey = true
				}
			} else {
				// 离开引号
				if m.readingKey {
					// 结束读取key
					m.readingKey = false
					m.readyToReadValue = true
					m.currentKey = m.valueBuffer.String()
					m.valueBuffer.Reset()
					// 判断是否需要脱敏
					if _, ok := m.maskingFields[m.currentKey]; ok {
						m.needMasking = true
					} else {
						m.needMasking = false
					}
					// 写入ResponseWriter，key不脱敏，但要加上引号
					_, _ = m.writeToResponse([]byte("\"" + m.currentKey + "\""))
				} else if m.readingValue {
					if !m.inArray {
						// 结束读取value
						m.readingValue = false
						m.readyToReadValue = false
					}

					value := m.valueBuffer.String()
					if m.needMasking {
						value = "\"" + m.maskingFields[m.currentKey].Mask(value, m.maskLevel) + "\""
					} else {
						value = "\"" + value + "\""
					}
					if !m.inArray {
						m.needMasking = false
					}
					m.valueBuffer.Reset()
					_, _ = m.writeToResponse([]byte(value))
				} else {
					// 写入ResponseWriter
					_, _ = m.writeToResponse([]byte{c})
				}
			}
		}
	case ':':
		if m.readyToReadValue && !m.inQuotes {
			// 冒号后面是value
			m.readyToReadValue = false
			m.readingValue = true
			// 写入ResponseWriter
			_, _ = m.writeToResponse([]byte{c})
		} else if m.readingValue {
			m.valueBuffer.WriteByte(c)
		} else {
			// 写入ResponseWriter
			_, _ = m.writeToResponse([]byte{c})
		}
	case ',', '}':
		if m.inQuotes {
			m.valueBuffer.WriteByte(c)
		} else if !m.inQuotes && m.readingValue {
			if m.inArray {
				// 不在引号中，正在读取值，但是在数组中，则说明可能是布尔或数字
				// 结束value
				value := m.valueBuffer.String()
				if m.needMasking {
					value = "\"" + m.maskingFields[m.currentKey].Mask(value, m.maskLevel) + "\""
				}
				m.valueBuffer.Reset()
				_, _ = m.writeToResponse([]byte(value))
				// 写入缓冲的其他字符
				_, _ = m.writeToResponse(m.cachedNonValueBuffer.Bytes())
				m.cachedNonValueBuffer.Reset()
			} else {
				// 结束value
				m.readingValue = false
				value := m.valueBuffer.String()
				if m.needMasking {
					value = "\"" + m.maskingFields[m.currentKey].Mask(value, m.maskLevel) + "\""
				}
				m.needMasking = false
				m.valueBuffer.Reset()
				_, _ = m.writeToResponse([]byte(value))
				// 写入缓冲的其他字符
				_, _ = m.writeToResponse(m.cachedNonValueBuffer.Bytes())
				m.cachedNonValueBuffer.Reset()
			}
			// 写入c
			_, _ = m.writeToResponse([]byte{c})
		} else {
			// 写入ResponseWriter
			_, _ = m.writeToResponse([]byte{c})
		}
	case '{':
		if m.readingValue {
			if !m.inQuotes && m.valueBuffer.Len() == 0 {
				// 正在读取值，但发现了新的对象，说明值是对象，不需要脱敏，重置读取状态
				m.readingValue = false
				m.readyToReadValue = false
				m.inArray = false
				// 写入ResponseWriter
				_, _ = m.writeToResponse([]byte{c})
			} else if m.inQuotes {
				m.valueBuffer.WriteByte(c)
			} else {
				// 写入ResponseWriter
				_, _ = m.writeToResponse([]byte{c})
			}
		} else {
			// 写入ResponseWriter
			_, _ = m.writeToResponse([]byte{c})
		}
	case '[':
		if m.readingValue && m.inQuotes {
			m.valueBuffer.WriteByte(c)
		} else if m.readingValue && m.valueBuffer.Len() == 0 {
			// 正在读取值，但发现了新的数组，说明值是数组，标记在数组中，但是要根据数组内容来判定是否脱敏
			m.inArray = true
			m.readingValue = true
			// 写入ResponseWriter
			_, _ = m.writeToResponse([]byte{c})
		} else {
			// 写入ResponseWriter
			_, _ = m.writeToResponse([]byte{c})
		}
	case ']':
		if m.readingValue && m.inQuotes {
			m.valueBuffer.WriteByte(c)
		} else if m.inArray {
			if m.readingValue && m.valueBuffer.Len() > 0 {
				value := m.valueBuffer.String()
				if m.needMasking {
					value = "\"" + m.maskingFields[m.currentKey].Mask(value, m.maskLevel) + "\""
				}
				m.valueBuffer.Reset()
				_, _ = m.writeToResponse([]byte(value))
				// 写入缓冲的其他字符
				_, _ = m.writeToResponse(m.cachedNonValueBuffer.Bytes())
				m.cachedNonValueBuffer.Reset()
			}
			// 结束数组
			m.inArray = false
			m.readingValue = false
			// 写入ResponseWriter
			_, _ = m.writeToResponse([]byte{c})
		} else {
			// 写入ResponseWriter
			_, _ = m.writeToResponse([]byte{c})
		}
	case ' ', '\n', '\t', '\r':
		// 空白字符如果是在读取值前不处理
		if m.readingValue && m.valueBuffer.Len() > 0 && m.inQuotes {
			m.valueBuffer.WriteByte(c)
		} else if m.readingValue && m.valueBuffer.Len() > 0 {
			m.cachedNonValueBuffer.WriteByte(c)
		} else {
			// 写入ResponseWriter
			_, _ = m.writeToResponse([]byte{c})
		}
	default:
		m.isEscaped = false
		if m.inQuotes || m.readingValue {
			m.valueBuffer.WriteByte(c)
		} else {
			// 写入ResponseWriter
			_, _ = m.writeToResponse([]byte{c})
		}
	}
}
